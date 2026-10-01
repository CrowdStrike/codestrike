import { describe, it, expect, vi } from 'vitest';
import { ClaudeCodeAgentReviewer } from '../../src/agent/claude/claude-code.js';
import pino from 'pino';

vi.mock('@anthropic-ai/claude-agent-sdk', () => ({
  query: vi.fn(),
}));

import { query } from '@anthropic-ai/claude-agent-sdk';

const logger = pino({ level: 'silent' });

function makeReviewer() {
  return new ClaudeCodeAgentReviewer({ allowedTools: ['Bash', 'Read'], maxTurns: 10 }, logger);
}

async function* fakeStream(messages: unknown[]) {
  for (const msg of messages) {
    yield msg;
  }
}

describe('ClaudeCodeAgentReviewer', () => {
  it('returns result on success', async () => {
    vi.mocked(query).mockReturnValue(
      fakeStream([
        { type: 'status', message: 'working...' },
        { type: 'result', subtype: 'success', result: 'Looks good!', total_cost_usd: 0.05 },
      ]) as ReturnType<typeof query>,
    );

    const reviewer = makeReviewer();
    const result = await reviewer.review('Review this code', 'claude-opus-5', '/tmp/repo');

    expect(result).toBe('Looks good!');
    expect(query).toHaveBeenCalledWith({
      prompt: 'Review this code',
      options: {
        model: 'claude-opus-5',
        cwd: '/tmp/repo',
        allowedTools: ['Bash', 'Read'],
        maxTurns: 10,
      },
    });
  });

  it('throws on error result', async () => {
    vi.mocked(query).mockReturnValue(
      fakeStream([
        { type: 'result', subtype: 'error_max_turns', errors: ['exceeded max turns'], is_error: true },
      ]) as ReturnType<typeof query>,
    );

    const reviewer = makeReviewer();
    await expect(reviewer.review('Review this', 'claude-opus-5', '/tmp/repo')).rejects.toThrow(
      'Agent failed: error_max_turns - exceeded max turns',
    );
  });

  it('returns empty string when no result message', async () => {
    vi.mocked(query).mockReturnValue(fakeStream([{ type: 'status', message: 'done' }]) as ReturnType<typeof query>);

    const reviewer = makeReviewer();
    const result = await reviewer.review('Review this', 'claude-opus-5', '/tmp/repo');

    expect(result).toBe('');
  });
});
