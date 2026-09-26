import { Command } from 'commander';
import { shortVersion } from './version.js';
import { reviewCommand } from './cli/review.js';
import { initCommand } from './cli/init.js';
import { versionCommand } from './cli/version.js';
import { serveCommand } from './cli/serve.js';

const program = new Command()
  .name('codestrike')
  .description('codestrike is an AI-driven pull request review tool')
  .version(shortVersion(), '-v, --version')
  .option('--config <path>', 'Path to app config file (default: OS user config dir)');

program.addCommand(reviewCommand());
program.addCommand(initCommand());
program.addCommand(versionCommand());
program.addCommand(serveCommand());

program.parseAsync(process.argv).catch((err) => {
  console.error(err.message);
  process.exit(1);
});
