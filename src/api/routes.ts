import Fastify from 'fastify';
import type { Logger } from '../logger.js';
import type { Config } from '../config/config.js';
import type { LLMClient } from '../llm/types.js';
import type { EnvConfig } from '../setup.js';
import { Handler } from './handler.js';
import type { ReviewRequest } from './types.js';

export function createServer(llmClient: LLMClient, appConfig: Config, envConfig: EnvConfig, logger: Logger) {
  const handler = new Handler(llmClient, appConfig, envConfig, logger);
  const app = Fastify({ logger: false });

  app.get('/api/v1/healthz', async (_request, reply) => {
    return reply.code(200).send(handler.healthz());
  });

  app.post<{ Body: ReviewRequest }>('/api/v1/review', async (request, reply) => {
    const result = await handler.review(request.body);
    return reply.code(result.status).send(result.body);
  });

  return app;
}
