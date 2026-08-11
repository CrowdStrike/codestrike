![CrowdStrike Logo (Light)](https://raw.githubusercontent.com/CrowdStrike/.github/main/assets/cs-logo-light-mode.png#gh-light-mode-only)
![CrowdStrike Logo (Dark)](https://raw.githubusercontent.com/CrowdStrike/.github/main/assets/cs-logo-dark-mode.png#gh-dark-mode-only)

# codestrike

**codestrike** is an open-source, AI-driven pull request review tool written in Go. It watches a pull request, runs an LLM-powered review over the diff, and posts the results back as a single comment. It works with any OpenAI-compatible /chat/completions endpoint, so you can point it at OpenAI, Azure OpenAI, or a self-hosted model, and lets you supply your own system prompt to tailor review behavior to your team's standards.

> [!IMPORTANT]
> **🚧 Public Preview**: This project is currently in public preview and under active development. Features and functionality may change before the stable 1.0 release. While we encourage exploration and testing, please avoid production deployments. We welcome your feedback through [GitHub Issues](https://github.com/crowdstrike/codestrike/issues) to help shape the final release.
