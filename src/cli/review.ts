import * as path from 'path';
import * as fs from 'fs';
import { Command } from 'commander';
import { config as loadDotenv } from 'dotenv';
import { load as loadConfig } from '../config/config.js';
import { resolvePath } from '../config/path.js';
import { parsePrUrl } from '../review/parser.js';
import { Pipeline } from '../review/pipeline.js';
import { loadEnvConfig, wire } from '../setup.js';
import { createLogger } from '../logger.js';
import { newForModel } from '../tokenizer.js';

export function reviewCommand(): Command {
  const cmd = new Command('review')
    .description('Run an AI review on a pull request and post the result as a comment (or print it with --dry-run)')
    .argument('<pr-url>', 'Pull request URL')
    .option('--full-context', 'Fetch full file content for richer reviews (slower, uses more tokens)', false)
    .option('--persona <name>', 'Review persona — maps to a prompt file in prompts/ (e.g., security, performance)')
    .option('--dry-run', 'Print the review to stdout instead of posting it as a PR comment', false)
    .action(async (prUrl: string, opts: { fullContext: boolean; persona?: string; dryRun: boolean }) => {
      loadDotenv();

      const envCfg = loadEnvConfig();
      const logger = createLogger(envCfg.logLevel);

      const configFlag = cmd.parent?.getOptionValue('config') as string | undefined;
      const configPath = resolvePath(configFlag);
      const appConfig = loadConfig(configPath);

      if (opts.persona) {
        if (path.basename(opts.persona) !== opts.persona) {
          throw new Error(`persona "${opts.persona}" must be a plain name, not a path`);
        }
        const promptPath = path.join(path.dirname(configPath), 'prompts', opts.persona + '.md');
        const data = fs.readFileSync(promptPath, 'utf-8');
        appConfig.review.system_prompt = data;
      }

      const ref = parsePrUrl(prUrl);
      const deps = wire(envCfg, appConfig, logger, ref);

      const tok = newForModel(appConfig.review.context.tokenizer_model);
      const pipeline = new Pipeline(deps.scmClient, deps.llmClient, deps.appConfig, tok, deps.logger, {
        fullContext: opts.fullContext,
        dryRun: opts.dryRun,
      });

      const body = await pipeline.run(ref);
      if (body) console.log(body);
    });

  return cmd;
}
