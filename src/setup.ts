import { getEnvString } from './env.js';
import type { Config } from './config/config.js';
import type { LLMClient } from './llm/types.js';
import { parseFamily } from './llm/types.js';
import { createLLMClient } from './llm/factory.js';
import type { SCMClient } from './scm/types.js';
import { GitHubClient } from './scm/github.js';
import { BitbucketClient } from './scm/bitbucket.js';
import { PROVIDER_GITHUB, PROVIDER_BITBUCKET, type PRReference } from './review/parser.js';
import type { Logger } from './logger.js';

export interface EnvConfig {
  gitHubToken: string;
  bitbucketToken: string;
  openAIURL: string;
  openAIKey: string;
  awsRegion: string;
  ollamaURL: string;
  modelFamily: string;
  modelID: string;
  logLevel: string;
  listenHttpPort: string;
}

export interface Dependencies {
  scmClient: SCMClient;
  llmClient: LLMClient;
  appConfig: Config;
  logger: Logger;
}

export function loadEnvConfig(): EnvConfig {
  return {
    gitHubToken: getEnvString('GITHUB_TOKEN', ''),
    bitbucketToken: getEnvString('BITBUCKET_TOKEN', ''),
    openAIURL: getEnvString('OPEN_AI_BASE_URL', 'https://api.openai.com/v1'),
    openAIKey: getEnvString('OPEN_AI_KEY', ''),
    awsRegion: getEnvString('AWS_REGION', 'us-east-1'),
    ollamaURL: getEnvString('OLLAMA_BASE_URL', 'http://localhost:11434/v1'),
    modelFamily: getEnvString('MODEL_FAMILY', 'openai_platform'),
    modelID: getEnvString('MODEL_ID', 'gpt-4o'),
    logLevel: getEnvString('LOG_LEVEL', 'info'),
    listenHttpPort: getEnvString('HTTP_LISTENING_PORT', '8000'),
  };
}

export function wire(envCfg: EnvConfig, appConfig: Config, logger: Logger, ref: PRReference): Dependencies {
  const scmClient = createSCMClient(envCfg, appConfig, ref);
  const llmClient = buildLLMClient(envCfg);
  return { scmClient, llmClient, appConfig, logger };
}

export function createSCMClient(envCfg: EnvConfig, appConfig: Config, ref: PRReference): SCMClient {
  switch (ref.provider) {
    case PROVIDER_GITHUB: {
      if (!envCfg.gitHubToken) throw new Error('GITHUB_TOKEN environment variable is required for GitHub PRs');
      return new GitHubClient({
        owner: ref.owner,
        repo: ref.repo,
        token: envCfg.gitHubToken,
        baseURL: appConfig.github.base_url,
      });
    }
    case PROVIDER_BITBUCKET: {
      if (!envCfg.bitbucketToken)
        throw new Error('BITBUCKET_TOKEN environment variable is required for Bitbucket PRs');
      return new BitbucketClient({
        workspace: ref.owner,
        repoSlug: ref.repo,
        token: envCfg.bitbucketToken,
        baseURL: appConfig.bitbucket?.base_url ?? '',
      });
    }
    default:
      throw new Error(`unsupported SCM provider: ${ref.provider}`);
  }
}

export function buildLLMClient(envCfg: EnvConfig): LLMClient {
  const family = parseFamily(envCfg.modelFamily);
  return createLLMClient({
    family,
    modelID: envCfg.modelID,
    openAIURL: envCfg.openAIURL,
    openAIKey: envCfg.openAIKey,
    awsRegion: envCfg.awsRegion,
    ollamaURL: envCfg.ollamaURL,
  });
}
