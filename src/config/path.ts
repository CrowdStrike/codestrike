import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

const APP_CONFIG_DIR_NAME = 'codestrike';
const APP_CONFIG_FILE_NAME = 'default.yaml';

export function resolvePath(flagValue?: string): string {
  if (flagValue) return flagValue;
  return path.join(getUserConfigDir(), APP_CONFIG_DIR_NAME, APP_CONFIG_FILE_NAME);
}

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
