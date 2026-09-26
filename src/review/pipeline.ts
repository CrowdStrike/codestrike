import * as fs from 'fs';
import * as path from 'path';
import type { Logger } from '../logger.js';
import type { Config } from '../config/config.js';
import type { LLMClient } from '../llm/types.js';
import type { SCMClient, PullRequestFile, ReviewComment, PRComment } from '../scm/types.js';
import type { Tokenizer } from '../tokenizer.js';
import { Budget } from '../context/budget.js';
import { Builder } from '../context/builder.js';
import { loadCursorContext } from '../context/cursor-rules.js';
import type { PRReference } from './parser.js';

export interface PipelineOptions {
  fullContext: boolean;
  dryRun: boolean;
}

export class Pipeline {
  private client: SCMClient;
  private llmClient: LLMClient;
  private config: Config;
  private tokenizer: Tokenizer;
  private fullContext: boolean;
  private dryRun: boolean;
  private logger: Logger;

  constructor(
    client: SCMClient,
    llmClient: LLMClient,
    config: Config,
    tokenizer: Tokenizer,
    logger: Logger,
    opts: PipelineOptions,
  ) {
    this.client = client;
    this.llmClient = llmClient;
    this.config = config;
    this.tokenizer = tokenizer;
    this.fullContext = opts.fullContext;
    this.dryRun = opts.dryRun;
    this.logger = logger;
  }

  async run(ref: PRReference): Promise<string> {
    // Step 1: Verify PR exists
    const exists = await this.client.pullRequestExists(ref.number);
    if (!exists) {
      throw new Error(`pull request #${ref.number} not found in ${ref.owner}/${ref.repo}`);
    }
    this.logger.info({ pr: ref.number }, 'pull request found');

    // Step 1b: Fetch PR description
    let prDescription = '';
    try {
      const { title, body } = await this.client.getPullRequestDescription(ref.number);
      prDescription = `**Title:** ${title}\n`;
      if (body) prDescription += body + '\n';
    } catch (err) {
      this.logger.warn({ err }, 'failed to fetch PR description, continuing without');
    }

    // Step 2: Fetch PR files
    const files = await this.client.getPullRequestFiles(ref.number);
    this.logger.info({ total_files: files.length }, 'fetched PR files');

    // Step 3: Apply guardrails
    let filtered = this.applyGuardrails(files);
    this.logger.info({ filtered_files: filtered.length }, 'files after guardrails');

    if (filtered.length === 0) {
      this.logger.warn('no files to review after applying guardrails');
      return '';
    }

    // Step 3b: Fetch full file content if requested
    if (this.fullContext) {
      filtered = await this.enrichWithContent(filtered);
    }

    // Step 4: Build context-aware prompts
    const budget = this.createBudget();
    const builder = new Builder(this.tokenizer, budget, this.config);

    const projectContext = this.loadProjectContext();
    const { ownComments, userFeedback } = await this.fetchExistingCommentsContext(ref.number);
    const memoryContext = '';

    const result = builder.build(filtered, ownComments, userFeedback, memoryContext, projectContext, prDescription);
    this.logger.info({ total_tokens: result.totalTokens, skipped_files: result.skippedFiles.length }, 'context built');

    if (result.skippedFiles.length > 0) {
      this.logger.warn({ files: result.skippedFiles }, 'files skipped due to context budget');
    }

    // Step 5: Run inference
    const comments = await this.runInference(result.prompt, result.outputMaxToken);

    // Step 6: Validate output
    const validated = this.validateComments(comments, filtered);
    this.logger.info({ valid_comments: validated.length }, 'comments validated');

    if (validated.length === 0) {
      this.logger.info('no actionable comments to post');
      return '';
    }

    // Step 7: Post comments
    const body = formatComments(validated);
    if (this.dryRun) {
      this.logger.info({ pr: ref.number }, 'review computed (dry run)');
      return body;
    }

    await this.client.publishComment(ref.number, body);
    this.logger.info({ pr: ref.number }, 'review posted');
    return '';
  }

  private createBudget(): Budget {
    const ctxCfg = this.config.review.context;
    const contextWindow = 128000;
    let reservedOutput = 4096;
    let maxInputRatio = 0.75;

    if (ctxCfg.reserved_output_tokens > 0) reservedOutput = ctxCfg.reserved_output_tokens;
    if (ctxCfg.max_input_ratio > 0) maxInputRatio = ctxCfg.max_input_ratio;

    return new Budget(this.tokenizer, contextWindow, maxInputRatio, reservedOutput);
  }

  private async runInference(prompt: string, maxTokens: number): Promise<ReviewComment[]> {
    const resp = await this.llmClient.invokeModelWithRetry({
      prompt,
      maxTokens,
      temperature: 0.2,
    });
    return parseResponse(resp.content);
  }

  private applyGuardrails(files: PullRequestFile[]): PullRequestFile[] {
    const guardrails = this.config.review.guardrails;
    const result: PullRequestFile[] = [];

    for (const f of files) {
      if (this.isIgnoredPath(f.filename, guardrails.ignored_paths)) {
        this.logger.debug({ file: f.filename }, 'skipped: ignored path');
        continue;
      }
      if (guardrails.max_patch_size_bytes > 0 && f.patch.length > guardrails.max_patch_size_bytes) {
        this.logger.debug({ file: f.filename }, 'skipped: exceeds max patch size');
        continue;
      }
      if (f.status === 'removed') {
        this.logger.debug({ file: f.filename }, 'skipped: file removed');
        continue;
      }
      result.push(f);
    }

    return result;
  }

  private isIgnoredPath(filename: string, patterns: string[]): boolean {
    for (const pattern of patterns) {
      if (pattern.endsWith('/') && filename.startsWith(pattern)) return true;
      if (matchGlob(pattern, filename)) return true;
      if (matchGlob(pattern, path.basename(filename))) return true;
    }
    return false;
  }

  private validateComments(comments: ReviewComment[], files: PullRequestFile[]): ReviewComment[] {
    const fileSet = new Set(files.map((f) => f.filename));
    return comments.filter((c) => {
      if (!fileSet.has(c.file)) {
        this.logger.debug({ file: c.file }, 'comment dropped: file not in diff');
        return false;
      }
      if (c.line <= 0) {
        this.logger.debug({ file: c.file }, 'comment dropped: invalid line number');
        return false;
      }
      if (!c.body.trim()) return false;
      return true;
    });
  }

  private async fetchExistingCommentsContext(prNumber: number): Promise<{ ownComments: string; userFeedback: string }> {
    let general: PRComment[];
    try {
      general = await this.client.getPRComments(prNumber);
    } catch (err) {
      this.logger.warn({ err }, 'failed to fetch PR comments, continuing without');
      return { ownComments: '', userFeedback: '' };
    }

    let inline: PRComment[];
    try {
      inline = await this.client.getPRReviewComments(prNumber);
    } catch (err) {
      this.logger.warn({ err }, 'failed to fetch review comments, continuing without');
      return { ownComments: '', userFeedback: '' };
    }

    const all = [...general, ...inline];
    if (all.length === 0) return { ownComments: '', userFeedback: '' };

    const codestrikeComments: PRComment[] = [];
    const otherComments: PRComment[] = [];
    for (const c of all) {
      if (c.body.includes('<!-- codestrike:review -->')) {
        codestrikeComments.push(c);
      } else {
        otherComments.push(c);
      }
    }

    let ownBuf = '';
    for (const c of codestrikeComments) {
      if (c.path && c.line > 0) {
        ownBuf += `- [${c.path}:${c.line}] "${truncate(c.body)}"\n`;
      } else {
        ownBuf += `- (general) "${truncate(c.body)}"\n`;
      }
    }

    if (codestrikeComments.length === 0) {
      return { ownComments: ownBuf, userFeedback: '' };
    }

    let latestReview = new Date(0);
    for (const c of codestrikeComments) {
      if (c.createdAt > latestReview) latestReview = c.createdAt;
    }

    let feedback = otherComments.filter((c) => c.createdAt > latestReview);
    feedback.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
    if (feedback.length > 10) feedback = feedback.slice(0, 10);

    let feedbackBuf = '';
    for (const c of feedback) {
      if (c.path && c.line > 0) {
        feedbackBuf += `- [${c.path}:${c.line}] "${truncate(c.body)}" — author: ${c.author}\n`;
      } else {
        feedbackBuf += `- (general) "${truncate(c.body)}" — author: ${c.author}\n`;
      }
    }

    return { ownComments: ownBuf, userFeedback: feedbackBuf };
  }

  private async enrichWithContent(files: PullRequestFile[]): Promise<PullRequestFile[]> {
    for (let i = 0; i < files.length; i++) {
      if (files[i].status === 'added') continue;
      try {
        files[i].content = await this.client.getFileContent(files[i].filename, '');
      } catch (err) {
        this.logger.debug({ file: files[i].filename, err }, 'could not fetch full content');
      }
    }
    this.logger.info({ files_enriched: files.length }, 'fetched full file content');
    return files;
  }

  private loadProjectContext(): string {
    let result = this.loadContextFiles();
    if (this.config.review.context.enable_cursor_rules) {
      try {
        result += loadCursorContext(process.cwd());
      } catch (err) {
        this.logger.debug({ err }, 'could not load cursor context');
      }
    }
    return result;
  }

  private loadContextFiles(): string {
    const files = this.config.review.context_files;
    if (!files || files.length === 0) return '';

    let result = '';
    for (const filePath of files) {
      try {
        const content = fs.readFileSync(filePath, 'utf-8').trim();
        if (content) {
          result += `### ${path.basename(filePath)}\n${content}\n\n`;
        }
      } catch {
        this.logger.debug({ file: filePath }, 'context file not found, skipping');
      }
    }
    return result;
  }
}

export function parseResponse(content: string): ReviewComment[] {
  if (content.includes('</reasoning>')) {
    const idx = content.indexOf('</reasoning>');
    content = content.slice(idx + '</reasoning>'.length);
  }

  content = content.trim();
  content = content.replace(/^```json\s*/, '').replace(/^```\s*/, '');
  content = content.replace(/\s*```$/, '');
  content = content.trim();

  if (content === '[]' || content === '') return [];

  const comments = JSON.parse(content) as ReviewComment[];
  return comments;
}

function formatComments(comments: ReviewComment[]): string {
  let s = '<!-- codestrike:review -->\n## Code Review\n\n';
  for (const c of comments) {
    s += `**${c.file}:${c.line}**\n${c.body}\n\n`;
  }
  return s;
}

function truncate(s: string): string {
  s = s.replace(/\n/g, ' ');
  if (s.length > 120) return s.slice(0, 120) + '...';
  return s;
}

function matchGlob(pattern: string, text: string): boolean {
  const regex = pattern
    .replace(/[.+^${}()|[\]\\]/g, '\\$&')
    .replace(/\*/g, '.*')
    .replace(/\?/g, '.');
  return new RegExp(`^${regex}$`).test(text);
}
