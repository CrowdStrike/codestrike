import { Command } from 'commander';
import { config as loadDotenv } from 'dotenv';
import { load as loadConfig } from '../config/config.js';
import { resolvePath } from '../config/path.js';
import { loadEnvConfig, buildLLMClient } from '../setup.js';
import { createLogger } from '../logger.js';
import { createServer } from '../api/routes.js';

export function serveCommand(): Command {
  const cmd = new Command('serve')
    .description('Run CodeStrike as a persistent HTTP service')
    .option('--addr <address>', 'Bind address (default :8000, or PORT env var)')
    .action(async (opts: { addr?: string }) => {
      loadDotenv();

      const envCfg = loadEnvConfig();
      const logger = createLogger(envCfg.logLevel);

      const configFlag = cmd.parent?.getOptionValue('config') as string | undefined;
      const configPath = resolvePath(configFlag);
      const appConfig = loadConfig(configPath);

      if (!envCfg.gitHubToken && !envCfg.bitbucketToken) {
        throw new Error('GITHUB_TOKEN or BITBUCKET_TOKEN environment variable is required');
      }

      const llmClient = buildLLMClient(envCfg);

      let addr = opts.addr ?? '';
      if (!addr) addr = `:${envCfg.listenHttpPort}`;

      const [host, portStr] = parseAddr(addr);
      const port = parseInt(portStr, 10);

      const server = createServer(llmClient, appConfig, envCfg, logger);

      try {
        await server.listen({ port, host: host || '0.0.0.0' });
        logger.info({ addr }, 'server listening');
      } catch (err) {
        logger.error({ err }, 'server error');
        process.exit(1);
      }

      const shutdown = async (signal: string) => {
        logger.info({ signal }, 'shutting down');
        await server.close();
        process.exit(0);
      };

      process.on('SIGINT', () => shutdown('SIGINT'));
      process.on('SIGTERM', () => shutdown('SIGTERM'));
    });

  return cmd;
}

function parseAddr(addr: string): [string, string] {
  if (addr.startsWith(':')) return ['', addr.slice(1)];
  const lastColon = addr.lastIndexOf(':');
  if (lastColon === -1) return ['', addr];
  return [addr.slice(0, lastColon), addr.slice(lastColon + 1)];
}
