export interface LLMRequest {
  prompt: string;
  maxTokens: number;
  temperature: number;
}

export interface LLMResponse {
  content: string;
  stopReason: string;
}

export interface LLMClient {
  invokeModel(request: LLMRequest): Promise<LLMResponse>;
  invokeModelWithRetry(request: LLMRequest): Promise<LLMResponse>;
}

export type LLMFamily = 'anthropic' | 'openai_platform' | 'ollama';

const validFamilies: LLMFamily[] = ['anthropic', 'openai_platform', 'ollama'];

export function parseFamily(s: string): LLMFamily {
  if (validFamilies.includes(s as LLMFamily)) {
    return s as LLMFamily;
  }
  throw new Error(`unsupported model family: ${s} (valid: ${validFamilies.join(', ')})`);
}
