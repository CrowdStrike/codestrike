import type { SCMClient, PullRequestFile, PRComment } from './types.js';

interface GitHubConfig {
  owner: string;
  repo: string;
  token: string;
  baseURL: string;
}

export class GitHubClient implements SCMClient {
  private config: GitHubConfig;

  constructor(cfg: GitHubConfig) {
    this.config = { ...cfg };
    if (!this.config.baseURL) this.config.baseURL = 'https://api.github.com';
  }

  private headers(accept?: string): Record<string, string> {
    const h: Record<string, string> = {
      Accept: accept ?? 'application/vnd.github.v3+json',
    };
    if (this.config.token) h['Authorization'] = `Bearer ${this.config.token}`;
    return h;
  }

  async pullRequestExists(number: number): Promise<boolean> {
    const url = `${this.config.baseURL}/repos/${this.config.owner}/${this.config.repo}/pulls/${number}`;
    const resp = await fetch(url, { headers: this.headers() });
    if (resp.status === 404) return false;
    if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);
    return true;
  }

  async getPullRequestDescription(number: number): Promise<{ title: string; body: string }> {
    const url = `${this.config.baseURL}/repos/${this.config.owner}/${this.config.repo}/pulls/${number}`;
    const resp = await fetch(url, { headers: this.headers() });
    if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);
    const pr = (await resp.json()) as { title: string; body: string };
    return { title: pr.title, body: pr.body ?? '' };
  }

  async getPullRequestDiff(number: number): Promise<string> {
    const url = `${this.config.baseURL}/repos/${this.config.owner}/${this.config.repo}/pulls/${number}`;
    const resp = await fetch(url, { headers: this.headers('application/vnd.github.v3.diff') });
    if (resp.status === 404) throw new Error(`pull request #${number} not found`);
    if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);
    return resp.text();
  }

  async getPullRequestFiles(number: number): Promise<PullRequestFile[]> {
    const url = `${this.config.baseURL}/repos/${this.config.owner}/${this.config.repo}/pulls/${number}/files`;
    const resp = await fetch(url, { headers: this.headers() });
    if (resp.status === 404) throw new Error(`pull request #${number} not found`);
    if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);

    const ghFiles = (await resp.json()) as { filename: string; status: string; patch: string }[];
    return ghFiles.map((f) => ({
      filename: f.filename,
      status: f.status,
      patch: f.patch ?? '',
      content: '',
    }));
  }

  async getFileContent(path: string, ref: string): Promise<string> {
    let url = `${this.config.baseURL}/repos/${this.config.owner}/${this.config.repo}/contents/${path}`;
    if (ref) url += `?ref=${ref}`;
    const resp = await fetch(url, { headers: this.headers('application/vnd.github.v3.raw') });
    if (resp.status === 404) throw new Error(`file "${path}" not found`);
    if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);
    return resp.text();
  }

  async publishComment(number: number, body: string): Promise<void> {
    const url = `${this.config.baseURL}/repos/${this.config.owner}/${this.config.repo}/issues/${number}/comments`;
    const resp = await fetch(url, {
      method: 'POST',
      headers: { ...this.headers(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ body }),
    });
    if (resp.status === 404) throw new Error(`pull request #${number} not found`);
    if (resp.status !== 201) {
      const err = await parseGitHubError(resp);
      throw new Error(err);
    }
  }

  async getPRComments(number: number): Promise<PRComment[]> {
    const url = `${this.config.baseURL}/repos/${this.config.owner}/${this.config.repo}/issues/${number}/comments`;
    return this.fetchComments(url);
  }

  async getPRReviewComments(number: number): Promise<PRComment[]> {
    const url = `${this.config.baseURL}/repos/${this.config.owner}/${this.config.repo}/pulls/${number}/comments`;
    return this.fetchComments(url);
  }

  private async fetchComments(baseUrl: string): Promise<PRComment[]> {
    const allComments: PRComment[] = [];
    let page = 1;

    while (true) {
      const url = `${baseUrl}?per_page=100&page=${page}`;
      const resp = await fetch(url, { headers: this.headers() });
      if (resp.status === 404) throw new Error('resource not found');
      if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);

      const ghComments = (await resp.json()) as {
        id: number;
        body: string;
        created_at: string;
        path?: string;
        line?: number;
        in_reply_to_id?: number;
        user: { login: string };
      }[];

      if (ghComments.length === 0) break;

      for (const gc of ghComments) {
        allComments.push({
          id: gc.id,
          author: gc.user.login,
          body: gc.body,
          createdAt: new Date(gc.created_at),
          path: gc.path ?? '',
          line: gc.line ?? 0,
          inReplyTo: gc.in_reply_to_id ?? 0,
        });
      }

      page++;
      if (ghComments.length < 100) break;
    }

    return allComments;
  }
}

async function parseGitHubError(resp: Response): Promise<string> {
  try {
    const body = (await resp.json()) as { message?: string; documentation_url?: string };
    if (body.message) {
      const suffix = body.documentation_url ? ` (${body.documentation_url})` : '';
      return `GitHub API returned ${resp.status} ${resp.statusText}: ${body.message}${suffix}`;
    }
  } catch {
    // ignore parse errors
  }
  return `GitHub API returned ${resp.status} ${resp.statusText}`;
}
