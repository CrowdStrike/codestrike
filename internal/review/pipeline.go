package review

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	"github.com/CrowdStrike/codestrike/internal/config"
	appcontext "github.com/CrowdStrike/codestrike/internal/context"
	"github.com/CrowdStrike/codestrike/internal/llm"
	"github.com/CrowdStrike/codestrike/internal/scm"
	"github.com/CrowdStrike/codestrike/internal/tokenizer"
)

type Pipeline struct {
	client      scm.Client
	llmClient   llm.LLMClient
	config      *config.Config
	tokenizer   tokenizer.Tokenizer
	fullContext bool
	logger      *zerolog.Logger
}

type Options struct {
	FullContext bool
}

func NewPipeline(client scm.Client, llmClient llm.LLMClient, cfg *config.Config, tok tokenizer.Tokenizer, logger *zerolog.Logger, opts Options) *Pipeline {
	return &Pipeline{
		client:      client,
		llmClient:   llmClient,
		config:      cfg,
		tokenizer:   tok,
		fullContext: opts.FullContext,
		logger:      logger,
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

	// Step 2: Fetch PR files
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

	// Step 3b: Fetch full file content if requested
	if p.fullContext {
		filtered = p.enrichWithContent(ctx, filtered)
	}

	// Step 4: Build context-aware prompts
	budget := p.createBudget()
	builder := appcontext.NewBuilder(p.tokenizer, budget, p.config)

	// Load project context files (CLAUDE.md, etc.)
	projectContext := p.loadContextFiles()

	// Fetch PR metadata (title, body, commits)
	prMetadata := p.fetchPRMetadata(ctx, ref.Number)

	// Fetch existing comments for dedup
	existingComments := p.fetchExistingCommentsContext(ctx, ref.Number)

	// Memory context (deferred — not yet implemented)
	memoryContext := ""

	result := builder.Build(filtered, prMetadata, existingComments, memoryContext, projectContext)
	p.logger.Info().
		Int("total_tokens", result.TotalTokens).
		Int("skipped_files", len(result.SkippedFiles)).
		Msg("context built")

	if len(result.SkippedFiles) > 0 {
		p.logger.Warn().
			Strs("files", result.SkippedFiles).
			Msg("files skipped due to context budget")
	}

	// Step 5: Run inference
	comments, err := p.runInference(ctx, result.Prompt, result.OutputMaxToken)
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

func (p *Pipeline) createBudget() *appcontext.Budget {
	ctxCfg := p.config.Review.Context
	contextWindow := 128000
	reservedOutput := 4096
	maxInputRatio := 0.75

	if ctxCfg.ReservedOutputTokens > 0 {
		reservedOutput = ctxCfg.ReservedOutputTokens
	}
	if ctxCfg.MaxInputRatio > 0 {
		maxInputRatio = ctxCfg.MaxInputRatio
	}

	return appcontext.NewBudget(p.tokenizer, contextWindow, maxInputRatio, reservedOutput)
}

func (p *Pipeline) runInference(ctx context.Context, prompt string, maxTokens int) ([]scm.ReviewComment, error) {
	resp, err := p.llmClient.InvokeModelWithRetry(ctx, llm.LLMRequest{
		Prompt:      prompt,
		MaxTokens:   maxTokens,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM invocation failed: %w", err)
	}

	return parseResponse(resp.Content)
}

func parseResponse(content string) ([]scm.ReviewComment, error) {
	// Strip chain-of-thought reasoning block
	if idx := strings.Index(content, "</reasoning>"); idx != -1 {
		content = content[idx+len("</reasoning>"):]
	}

	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Handle empty review (no issues found)
	if content == "[]" || content == "" {
		return nil, nil
	}

	var comments []scm.ReviewComment
	if err := json.Unmarshal([]byte(content), &comments); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w", err)
	}

	return comments, nil
}

func (p *Pipeline) applyGuardrails(files []scm.PullRequestFile) []scm.PullRequestFile {
	guardrails := p.config.Review.Guardrails
	var result []scm.PullRequestFile

	for _, f := range files {
		if p.isIgnoredPath(f.Filename, guardrails.IgnoredPaths) {
			p.logger.Debug().Str("file", f.Filename).Msg("skipped: ignored path")
			continue
		}
		if guardrails.MaxPatchSizeBytes > 0 && len(f.Patch) > guardrails.MaxPatchSizeBytes {
			p.logger.Debug().Str("file", f.Filename).Msg("skipped: exceeds max patch size")
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
		// Repository paths always use forward slashes. A trailing slash denotes
		// a directory prefix, while glob patterns can target either the full
		// path or a file name at any depth.
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(filename, pattern) {
			return true
		}
		if matched, _ := path.Match(pattern, filename); matched {
			return true
		}
		if matched, _ := path.Match(pattern, path.Base(filename)); matched {
			return true
		}
	}
	return false
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
	sb.WriteString("<!-- codestrike:review -->\n")
	sb.WriteString("## Code Review\n\n")

	for _, c := range comments {
		fmt.Fprintf(&sb, "**%s:%d**\n%s\n\n", c.File, c.Line, c.Body)
	}

	return sb.String()
}

func (p *Pipeline) fetchExistingCommentsContext(ctx context.Context, prNumber int) string {
	general, err := p.client.GetPRComments(ctx, prNumber)
	if err != nil {
		p.logger.Warn().Err(err).Msg("failed to fetch PR comments, continuing without")
		return ""
	}

	inline, err := p.client.GetPRReviewComments(ctx, prNumber)
	if err != nil {
		p.logger.Warn().Err(err).Msg("failed to fetch review comments, continuing without")
		return ""
	}

	all := append(general, inline...)
	if len(all) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, c := range all {
		isCodestrike := strings.Contains(c.Body, "<!-- codestrike:review -->")
		source := c.Author
		if isCodestrike {
			source = "codestrike"
		}

		if c.Path != "" && c.Line > 0 {
			fmt.Fprintf(&sb, "- [%s:%d] %q — reviewer: %s\n", c.Path, c.Line, truncate(c.Body, 120), source)
		} else {
			fmt.Fprintf(&sb, "- (general) %q — reviewer: %s\n", truncate(c.Body, 120), source)
		}
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func (p *Pipeline) enrichWithContent(ctx context.Context, files []scm.PullRequestFile) []scm.PullRequestFile {
	for i, f := range files {
		if f.Status == "added" {
			continue
		}
		content, err := p.client.GetFileContent(ctx, f.Filename, "")
		if err != nil {
			p.logger.Debug().Str("file", f.Filename).Err(err).Msg("could not fetch full content")
			continue
		}
		files[i].Content = content
	}
	p.logger.Info().Int("files_enriched", len(files)).Msg("fetched full file content")
	return files
}

func (p *Pipeline) loadContextFiles() string {
	files := p.config.Review.ContextFiles
	if len(files) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, filePath := range files {
		data, err := os.ReadFile(filePath)
		if err != nil {
			p.logger.Debug().Str("file", filePath).Msg("context file not found, skipping")
			continue
		}
		content := strings.TrimSpace(string(data))
		if content == "" {
			continue
		}
		fmt.Fprintf(&sb, "### %s\n%s\n\n", filepath.Base(filePath), content)
	}

	return sb.String()
}

func (p *Pipeline) fetchPRMetadata(ctx context.Context, prNumber int) string {
	meta, err := p.client.GetPullRequestMetadata(ctx, prNumber)
	if err != nil {
		p.logger.Warn().Err(err).Msg("failed to fetch PR metadata, continuing without")
		return ""
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "**Title:** %s\n", meta.Title)
	fmt.Fprintf(&sb, "**Branch:** %s → %s\n", meta.HeadBranch, meta.BaseBranch)

	if meta.Body != "" {
		fmt.Fprintf(&sb, "\n**Description:**\n%s\n", meta.Body)
	}

	if len(meta.CommitMessages) > 0 {
		sb.WriteString("\n**Commits:**\n")
		for _, msg := range meta.CommitMessages {
			first, _, _ := strings.Cut(msg, "\n")
			fmt.Fprintf(&sb, "- %s\n", first)
		}
	}

	return sb.String()
}
