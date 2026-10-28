import { query } from '@anthropic-ai/claude-agent-sdk';
import { AgentReviewer } from '../types.js';
import { Logger } from '../../logger.js';
import { AgentConfig } from '../types.js';

export class ClaudeCodeAgentReviewer implements AgentReviewer {
  private allowedTools: string[];
  private maxTurns: number;
  private logger: Logger;

  constructor(agentConfig: AgentConfig, logger: Logger) {
    this.allowedTools = agentConfig.allowedTools;
    this.maxTurns = agentConfig.maxTurns;
    this.logger = logger;
  }

  async review(prompt: string, model: string, cwd: string): Promise<string> {
    const stream = query({
      prompt,
      options: {
        model: model,
        cwd,
        allowedTools: this.allowedTools,
        maxTurns: this.maxTurns,
      },
    });

    let result = '';

    for await (const message of stream) {
      if (message.type === 'result') {
        if (message.subtype === 'success') {
          result = message.result;
          this.logger.info({ costUsd: message.total_cost_usd }, 'Agent review completed');
        } else {
          // error during execution in max turns.
          throw new Error(`Agent failed: ${message.subtype} - ${message.errors.join(', ')}`);
        }
      }
    }

    return result;
  }
}
