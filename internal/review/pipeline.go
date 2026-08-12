package review

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/llm"
	"github.com/CrowdStrike/codestrike/internal/scm"
)

type Pipeline struct {
	client    scm.Client
	llmClient llm.LLMClient
	config    *config.Config
	logger    *zerolog.Logger
}

func NewPipeline(client scm.Client, llmClient llm.LLMClient, cfg *config.Config, logger *zerolog.Logger) *Pipeline {
	return &Pipeline{
		client:    client,
		llmClient: llmClient,
		config:    cfg,
		logger:    logger,
	}
}

func (p *Pipeline) Run(ctx context.Context, ref PRReference) error {
	// Step 1: Verify PR exists
	exists, err := p.client.PullRequestExists(ctx, ref.Number)
	if err != nil {
		return fmt.Errorf("checking PR existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("pull request #%d not found in %s/%s", ref.Number, ref.Owner, ref.Repo)
	}
	p.logger.Info().Int("pr", ref.Number).Msg("pull request found")

	// Step 2: Fetch PR files (diff + content)
	files, err := p.client.GetPullRequestFiles(ctx, ref.Number)
	if err != nil {
		return fmt.Errorf("fetching PR files: %w", err)
	}
	p.logger.Info().Int("total_files", len(files)).Msg("fetched PR files")

	// Step 3: Apply guardrails
	filtered := p.applyGuardrails(files)
	p.logger.Info().Int("filtered_files", len(filtered)).Msg("files after guardrails")

	if len(filtered) == 0 {
		p.logger.Warn().Msg("no files to review after applying guardrails")
		return nil
	}

	// Step 4: Build prompt
	prompt := p.buildPrompt(filtered)
	p.logger.Debug().Int("prompt_length", len(prompt)).Msg("prompt built")

	// Step 5: Inference
	comments, err := p.runInference(ctx, prompt)
	if err != nil {
		return fmt.Errorf("running inference: %w", err)
	}

	// Step 6: Validate output
	validated := p.validateComments(comments, filtered)
	p.logger.Info().Int("valid_comments", len(validated)).Msg("comments validated")

	if len(validated) == 0 {
		p.logger.Info().Msg("no actionable comments to post")
		return nil
	}

	// Step 7: Post comments
	body := formatComments(validated)
	if err := p.client.PublishComment(ctx, ref.Number, body); err != nil {
		return fmt.Errorf("publishing comment: %w", err)
	}
	p.logger.Info().Int("pr", ref.Number).Msg("review posted")

	return nil
}

func (p *Pipeline) applyGuardrails(files []scm.PullRequestFile) []scm.PullRequestFile {
	guardrails := p.config.Review.Guardrails
	var result []scm.PullRequestFile

	for _, f := range files {
		if p.isIgnoredPath(f.Filename, guardrails.IgnoredPaths) {
			p.logger.Debug().Str("file", f.Filename).Msg("skipped: ignored path")
			continue
		}
		if p.isIgnoredFile(f.Filename, guardrails.IgnoredFiles) {
			p.logger.Debug().Str("file", f.Filename).Msg("skipped: ignored file pattern")
			continue
		}
		if guardrails.MaxFileSize > 0 && len(f.Patch) > guardrails.MaxFileSize {
			p.logger.Debug().Str("file", f.Filename).Msg("skipped: exceeds max file size")
			continue
		}
		if f.Status == "removed" {
			p.logger.Debug().Str("file", f.Filename).Msg("skipped: file removed")
			continue
		}
		result = append(result, f)
	}

	return result
}

func (p *Pipeline) isIgnoredPath(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasPrefix(filename, pattern) {
			return true
		}
	}
	return false
}

func (p *Pipeline) isIgnoredFile(filename string, patterns []string) bool {
	base := filepath.Base(filename)
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}
	return false
}

func (p *Pipeline) buildPrompt(files []scm.PullRequestFile) string {
	var sb strings.Builder

	sb.WriteString(p.config.Review.SystemPrompt)
	sb.WriteString("\n\n")
	fmt.Fprintf(&sb, "Tone: %s\n\n", p.config.Review.Tone)
	sb.WriteString("Review the following changes and provide feedback as JSON array.\n")
	sb.WriteString("Each item must have: {\"file\": \"<path>\", \"line\": <number>, \"body\": \"<comment>\"}\n\n")

	for _, f := range files {
		fmt.Fprintf(&sb, "--- File: %s (status: %s) ---\n", f.Filename, f.Status)
		sb.WriteString(f.Patch)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func (p *Pipeline) runInference(ctx context.Context, prompt string) ([]scm.ReviewComment, error) {
	resp, err := p.llmClient.InvokeModelWithRetry(ctx, llm.LLMRequest{
		Prompt:      prompt,
		MaxTokens:   4096,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM invocation failed: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var comments []scm.ReviewComment
	if err := json.Unmarshal([]byte(content), &comments); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w", err)
	}

	return comments, nil
}

func (p *Pipeline) validateComments(comments []scm.ReviewComment, files []scm.PullRequestFile) []scm.ReviewComment {
	fileSet := make(map[string]bool, len(files))
	for _, f := range files {
		fileSet[f.Filename] = true
	}

	var valid []scm.ReviewComment
	for _, c := range comments {
		if !fileSet[c.File] {
			p.logger.Debug().Str("file", c.File).Msg("comment dropped: file not in diff")
			continue
		}
		if c.Line <= 0 {
			p.logger.Debug().Str("file", c.File).Msg("comment dropped: invalid line number")
			continue
		}
		if strings.TrimSpace(c.Body) == "" {
			continue
		}
		valid = append(valid, c)
	}

	return valid
}

func formatComments(comments []scm.ReviewComment) string {
	var sb strings.Builder
	sb.WriteString("## Code Review\n\n")

	for _, c := range comments {
		fmt.Fprintf(&sb, "**%s:%d**\n%s\n\n", c.File, c.Line, c.Body)
	}

	return sb.String()
}
