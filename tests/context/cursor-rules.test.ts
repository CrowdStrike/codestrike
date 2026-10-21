import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { describe, it, expect, afterEach } from 'vitest';
import { parseMDCFrontmatter, isUnconditionallyApplicable, loadCursorContext } from '../../src/context/cursor-rules.js';

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

describe('loadCursorContext', () => {
  const tempDirs: string[] = [];

  function makeTempDir(): string {
    const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'codestrike-test-'));
    tempDirs.push(dir);
    return dir;
  }

  afterEach(() => {
    for (const dir of tempDirs) {
      fs.rmSync(dir, { recursive: true, force: true });
    }
    tempDirs.length = 0;
  });

  it('reads a normal AGENTS.md file', () => {
    const repo = makeTempDir();
    fs.writeFileSync(path.join(repo, 'AGENTS.md'), 'Project guidance here');

    const result = loadCursorContext(repo);
    expect(result).toContain('### AGENTS.md');
    expect(result).toContain('Project guidance here');
  });

  it('reads normal .cursor/rules/*.mdc files', () => {
    const repo = makeTempDir();
    const rulesDir = path.join(repo, '.cursor', 'rules');
    fs.mkdirSync(rulesDir, { recursive: true });
    fs.writeFileSync(
      path.join(rulesDir, 'style.mdc'),
      '---\nalwaysApply: true\n---\nUse semicolons',
    );

    const result = loadCursorContext(repo);
    expect(result).toContain('### .cursor/rules/style.mdc');
    expect(result).toContain('Use semicolons');
  });

  it('skips symlinked AGENTS.md pointing outside repo', () => {
    const repo = makeTempDir();
    const outside = makeTempDir();
    const secretFile = path.join(outside, 'secret.txt');
    fs.writeFileSync(secretFile, 'TOP SECRET CREDENTIALS');
    fs.symlinkSync(secretFile, path.join(repo, 'AGENTS.md'));

    const result = loadCursorContext(repo);
    expect(result).not.toContain('TOP SECRET CREDENTIALS');
    expect(result).not.toContain('### AGENTS.md');
  });

  it('skips symlinked .mdc file pointing outside repo', () => {
    const repo = makeTempDir();
    const outside = makeTempDir();
    const secretFile = path.join(outside, 'leaked.txt');
    fs.writeFileSync(secretFile, '---\nalwaysApply: true\n---\nSECRET PAYLOAD');

    const rulesDir = path.join(repo, '.cursor', 'rules');
    fs.mkdirSync(rulesDir, { recursive: true });
    fs.symlinkSync(secretFile, path.join(rulesDir, 'evil.mdc'));

    const result = loadCursorContext(repo);
    expect(result).not.toContain('SECRET PAYLOAD');
  });

  it('skips when .cursor/rules is a symlinked directory pointing outside repo', () => {
    const repo = makeTempDir();
    const outside = makeTempDir();
    const externalRulesDir = path.join(outside, 'rules');
    fs.mkdirSync(externalRulesDir, { recursive: true });
    fs.writeFileSync(
      path.join(externalRulesDir, 'injected.mdc'),
      '---\nalwaysApply: true\n---\nINJECTED CONTENT',
    );

    fs.mkdirSync(path.join(repo, '.cursor'), { recursive: true });
    fs.symlinkSync(externalRulesDir, path.join(repo, '.cursor', 'rules'));

    const result = loadCursorContext(repo);
    expect(result).not.toContain('INJECTED CONTENT');
  });

  it('handles broken symlinks gracefully', () => {
    const repo = makeTempDir();
    fs.symlinkSync('/nonexistent/path/file.md', path.join(repo, 'AGENTS.md'));

    const rulesDir = path.join(repo, '.cursor', 'rules');
    fs.mkdirSync(rulesDir, { recursive: true });
    fs.symlinkSync('/nonexistent/path/rule.mdc', path.join(rulesDir, 'broken.mdc'));

    const result = loadCursorContext(repo);
    expect(result).toBe('');
  });

  it('returns empty string when no context files exist', () => {
    const repo = makeTempDir();
    const result = loadCursorContext(repo);
    expect(result).toBe('');
  });
});
