import type { LLMClient, LLMFamily } from './types.js';
import { OpenAIPlatformClient } from './openai.js';
import { BedrockClient } from './bedrock.js';
import { OllamaClient } from './ollama.js';

export interface LLMFactoryConfig {
  family: LLMFamily;
  modelID: string;
  openAIURL: string;
  openAIKey: string;
  awsRegion: string;
  ollamaURL: string;
}

export function createLLMClient(cfg: LLMFactoryConfig): LLMClient {
  switch (cfg.family) {
    case 'anthropic': {
      if (!cfg.awsRegion) throw new Error('AWS_REGION is required for Anthropic/Bedrock provider');
      return new BedrockClient(cfg.awsRegion, cfg.modelID);
    }
    case 'openai_platform': {
      if (!cfg.openAIKey) throw new Error('OPEN_AI_KEY is required for OpenAI provider');
      return new OpenAIPlatformClient(cfg.openAIURL, cfg.openAIKey, cfg.modelID);
    }
    case 'ollama': {
      if (!cfg.ollamaURL) throw new Error('OLLAMA_BASE_URL is required for Ollama provider');
      return new OllamaClient(cfg.ollamaURL, cfg.modelID);
    }
    default:
      throw new Error(`unsupported model family: ${cfg.family}`);
  }
}
