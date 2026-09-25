export enum AgentType {
  ClaudeCode = 'claude-code',
}

export interface AgentConfig {
  allowedTools: string[];
  maxTurns: number;
}

export interface AgentReviewer {
  review(prompt: string, model: string, cwd: string): Promise<string>;
}
