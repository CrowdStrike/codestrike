import { describe, it, expect } from 'vitest';
import { parseMDCFrontmatter, isUnconditionallyApplicable } from '../../src/context/cursor-rules.js';

describe('parseMDCFrontmatter', () => {
  it('parses valid frontmatter', () => {
    const data = '---\ndescription: test rule\nalwaysApply: true\n---\nRule body here';
    const { fm, body, ok } = parseMDCFrontmatter(data);
    expect(ok).toBe(true);
    expect(fm.description).toBe('test rule');
    expect(fm.alwaysApply).toBe(true);
    expect(body).toBe('Rule body here');
  });

  it('returns ok=false for missing frontmatter', () => {
    const { ok } = parseMDCFrontmatter('No frontmatter here');
    expect(ok).toBe(false);
  });

  it('returns ok=false for unclosed frontmatter', () => {
    const { ok } = parseMDCFrontmatter('---\nkey: value\nno closing delimiter');
    expect(ok).toBe(false);
  });

  it('parses globs field', () => {
    const data = '---\nglobs: "*.ts"\n---\nBody';
    const { fm, ok } = parseMDCFrontmatter(data);
    expect(ok).toBe(true);
    expect(fm.globs).toBe('*.ts');
  });
});

describe('isUnconditionallyApplicable', () => {
  it('returns true when alwaysApply is true', () => {
    expect(isUnconditionallyApplicable({ alwaysApply: true })).toBe(true);
  });

  it('returns false when alwaysApply is false', () => {
    expect(isUnconditionallyApplicable({ alwaysApply: false })).toBe(false);
  });

  it('returns true when no globs and no alwaysApply', () => {
    expect(isUnconditionallyApplicable({})).toBe(true);
  });

  it('returns false when globs are set and no alwaysApply', () => {
    expect(isUnconditionallyApplicable({ globs: '*.ts' })).toBe(false);
  });
});
