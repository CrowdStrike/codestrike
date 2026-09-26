import { Command } from 'commander';
import { Version, Commit, CommitDate, BuiltBy, BuildDate, shortVersion } from '../version.js';

export function versionCommand(): Command {
  return new Command('version')
    .description('Print the version')
    .option('--long', 'Print full build information', false)
    .action((opts: { long: boolean }) => {
      if (opts.long) {
        console.log(`Version: ${Version}`);
        console.log();
        console.log(`Commit:      ${Commit}`);
        console.log(`Commit Date: ${CommitDate}`);
        console.log(`Built By:    ${BuiltBy}`);
        console.log(`Build Date:  ${BuildDate}`);
      } else {
        console.log(`codestrike ${shortVersion()}`);
      }
    });
}
