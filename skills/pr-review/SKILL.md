---
name: pr-review
description: Runs codestrike's AI code review on a GitHub pull request via the codestrike CLI. Use when the user asks to review a PR or get a second opinion on a pull request.
---

# PR Review (codestrike)

Run the `codestrike` CLI to review a GitHub pull request and post the result as a PR comment.

## When to use

- The user pastes a GitHub PR URL and asks for a review.
- The user asks to "review this PR" while a PR URL is visible in context.

## How to invoke

Run in the terminal:

    npx codestrike review <pr-url>

Optional flags:
- `--persona <name>` — pick a review persona (e.g. `security`, `critical-strike`). Check `prompts/` next to the codestrike config for available names if unsure, or ask the user.
- `--full-context` — fetch full file content instead of just the diff, for a deeper (slower, more token-hungry) review. Only use when asked or when the diff alone seems insufficient.
- `--dry-run` — print the review to stdout instead of posting it as a PR comment. Use this when the user wants to preview a review, or explicitly asks you not to post anything to the PR.

codestrike requires `GITHUB_TOKEN` and LLM provider credentials (`MODEL_FAMILY` + matching key/region) to already be configured in the environment (see the project's `.env` / README setup). If the command fails with a missing-credential error, tell the user what's missing rather than guessing values.

## Interpreting output

By default, codestrike posts the review directly as a GitHub PR comment tagged `<!-- codestrike:review -->`. Report to the user what the command printed (files reviewed, comments posted, or "no actionable comments" if none were found). If the command errors, surface the error message rather than retrying blindly.

With `--dry-run`, codestrike prints the formatted review body to stdout instead of posting anything — show that printed content to the user directly, and make clear that no comment was posted to the PR.
