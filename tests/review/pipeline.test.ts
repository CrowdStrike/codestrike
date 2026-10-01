import { describe, it, expect } from 'vitest';
import { parseResponse } from '../../src/review/pipeline.js';

describe('parseResponse', () => {
  it('parses a valid JSON array of comments', () => {
    const input = '[{"file":"main.go","line":10,"body":"issue here"}]';
    const result = parseResponse(input);
    expect(result).toHaveLength(1);
    expect(result[0].file).toBe('main.go');
    expect(result[0].line).toBe(10);
    expect(result[0].body).toBe('issue here');
  });

  it('strips reasoning block', () => {
    const input = '<reasoning>\nsome analysis\n</reasoning>\n[{"file":"a.ts","line":1,"body":"fix"}]';
    const result = parseResponse(input);
    expect(result).toHaveLength(1);
    expect(result[0].file).toBe('a.ts');
  });

  it('strips markdown code fences', () => {
    const input = '```json\n[{"file":"b.ts","line":2,"body":"bug"}]\n```';
    const result = parseResponse(input);
    expect(result).toHaveLength(1);
  });

  it('returns empty array for empty response', () => {
    expect(parseResponse('')).toEqual([]);
    expect(parseResponse('[]')).toEqual([]);
  });

  it('throws on invalid JSON', () => {
    expect(() => parseResponse('not json')).toThrow();
  });

  it('handles reasoning + code fences together', () => {
    const input = '<reasoning>\nthinking...\n</reasoning>\n\n```json\n[{"file":"c.go","line":5,"body":"check"}]\n```';
    const result = parseResponse(input);
    expect(result).toHaveLength(1);
    expect(result[0].file).toBe('c.go');
  });
});
