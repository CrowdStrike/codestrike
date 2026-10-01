import type { Logger } from '../logger.js';
import type { Config } from '../config/config.js';
import type { LLMClient } from '../llm/types.js';
import type { EnvConfig } from '../setup.js';
import { createSCMClient } from '../setup.js';
import { parsePrUrl } from '../review/parser.js';
import { Pipeline } from '../review/pipeline.js';
import { newForModel } from '../tokenizer.js';
import type { HealthResponse, ReviewRequest, ReviewResponse, ErrorResponse } from './types.js';

export class Handler {
  private llmClient: LLMClient;
  private appConfig: Config;
  private envConfig: EnvConfig;
  private logger: Logger;

  constructor(llmClient: LLMClient, appConfig: Config, envConfig: EnvConfig, logger: Logger) {
    this.llmClient = llmClient;
    this.appConfig = appConfig;
    this.envConfig = envConfig;
    this.logger = logger;
  }

  healthz(): HealthResponse {
    return { status: 'ok', version: '1.0.0' };
  }

  async review(body: ReviewRequest): Promise<{ status: number; body: ReviewResponse | ErrorResponse }> {
    if (!body.pr_url) {
      return { status: 400, body: { error: 'pr_url is required' } };
    }

    let ref;
    try {
      ref = parsePrUrl(body.pr_url);
    } catch (err) {
      return { status: 400, body: { error: `invalid PR URL: ${(err as Error).message}` } };
    }

    let scmClient;
    try {
      scmClient = createSCMClient(this.envConfig, this.appConfig, ref);
    } catch (err) {
      return { status: 400, body: { error: `creating SCM client: ${(err as Error).message}` } };
    }

    const tok = newForModel(this.appConfig.review.context.tokenizer_model);
    const pipeline = new Pipeline(scmClient, this.llmClient, this.appConfig, tok, this.logger, {
      fullContext: body.full_context ?? false,
      dryRun: false,
    });

    try {
      await pipeline.run(ref);
    } catch (err) {
      this.logger.error({ err, pr_url: body.pr_url }, 'review failed');
      return { status: 500, body: { error: `review failed: ${(err as Error).message}` } };
    }

    return {
      status: 200,
      body: { status: 'reviewed', owner: ref.owner, repo: ref.repo, pr: ref.number },
    };
  }
}
