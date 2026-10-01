import OpenAI from 'openai';
import type { LLMClient, LLMRequest, LLMResponse } from './types.js';

export class OllamaClient implements LLMClient {
  private client: OpenAI;
  private modelID: string;

  constructor(baseURL: string, modelID: string) {
    this.client = new OpenAI({
      apiKey: 'ollama',
      baseURL,
      maxRetries: 3,
    });
    this.modelID = modelID;
  }

  async invokeModel(request: LLMRequest): Promise<LLMResponse> {
    const response = await this.client.chat.completions.create({
      model: this.modelID,
      messages: [{ role: 'user', content: request.prompt }],
      max_completion_tokens: request.maxTokens,
      temperature: request.temperature,
    });

    if (!response.choices.length) {
      throw new Error('no choices in response');
    }

    const choice = response.choices[0];
    return {
      content: choice.message.content ?? '',
      stopReason: choice.finish_reason ?? '',
    };
  }

  async invokeModelWithRetry(request: LLMRequest): Promise<LLMResponse> {
    return this.invokeModel(request);
  }
}
