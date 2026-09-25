import { describe, it, expect, vi } from 'vitest';
import { createAgentReviewer } from '../../src/agent/factory.js';
import { AgentType } from '../../src/agent/types.js';
import { ClaudeCodeAgentReviewer } from '../../src/agent/claude/claude-code.js';
import pino from 'pino';

const logger = pino({ level: 'silent' });

describe('createAgentReviewer', () => {
  it('creates a ClaudeCodeAgentReviewer for ClaudeCode type', () => {
    const reviewer = createAgentReviewer(
      AgentType.ClaudeCode,
      { allowedTools: ['Bash', 'Read'], maxTurns: 10 },
      logger,
    );
    expect(reviewer).toBeInstanceOf(ClaudeCodeAgentReviewer);
  });

  it('throws on unsupported agent type', () => {
    expect(() => createAgentReviewer('unknown' as AgentType, { allowedTools: [], maxTurns: 5 }, logger)).toThrow(
      'unsupported agent type',
    );
  });
});
