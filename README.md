![CrowdStrike Logo (Light)](https://raw.githubusercontent.com/CrowdStrike/.github/main/assets/cs-logo-light-mode.png#gh-light-mode-only)
![CrowdStrike Logo (Dark)](https://raw.githubusercontent.com/CrowdStrike/.github/main/assets/cs-logo-dark-mode.png#gh-dark-mode-only)

# codestrike

AI-driven pull request review tool using configurable LLM judges.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)

**codestrike** is an open-source, AI-driven pull request review tool written in Go. Given a pull request, it runs an LLM-powered review over the diff, and posts the results back as a single comment.

> [!IMPORTANT]
> **Public Preview**: This project is currently in public preview and under active development. Features and functionality may change before the stable 1.0 release. While we encourage exploration and testing, please avoid production deployments. We welcome your feedback through [GitHub Issues](https://github.com/crowdstrike/codestrike/issues) to help shape the final release.

## Prerequisites

- Go 1.25+
- A GitHub personal access token with `repo` scope

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/CrowdStrike/codestrike.git
cd codestrike
```

### 2. Configure environment

Copy the example `.env` file and fill in your secrets:

```bash
cp .env.example .env
```

Required variables:

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub personal access token with `repo` scope |
| `OPEN_AI_BASE_URL` | OpenAI-compatible API base URL (default: `https://api.openai.com/v1`) |
| `OPEN_AI_KEY` | OpenAI API key (or compatible provider) |
| `LOG_LEVEL` | Log level: `debug`, `info`, `warn`, `error` (default: `info`) |

### 3. Review application config

Application settings (system prompt, tone, guardrails) are loaded from a YAML file. By default, codestrike looks for this file at the OS-specific user config directory under a `codestrike/` subdirectory:

| OS | Default path |
|----|--------------|
| Linux | `$XDG_CONFIG_HOME/codestrike/default.yaml` (falls back to `~/.config/codestrike/default.yaml`) |
| macOS | `~/Library/Application Support/codestrike/default.yaml` |
| Windows | `%AppData%\codestrike\default.yaml` |

Install the default configuration, prompts, and tones from the files embedded
in the binary:

```bash
codestrike init
```

Existing files are preserved. To overwrite the bundled files with the versions
from the current binary:

```bash
codestrike init --force
```

If no file exists at the default path, codestrike exits with an error explaining
the path it looked for. Use `--config <path>` to point at a config file anywhere
else instead.

You can customize the system prompt, tone, and guardrails:

```yaml
github:
  base_url: https://api.github.com

review:
  # Use inline text, or a file name from prompts/ or tones/ next to this file.
  system_prompt: default
  tone: diplomatic
  guardrails:
    max_patch_size_bytes: 1048576
    ignored_paths:
      - vendor/
      - node_modules/
      - "*.lock"
      - "*.min.js"
```

`ignored_paths` accepts directory prefixes and glob patterns. Globs are matched
against both complete repository paths and file names at any depth.

### 4. Build

```bash
go build -o codestrike ./cmd/codestrike
```

### 5. Run a review

```bash
./codestrike review https://github.com/{owner}/{repo}/pull/{number}
```

Or point at a specific config file with `--config`:

```bash
./codestrike review --config /path/to/default.yaml https://github.com/{owner}/{repo}/pull/{number}
```

## Running Tests

```bash
go test ./... -v
```
