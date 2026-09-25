import { describe, it, expect } from 'vitest';
import { parsePrUrl, PROVIDER_GITHUB, PROVIDER_BITBUCKET } from '../../src/review/parser.js';

describe('parsePrUrl', () => {
  it('parses a GitHub PR URL', () => {
    const ref = parsePrUrl('https://github.com/CrowdStrike/codestrike/pull/42');
    expect(ref.provider).toBe(PROVIDER_GITHUB);
    expect(ref.owner).toBe('CrowdStrike');
    expect(ref.repo).toBe('codestrike');
    expect(ref.number).toBe(42);
  });

  it('parses a GitHub PR URL with trailing slash', () => {
    const ref = parsePrUrl('https://github.com/CrowdStrike/codestrike/pull/42/');
    expect(ref.provider).toBe(PROVIDER_GITHUB);
    expect(ref.number).toBe(42);
  });

  it('parses a Bitbucket PR URL', () => {
    const ref = parsePrUrl('https://bitbucket.org/myworkspace/myrepo/pull-requests/7');
    expect(ref.provider).toBe(PROVIDER_BITBUCKET);
    expect(ref.owner).toBe('myworkspace');
    expect(ref.repo).toBe('myrepo');
    expect(ref.number).toBe(7);
  });

  it('throws on missing /pull/ segment', () => {
    expect(() => parsePrUrl('https://github.com/CrowdStrike/codestrike')).toThrow(
      'missing /pull/ or /pull-requests/ segment',
    );
  });

  it('throws on non-numeric PR number', () => {
    expect(() => parsePrUrl('https://github.com/CrowdStrike/codestrike/pull/abc')).toThrow('invalid PR number');
  });

  it('throws on URL with insufficient segments', () => {
    expect(() => parsePrUrl('https://pull/42')).toThrow('missing owner or repo');
  });
});
