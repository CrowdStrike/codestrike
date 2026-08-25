![CrowdStrike Logo (Light)](https://raw.githubusercontent.com/CrowdStrike/.github/main/assets/cs-logo-light-mode.png#gh-light-mode-only)
![CrowdStrike Logo (Dark)](https://raw.githubusercontent.com/CrowdStrike/.github/main/assets/cs-logo-dark-mode.png#gh-dark-mode-only)

# codestrike

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

You can customize the system prompt, tone, guardrails, context files, and token
budget:

```yaml
github:
  base_url: https://api.github.com

review:
  # Use inline text, or a file name from prompts/ or tones/ next to this file.
  system_prompt: default
  tone: diplomatic
  context_files:
    - CLAUDE.md
  guardrails:
    max_patch_size_bytes: 1048576
    ignored_paths:
      - vendor/
      - node_modules/
      - .git/
      - "*.lock"
      - "*.sum"
      - "*.min.js"
      - "*.min.css"
  context:
    max_input_ratio: 0.75
    reserved_output_tokens: 4096
    tokenizer_model: o200k_base
    model_limits:
      gpt-4o:
        context_window: 128000
        max_output_tokens: 16384
      claude-sonnet-4-20250514:
        context_window: 200000
        max_output_tokens: 8192
```

| Section | Description |
|---------|-------------|
| `context_files` | Project files (e.g., `CLAUDE.md`) loaded as additional context for the reviewer |
| `ignored_paths` | Directory prefixes and glob patterns. Globs are matched against both complete repository paths and file names at any depth |
| `context.max_input_ratio` | Maximum fraction of the model's context window used for input (default: `0.75`) |
| `context.reserved_output_tokens` | Tokens reserved for the model's response (default: `4096`) |
| `context.tokenizer_model` | Tiktoken encoding used for token counting (default: `o200k_base`) |
| `context.model_limits` | Per-model context window and max output token settings |

### 4. Build

```bash
make build
```

### 5. Run a review

```bash
./codestrike review https://github.com/{owner}/{repo}/pull/{number}
```

Or point at a specific config file with `--config`:

```bash
./codestrike review --config /path/to/default.yaml https://github.com/{owner}/{repo}/pull/{number}
```

#### Flags

| Flag | Description |
|------|-------------|
| `--config <path>` | Path to a YAML config file |
| `--full-context` | Fetch full file content for richer reviews (slower, uses more tokens) |
| `--persona <name>` | Select a review persona — maps to a prompt file in `prompts/` (e.g., `security`, `performance`) |

Example with persona and full context:

```bash
./codestrike review --persona security --full-context https://github.com/{owner}/{repo}/pull/{number}
```

#### Chain-of-thought reasoning

codestrike uses chain-of-thought prompting: the LLM is asked to reason
step-by-step about each file's changes before producing review comments. The
reasoning block is stripped from the final output automatically.

## Development

Common tasks are wrapped in the `Makefile`; run `make help` to list all targets.

| Target | Description |
|--------|--------------|
| `make build` | Build the `codestrike` binary for the host platform |
| `make run` | Run codestrike from source |
| `make test` | Run unit tests with the race detector and coverage |
| `make fmt` | Run `go fmt` against the code |
| `make vet` | Run `go vet` against the code |
| `make lint` | Run `golangci-lint` (downloaded locally if needed) |
| `make lint-fix` | Run `golangci-lint` and apply automatic fixes |
| `make snapshot` | Build unpublished per-platform binaries into `dist/` via GoReleaser |
| `make clean` | Remove build, packaging, and tool artifacts |

## Continuous Integration

Pull requests and pushes to `main` run the `ci` workflow (`.github/workflows/ci.yml`), which verifies modules are tidy, checks formatting, and runs `make vet`, `make test`, and `make build`, plus a lint step via `golangci-lint-action`.

Tagged pushes (`v*`) run the `release` workflow (`.github/workflows/release.yml`), which re-runs `ci`, then builds and publishes release artifacts with GoReleaser and generates a changelog.

Changes to `.goreleaser.yaml` are validated by the `goreleaser-check` workflow (`.github/workflows/goreleaser-check.yml`), which checks the GoReleaser config and builds a snapshot release.
