import { describe, it, expect } from 'vitest';
import { resolvePath } from '../../src/config/path.js';

describe('resolvePath', () => {
  it('returns flag value when provided', () => {
    expect(resolvePath('/custom/path.yaml')).toBe('/custom/path.yaml');
  });

  it('returns default path when no flag', () => {
    const result = resolvePath();
    expect(result).toContain('codestrike');
    expect(result).toContain('default.yaml');
  });

  it('uses platform-appropriate config dir', () => {
    const result = resolvePath();
    if (process.platform === 'darwin') {
      expect(result).toContain('Library/Application Support');
    } else if (process.platform !== 'win32') {
      expect(result).toContain('.config');
    }
  });
});
