package context

import (
	"fmt"
	"strings"

	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/scm"
	"github.com/CrowdStrike/codestrike/internal/tokenizer"
)

type Builder struct {
	tokenizer tokenizer.Tokenizer
	budget    *Budget
	config    *config.Config
}

type BuildResult struct {
	Prompt         string
	TotalTokens    int
	SkippedFiles   []string
	OutputMaxToken int
}

func NewBuilder(tok tokenizer.Tokenizer, budget *Budget, cfg *config.Config) *Builder {
	return &Builder{
		tokenizer: tok,
		budget:    budget,
		config:    cfg,
	}
}

func (b *Builder) Build(files []scm.PullRequestFile, existingComments, memoryContext, projectContext string) BuildResult {
	systemSection := b.buildSystemSection()
	systemTokens := b.tokenizer.CountTokens(systemSection)
	projectTokens := b.tokenizer.CountTokens(projectContext)
	commentsTokens := b.tokenizer.CountTokens(existingComments)
	memoryTokens := b.tokenizer.CountTokens(memoryContext)

	fixedTokens := systemTokens + projectTokens
	alloc := b.budget.Allocate(fixedTokens, commentsTokens, memoryTokens, 0)
	patchBudget := b.budget.AvailableInputTokens() - alloc.SystemPrompt - alloc.ExistingComments - alloc.Memory

	// Fit as many files as possible within budget
	var included []scm.PullRequestFile
	var skipped []string
	usedTokens := 0

	for _, f := range files {
		fileTokens := b.tokenizer.CountTokens(f.Patch)
		if usedTokens+fileTokens > patchBudget {
			skipped = append(skipped, f.Filename)
			continue
		}
		included = append(included, f)
		usedTokens += fileTokens
	}

	prompt := b.assemblePrompt(systemSection, projectContext, existingComments, memoryContext, included, skipped)
	totalTokens := b.tokenizer.CountTokens(prompt)

	return BuildResult{
		Prompt:         prompt,
		TotalTokens:    totalTokens,
		SkippedFiles:   skipped,
		OutputMaxToken: b.budget.reservedOutputTokens,
	}
}

func (b *Builder) buildSystemSection() string {
	var sb strings.Builder
	sb.WriteString(b.config.Review.SystemPrompt)
	sb.WriteString("\n\n")
	fmt.Fprintf(&sb, "Tone: %s\n\n", b.config.Review.Tone)
	sb.WriteString("Review the following changes and provide feedback as JSON array.\n")
	sb.WriteString("Each item must have: {\"file\": \"<path>\", \"line\": <number>, \"body\": \"<comment>\"}\n")
	sb.WriteString("Respond ONLY with the JSON array, no markdown fences or other text.\n")
	return sb.String()
}

func (b *Builder) assemblePrompt(system, projectContext, comments, memory string, files []scm.PullRequestFile, skipped []string) string {
	var sb strings.Builder
	sb.WriteString(system)
	sb.WriteString("\n")

	if projectContext != "" {
		sb.WriteString("\n## Project Conventions\n")
		sb.WriteString(projectContext)
		sb.WriteString("\n")
	}

	if memory != "" {
		sb.WriteString("\n## Previous Review Context\n")
		sb.WriteString(memory)
		sb.WriteString("\n")
	}

	if comments != "" {
		sb.WriteString("\n## Existing Comments (do NOT repeat these)\n")
		sb.WriteString(comments)
		sb.WriteString("\n")
	}

	if len(skipped) > 0 {
		sb.WriteString("\n## Note\n")
		fmt.Fprintf(&sb, "The following %d files were excluded due to context budget constraints:\n", len(skipped))
		for _, f := range skipped {
			sb.WriteString("- ")
			sb.WriteString(f)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n## Changed Files\n\n")
	for _, f := range files {
		fmt.Fprintf(&sb, "--- File: %s (status: %s) ---\n", f.Filename, f.Status)
		sb.WriteString(f.Patch)
		sb.WriteString("\n\n")
	}

	return sb.String()
}
