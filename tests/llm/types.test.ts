import { describe, it, expect } from 'vitest';
import { parseFamily } from '../../src/llm/types.js';

describe('parseFamily', () => {
  it('parses anthropic', () => {
    expect(parseFamily('anthropic')).toBe('anthropic');
  });

  it('parses openai_platform', () => {
    expect(parseFamily('openai_platform')).toBe('openai_platform');
  });

  it('parses ollama', () => {
    expect(parseFamily('ollama')).toBe('ollama');
  });

  it('throws on unsupported family', () => {
    expect(() => parseFamily('invalid')).toThrow('unsupported model family');
  });
});
