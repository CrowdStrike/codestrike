import type { SCMClient, PullRequestFile, PRComment } from './types.js';

interface BitbucketConfig {
  workspace: string;
  repoSlug: string;
  token: string;
  baseURL: string;
}

export class BitbucketClient implements SCMClient {
  private config: BitbucketConfig;

  constructor(cfg: BitbucketConfig) {
    this.config = { ...cfg };
    if (!this.config.baseURL) this.config.baseURL = 'https://api.bitbucket.org/2.0';
  }

  private headers(): Record<string, string> {
    const h: Record<string, string> = {};
    if (this.config.token) h['Authorization'] = `Bearer ${this.config.token}`;
    return h;
  }

  async pullRequestExists(number: number): Promise<boolean> {
    const url = `${this.config.baseURL}/repositories/${this.config.workspace}/${this.config.repoSlug}/pullrequests/${number}`;
    const resp = await fetch(url, { headers: this.headers() });
    if (resp.status === 404) return false;
    if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);
    return true;
  }

  async getPullRequestDescription(number: number): Promise<{ title: string; body: string }> {
    const url = `${this.config.baseURL}/repositories/${this.config.workspace}/${this.config.repoSlug}/pullrequests/${number}`;
    const resp = await fetch(url, { headers: this.headers() });
    if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);
    const pr = (await resp.json()) as { title: string; summary: { raw: string } };
    return { title: pr.title, body: pr.summary?.raw ?? '' };
  }

  async getPullRequestDiff(number: number): Promise<string> {
    const url = `${this.config.baseURL}/repositories/${this.config.workspace}/${this.config.repoSlug}/pullrequests/${number}/diff`;
    const resp = await fetch(url, { headers: this.headers() });
    if (resp.status === 404) throw new Error(`pull request #${number} not found`);
    if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);
    return resp.text();
  }

  async getPullRequestFiles(number: number): Promise<PullRequestFile[]> {
    const entries = await this.fetchDiffstat(number);
    const fullDiff = await this.getPullRequestDiff(number);
    const patches = splitDiff(fullDiff);

    return entries.map((e) => ({
      filename: e.path,
      status: e.status,
      patch: patches[e.path] ?? '',
      content: '',
    }));
  }

  async getFileContent(path: string, ref: string): Promise<string> {
    if (!ref) ref = 'HEAD';
    const url = `${this.config.baseURL}/repositories/${this.config.workspace}/${this.config.repoSlug}/src/${ref}/${path}`;
    const resp = await fetch(url, { headers: this.headers() });
    if (resp.status === 404) throw new Error(`file "${path}" not found`);
    if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);
    return resp.text();
  }

  async publishComment(number: number, body: string): Promise<void> {
    const url = `${this.config.baseURL}/repositories/${this.config.workspace}/${this.config.repoSlug}/pullrequests/${number}/comments`;
    const resp = await fetch(url, {
      method: 'POST',
      headers: { ...this.headers(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: { raw: body } }),
    });
    if (resp.status === 404) throw new Error(`pull request #${number} not found`);
    if (resp.status !== 201) {
      const err = await parseBitbucketError(resp);
      throw new Error(err);
    }
  }

  async getPRComments(number: number): Promise<PRComment[]> {
    const all = await this.fetchAllComments(number);
    return all.filter((c) => !c.path);
  }

  async getPRReviewComments(number: number): Promise<PRComment[]> {
    const all = await this.fetchAllComments(number);
    return all.filter((c) => !!c.path);
  }

  private async fetchAllComments(number: number): Promise<PRComment[]> {
    let url: string | null =
      `${this.config.baseURL}/repositories/${this.config.workspace}/${this.config.repoSlug}/pullrequests/${number}/comments`;
    const allComments: PRComment[] = [];

    while (url) {
      const resp = await fetch(url, { headers: this.headers() });
      if (resp.status === 404) throw new Error(`pull request #${number} not found`);
      if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);

      const page = (await resp.json()) as {
        values: {
          id: number;
          created_on: string;
          content: { raw: string };
          user: { display_name: string };
          inline?: { path: string; to: number };
          parent?: { id: number };
        }[];
        next?: string;
      };

      for (const bc of page.values) {
        allComments.push({
          id: bc.id,
          author: bc.user.display_name,
          body: bc.content.raw,
          createdAt: new Date(bc.created_on),
          path: bc.inline?.path ?? '',
          line: bc.inline?.to ?? 0,
          inReplyTo: bc.parent?.id ?? 0,
        });
      }

      url = page.next ?? null;
    }

    return allComments;
  }

  private async fetchDiffstat(number: number): Promise<{ status: string; path: string }[]> {
    let url: string | null =
      `${this.config.baseURL}/repositories/${this.config.workspace}/${this.config.repoSlug}/pullrequests/${number}/diffstat`;
    const entries: { status: string; path: string }[] = [];

    while (url) {
      const resp = await fetch(url, { headers: this.headers() });
      if (!resp.ok) throw new Error(`unexpected status ${resp.status}`);

      const page = (await resp.json()) as {
        values: { status: string; new: { path: string }; old: { path: string } }[];
        next?: string;
      };

      for (const v of page.values) {
        entries.push({
          status: v.status,
          path: v.new?.path || v.old?.path,
        });
      }

      url = page.next ?? null;
    }

    return entries;
  }
}

function splitDiff(fullDiff: string): Record<string, string> {
  const patches: Record<string, string> = {};
  const chunks = fullDiff.split('diff --git ');

  for (const chunk of chunks.slice(1)) {
    const newlineIdx = chunk.indexOf('\n');
    if (newlineIdx === -1) continue;
    const header = chunk.slice(0, newlineIdx);
    const parts = header.split(/\s+/);
    if (parts.length < 2) continue;
    const filePath = parts[1].replace(/^b\//, '');
    patches[filePath] = 'diff --git ' + chunk;
  }

  return patches;
}

async function parseBitbucketError(resp: Response): Promise<string> {
  try {
    const body = (await resp.json()) as { error?: { message?: string } };
    if (body.error?.message) {
      return `bitbucket API returned ${resp.status} ${resp.statusText}: ${body.error.message}`;
    }
  } catch {
    // ignore parse errors
  }
  return `bitbucket API returned ${resp.status} ${resp.statusText}`;
}
