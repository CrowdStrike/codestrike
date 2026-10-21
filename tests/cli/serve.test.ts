import { describe, it, expect } from 'vitest';
import { parseAddr, resolveBindHost } from '../../src/cli/serve.js';

describe('parseAddr', () => {
  it('handles bare port with colon prefix', () => {
    expect(parseAddr(':8080')).toEqual(['', '8080']);
  });

  it('handles host and port', () => {
    expect(parseAddr('0.0.0.0:9090')).toEqual(['0.0.0.0', '9090']);
  });

  it('handles port without colon', () => {
    expect(parseAddr('8080')).toEqual(['', '8080']);
  });
});

describe('resolveBindHost', () => {
  it('returns 127.0.0.1 for empty host (secure default)', () => {
    expect(resolveBindHost('')).toBe('127.0.0.1');
  });

  it('returns explicit 0.0.0.0 when operator opts in', () => {
    expect(resolveBindHost('0.0.0.0')).toBe('0.0.0.0');
  });

  it('preserves arbitrary explicit host', () => {
    expect(resolveBindHost('10.0.0.5')).toBe('10.0.0.5');
  });
});
