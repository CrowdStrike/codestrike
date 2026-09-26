import { describe, it, expect } from 'vitest';
import { getModelLimits } from '../src/tokenizer.js';

describe('getModelLimits', () => {
  it('returns known limits for gpt-4o', () => {
    const limits = getModelLimits('gpt-4o');
    expect(limits.contextWindow).toBe(128000);
    expect(limits.maxOutputTokens).toBe(16384);
  });

  it('returns known limits for claude', () => {
    const limits = getModelLimits('claude-sonnet-4-20250514');
    expect(limits.contextWindow).toBe(200000);
    expect(limits.maxOutputTokens).toBe(8192);
  });

  it('returns fallback limits for unknown model', () => {
    const limits = getModelLimits('unknown-model');
    expect(limits.contextWindow).toBe(8192);
    expect(limits.maxOutputTokens).toBe(4096);
  });
});
