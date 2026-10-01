import * as fs from 'fs';
import * as path from 'path';
import { parse as parseYaml } from 'yaml';

interface CursorRuleFrontmatter {
  description?: string;
  globs?: string;
  alwaysApply?: boolean;
}

export function parseMDCFrontmatter(data: string): { fm: CursorRuleFrontmatter; body: string; ok: boolean } {
  if (!data.startsWith('---\n') && !data.startsWith('---\r\n')) {
    return { fm: {}, body: '', ok: false };
  }

  const firstNewline = data.indexOf('\n');
  const rest = data.slice(firstNewline + 1);

  let end = rest.indexOf('\n---\n');
  let sepLen = '\n---\n'.length;
  if (end === -1) {
    end = rest.indexOf('\n---\r\n');
    sepLen = '\n---\r\n'.length;
  }
  if (end === -1) {
    return { fm: {}, body: '', ok: false };
  }

  const raw = rest.slice(0, end);
  const remainder = rest.slice(end + sepLen);

  try {
    const fm = parseYaml(raw) as CursorRuleFrontmatter;
    return { fm: fm ?? {}, body: remainder.trim(), ok: true };
  } catch {
    return { fm: {}, body: '', ok: false };
  }
}

export function isUnconditionallyApplicable(fm: CursorRuleFrontmatter): boolean {
  if (fm.alwaysApply !== undefined) {
    return fm.alwaysApply;
  }
  return !(fm.globs ?? '').trim();
}

export function loadCursorContext(repoRoot: string): string {
  let result = '';

  try {
    const agentsPath = path.join(repoRoot, 'AGENTS.md');
    const content = fs.readFileSync(agentsPath, 'utf-8').trim();
    if (content) {
      result += `### AGENTS.md\n${content}\n\n`;
    }
  } catch {
    // file not found, skip
  }

  const rulesDir = path.join(repoRoot, '.cursor', 'rules');
  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(rulesDir, { withFileTypes: true });
  } catch {
    return result;
  }

  for (const entry of entries) {
    if (entry.isDirectory() || path.extname(entry.name) !== '.mdc') continue;

    let data: string;
    try {
      data = fs.readFileSync(path.join(rulesDir, entry.name), 'utf-8');
    } catch {
      continue;
    }

    const { fm, body, ok } = parseMDCFrontmatter(data);
    if (!ok || !body || !isUnconditionallyApplicable(fm)) continue;

    result += `### .cursor/rules/${entry.name}\n${body}\n\n`;
  }

  return result;
}
