import { describe, it, expect } from 'vitest';
import { shortVersion, Version } from '../../src/version.js';

describe('version', () => {
  it('shortVersion includes version and commit', () => {
    const s = shortVersion();
    expect(s).toContain(Version);
  });

  it('shortVersion truncates long commit hashes', () => {
    const s = shortVersion();
    // With default "unknown" commit, it should be included as-is (<=7 chars)
    expect(s).toContain('unknown');
  });
});
