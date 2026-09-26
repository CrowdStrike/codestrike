import { get_encoding, encoding_for_model, type Tiktoken } from 'tiktoken';

export interface Tokenizer {
  countTokens(text: string): number;
}

export interface ModelLimits {
  contextWindow: number;
  maxOutputTokens: number;
}

const defaultModelLimits: Record<string, ModelLimits> = {
  'gpt-4o': { contextWindow: 128000, maxOutputTokens: 16384 },
  'gpt-4o-mini': { contextWindow: 128000, maxOutputTokens: 16384 },
  'gpt-4-turbo': { contextWindow: 128000, maxOutputTokens: 4096 },
  'claude-sonnet-4-20250514': { contextWindow: 200000, maxOutputTokens: 8192 },
  'claude-haiku-4-5-20251001': { contextWindow: 200000, maxOutputTokens: 8192 },
  'anthropic.claude-3-5-sonnet-20241022-v2:0': { contextWindow: 200000, maxOutputTokens: 8192 },
  'anthropic.claude-3-haiku-20240307-v1:0': { contextWindow: 200000, maxOutputTokens: 4096 },
  llama3: { contextWindow: 8192, maxOutputTokens: 2048 },
  'llama3:70b': { contextWindow: 8192, maxOutputTokens: 2048 },
  codellama: { contextWindow: 16384, maxOutputTokens: 4096 },
};

export function getModelLimits(modelID: string): ModelLimits {
  return defaultModelLimits[modelID] ?? { contextWindow: 8192, maxOutputTokens: 4096 };
}

const encodingCache = new Map<string, Tiktoken>();

function getEncoding(model: string): Tiktoken {
  if (encodingCache.has(model)) {
    return encodingCache.get(model)!;
  }

  let enc: Tiktoken;
  try {
    enc = get_encoding(model as Parameters<typeof get_encoding>[0]);
  } catch {
    try {
      enc = encoding_for_model(model as Parameters<typeof encoding_for_model>[0]);
    } catch {
      enc = get_encoding('cl100k_base');
    }
  }

  encodingCache.set(model, enc);
  return enc;
}

export class TiktokenTokenizer implements Tokenizer {
  private encoding: Tiktoken;

  constructor(model?: string) {
    this.encoding = getEncoding(model || 'o200k_base');
  }

  countTokens(text: string): number {
    if (!this.encoding) {
      return Math.ceil(text.length / 4);
    }
    return this.encoding.encode(text).length;
  }
}

export function newForModel(model?: string): Tokenizer {
  return new TiktokenTokenizer(model);
}
