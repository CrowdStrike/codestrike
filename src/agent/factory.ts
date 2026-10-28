import { Logger } from '../logger.js';
import { AgentReviewer } from './types.js';
import { ClaudeCodeAgentReviewer } from './claude/claude-code.js';
import { AgentConfig, AgentType } from './types.js';

export function createAgentReviewer(agentType: AgentType, agentConfig: AgentConfig, logger: Logger): AgentReviewer {
  switch (agentType) {
    case AgentType.ClaudeCode:
      return new ClaudeCodeAgentReviewer(agentConfig, logger);
    default:
      throw new Error(`unsupported agent type: ${agentType}`);
  }
}
