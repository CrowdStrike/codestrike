import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { fileURLToPath } from 'url';

const APP_CONFIG_DIR_NAME = 'codestrike';

function getUserConfigDir(): string {
  switch (process.platform) {
    case 'darwin':
      return path.join(os.homedir(), 'Library', 'Application Support');
    case 'win32':
      return process.env.APPDATA || path.join(os.homedir(), 'AppData', 'Roaming');
    default:
      return process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');
  }
}

function getBundledAssetsDir(): string {
  const thisFile = fileURLToPath(import.meta.url);
  const srcDir = path.dirname(thisFile);
  const assetsInSrc = path.join(srcDir, 'config', 'assets');
  if (fs.existsSync(assetsInSrc)) return assetsInSrc;
  const assetsFromDist = path.join(srcDir, '..', 'src', 'config', 'assets');
  if (fs.existsSync(assetsFromDist)) return assetsFromDist;
  throw new Error('could not locate bundled config assets');
}

export function install(force: boolean): string {
  const userConfigDir = getUserConfigDir();
  const targetDir = path.join(userConfigDir, APP_CONFIG_DIR_NAME);
  const assetsDir = getBundledAssetsDir();

  const entries = walkDir(assetsDir);

  if (!force) {
    for (const relPath of entries) {
      const target = path.join(targetDir, relPath);
      if (fs.existsSync(target)) {
        throw new Error(`config file "${target}" already exists; use --force to overwrite bundled files`);
      }
    }
  }

  for (const relPath of entries) {
    const source = path.join(assetsDir, relPath);
    const target = path.join(targetDir, relPath);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.copyFileSync(source, target);
    fs.chmodSync(target, 0o600);
  }

  return targetDir;
}

function walkDir(dir: string, prefix = ''): string[] {
  const results: string[] = [];
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const entry of entries) {
    const relPath = path.join(prefix, entry.name);
    if (entry.isDirectory()) {
      results.push(...walkDir(path.join(dir, entry.name), relPath));
    } else {
      results.push(relPath);
    }
  }
  return results;
}
