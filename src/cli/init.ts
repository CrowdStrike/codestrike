import { Command } from 'commander';
import { install } from '../config/install.js';

export function initCommand(): Command {
  return new Command('init')
    .description('Install the bundled configuration in the user config directory')
    .option('--force', 'Overwrite existing bundled configuration files', false)
    .action((opts: { force: boolean }) => {
      const dir = install(opts.force);
      console.log(`Configuration installed in ${dir}`);
    });
}
