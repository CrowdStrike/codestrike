![CrowdStrike Logo (Light)](https://raw.githubusercontent.com/CrowdStrike/.github/main/assets/cs-logo-light-mode.png#gh-light-mode-only)
![CrowdStrike Logo (Dark)](https://raw.githubusercontent.com/CrowdStrike/.github/main/assets/cs-logo-dark-mode.png#gh-dark-mode-only)

# codestrike

AI-driven pull request review tool using configurable LLM judges.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)

**codestrike** is an open-source, AI-driven pull request review tool written in Go. It watches a pull request, runs an LLM-powered review over the diff, and posts the results back as a single comment. It works with any OpenAI-compatible /chat/completions endpoint, so you can point it at OpenAI, Azure OpenAI, or a self-hosted model, and lets you supply your own system prompt to tailor review behavior to your team's standards.

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

Application settings live in `configs/codestrike.yaml`. You can customize the system prompt, tone, and guardrails:

```yaml
github:
  base_url: https://api.github.com

review:
  system_prompt: |
    You are a senior software engineer performing a code review.
    ...
  tone: constructive
  guardrails:
    max_file_size: 1048576
    ignored_paths:
      - vendor/
      - node_modules/
    ignored_files:
      - "*.lock"
      - "*.min.js"
```

### 4. Build

```bash
go build -o codestrike ./cmd/codestrike
```

### 5. Run a review

```bash
./codestrike review https://github.com/{owner}/{repo}/pull/{number}
```

## Running Tests

```bash
go test ./... -v
```
