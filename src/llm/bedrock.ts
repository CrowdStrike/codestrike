import {
  BedrockRuntimeClient,
  InvokeModelCommand,
  type InvokeModelCommandInput,
} from '@aws-sdk/client-bedrock-runtime';
import type { LLMClient, LLMRequest, LLMResponse } from './types.js';

interface ClaudeMessageRequest {
  anthropic_version: string;
  max_tokens: number;
  temperature: number;
  messages: { role: string; content: string }[];
}

interface ClaudeMessageResponse {
  content: { type: string; text: string }[];
  stop_reason: string;
}

const ANTHROPIC_VERSION = 'bedrock-2023-05-31';

export class BedrockClient implements LLMClient {
  private client: BedrockRuntimeClient;
  private modelID: string;
  private maxRetries: number;
  private initialDelay: number;
  private maxDelay: number;

  constructor(region: string, modelID: string) {
    this.client = new BedrockRuntimeClient({ region });
    this.modelID = modelID;
    this.maxRetries = 3;
    this.initialDelay = 100;
    this.maxDelay = 12000;
  }

  async invokeModel(request: LLMRequest): Promise<LLMResponse> {
    const payload: ClaudeMessageRequest = {
      anthropic_version: ANTHROPIC_VERSION,
      max_tokens: request.maxTokens,
      temperature: request.temperature,
      messages: [{ role: 'user', content: request.prompt }],
    };

    const input: InvokeModelCommandInput = {
      modelId: this.modelID,
      body: JSON.stringify(payload),
      accept: 'application/json',
      contentType: 'application/json',
    };

    const output = await this.client.send(new InvokeModelCommand(input));
    const responseBody = new TextDecoder().decode(output.body);
    const response: ClaudeMessageResponse = JSON.parse(responseBody);

    const content = response.content.length > 0 ? response.content[0].text : '';
    return {
      content,
      stopReason: response.stop_reason,
    };
  }

  async invokeModelWithRetry(request: LLMRequest): Promise<LLMResponse> {
    let lastErr: Error | undefined;

    for (let attempt = 0; attempt < this.maxRetries; attempt++) {
      try {
        return await this.invokeModel(request);
      } catch (err) {
        lastErr = err as Error;
        if (!isRetryableError(lastErr)) {
          throw new Error(`non-retryable error: ${lastErr.message}`);
        }
        const delay = calculateBackoff(attempt, this.initialDelay, this.maxDelay);
        await sleep(delay);
      }
    }

    throw new Error(`max retries ${this.maxRetries} exceeded: ${lastErr?.message}`);
  }
}

function isRetryableError(err: Error): boolean {
  const msg = err.message;
  if (
    msg.includes('ThrottlingException') ||
    msg.includes('TooManyRequestsException') ||
    msg.includes('Rate exceeded')
  ) {
    return true;
  }
  if (
    msg.includes('InternalServerException') ||
    msg.includes('ServiceUnavailableException') ||
    msg.includes('500') ||
    msg.includes('503')
  ) {
    return true;
  }
  if (msg.includes('connection reset') || msg.includes('EOF') || msg.includes('timeout')) {
    return true;
  }
  return false;
}

function calculateBackoff(attempt: number, initialDelay: number, maxDelay: number): number {
  let backoff = initialDelay + Math.pow(2, attempt);
  if (backoff > maxDelay) backoff = maxDelay;
  const jitter = backoff * 0.2 * (2 * Math.random() - 1);
  return backoff + jitter;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
