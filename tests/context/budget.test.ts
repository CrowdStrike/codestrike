import { describe, it, expect } from 'vitest';
import { Budget } from '../../src/context/budget.js';
import type { Tokenizer } from '../../src/tokenizer.js';

class FakeTokenizer implements Tokenizer {
  countTokens(text: string): number {
    return text.length;
  }
}

describe('Budget', () => {
  const tok = new FakeTokenizer();

  it('calculates available input tokens', () => {
    const b = new Budget(tok, 100000, 0.75, 4096);
    expect(b.availableInputTokens()).toBe(100000 * 0.75 - 4096);
  });

  it('clamps invalid maxInputRatio to 0.75', () => {
    const b = new Budget(tok, 100000, -1, 4096);
    expect(b.availableInputTokens()).toBe(100000 * 0.75 - 4096);
  });

  it('allocates sections by priority', () => {
    const b = new Budget(tok, 100000, 0.75, 4096);
    const available = b.availableInputTokens();
    const alloc = b.allocate(1000, 500, 200, 0);

    expect(alloc.systemPrompt).toBe(1000);
    expect(alloc.existingComments).toBeLessThanOrEqual(Math.floor(available * 0.15));
    expect(alloc.existingComments).toBe(500);
    expect(alloc.memory).toBeLessThanOrEqual(Math.floor(available * 0.1));
    expect(alloc.memory).toBe(200);
  });

  it('caps comments at 15% of available', () => {
    const b = new Budget(tok, 100000, 0.75, 4096);
    const available = b.availableInputTokens();
    const alloc = b.allocate(100, 999999, 0, 0);
    expect(alloc.existingComments).toBe(Math.floor(available * 0.15));
  });

  it('caps memory at 10% of available', () => {
    const b = new Budget(tok, 100000, 0.75, 4096);
    const available = b.availableInputTokens();
    const alloc = b.allocate(100, 0, 999999, 0);
    expect(alloc.memory).toBe(Math.floor(available * 0.1));
  });

  it('fitsInSingleCall returns true when under budget', () => {
    const b = new Budget(tok, 100000, 0.75, 4096);
    expect(b.fitsInSingleCall(1000)).toBe(true);
    expect(b.fitsInSingleCall(999999)).toBe(false);
  });
});
