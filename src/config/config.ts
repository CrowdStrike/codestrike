import * as fs from 'fs';
import * as path from 'path';
import { parse as parseYaml } from 'yaml';

export interface Config {
  github: GitHubConfig;
  bitbucket: BitbucketConfig;
  review: ReviewConfig;
}

export interface GitHubConfig {
  base_url: string;
}

export interface BitbucketConfig {
  base_url: string;
}

export interface ReviewConfig {
  system_prompt: string;
  tone: string;
  context_files: string[];
  guardrails: Guardrails;
  context: ContextConfig;
}

export interface ContextConfig {
  max_input_ratio: number;
  reserved_output_tokens: number;
  tokenizer_model: string;
  model_limits: Record<string, ModelLimit>;
  enable_cursor_rules: boolean;
}

export interface ModelLimit {
  context_window: number;
  max_output_tokens: number;
}

export interface Guardrails {
  max_patch_size_bytes: number;
  ignored_paths: string[];
}

export function load(configPath: string): Config {
  let data: string;
  try {
    data = fs.readFileSync(configPath, 'utf-8');
  } catch (err: unknown) {
    if ((err as NodeJS.ErrnoException).code === 'ENOENT') {
      throw new Error(`config file not found at "${configPath}"; pass --config to specify a different path`);
    }
    throw new Error(`reading config file: ${err}`);
  }

  const cfg = parseYaml(data) as Config;

  const configDir = path.dirname(configPath);
  cfg.review.system_prompt = resolveText(configDir, 'prompts', cfg.review.system_prompt);
  cfg.review.tone = resolveText(configDir, 'tones', cfg.review.tone);

  return cfg;
}

function resolveText(configDir: string, subdir: string, value: string): string {
  if (!value) return value;

  for (const name of [value, value + '.md']) {
    if (!name || path.basename(name) !== name) continue;
    const filePath = path.join(configDir, subdir, name);
    try {
      return fs.readFileSync(filePath, 'utf-8');
    } catch (err: unknown) {
      if ((err as NodeJS.ErrnoException).code !== 'ENOENT') {
        throw new Error(`reading "${filePath}": ${err}`);
      }
    }
  }

  return value;
}
