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

func (b *Builder) Build(files []scm.PullRequestFile, prMetadata, existingComments, memoryContext, projectContext string) BuildResult {
	systemSection := b.buildSystemSection()
	systemTokens := b.tokenizer.CountTokens(systemSection)
	projectTokens := b.tokenizer.CountTokens(projectContext)
	metadataTokens := b.tokenizer.CountTokens(prMetadata)
	commentsTokens := b.tokenizer.CountTokens(existingComments)
	memoryTokens := b.tokenizer.CountTokens(memoryContext)

	fixedTokens := systemTokens + projectTokens
	alloc := b.budget.Allocate(fixedTokens, metadataTokens, commentsTokens, memoryTokens, 0)
	patchBudget := b.budget.AvailableInputTokens() - alloc.SystemPrompt - alloc.PRMetadata - alloc.ExistingComments - alloc.Memory

	// Fit as many files as possible within budget
	var included []scm.PullRequestFile
	var skipped []string
	usedTokens := 0

	for _, f := range files {
		fileTokens := b.tokenizer.CountTokens(f.Patch)
		if f.Content != "" {
			fileTokens += b.tokenizer.CountTokens(f.Content)
		}
		if usedTokens+fileTokens > patchBudget {
			skipped = append(skipped, f.Filename)
			continue
		}
		included = append(included, f)
		usedTokens += fileTokens
	}

	prompt := b.assemblePrompt(systemSection, projectContext, prMetadata, existingComments, memoryContext, included, skipped)
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
	sb.WriteString("## Instructions\n\n")
	sb.WriteString("Think through each file's changes step by step. For each file, reason about:\n")
	sb.WriteString("- What the change does and why it might be wrong\n")
	sb.WriteString("- Edge cases, error handling gaps, security implications\n")
	sb.WriteString("- Whether the change is consistent with the surrounding code\n\n")
	sb.WriteString("After your analysis, output your findings as a JSON array.\n")
	sb.WriteString("Each item must have: {\"file\": \"<path>\", \"line\": <number>, \"body\": \"<comment>\"}\n\n")
	sb.WriteString("Format your response as:\n")
	sb.WriteString("<reasoning>\n...your step-by-step analysis here...\n</reasoning>\n\n")
	sb.WriteString("```json\n[...your comments here...]\n```\n")
	return sb.String()
}

func (b *Builder) assemblePrompt(system, projectContext, prMetadata, comments, memory string, files []scm.PullRequestFile, skipped []string) string {
	var sb strings.Builder
	sb.WriteString(system)
	sb.WriteString("\n")

	if projectContext != "" {
		sb.WriteString("\n## Project Conventions\n")
		sb.WriteString(projectContext)
		sb.WriteString("\n")
	}

	if prMetadata != "" {
		sb.WriteString("\n## Author's Intent\n")
		sb.WriteString(prMetadata)
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
		if f.Content != "" {
			sb.WriteString("Full file content:\n")
			sb.WriteString(f.Content)
			sb.WriteString("\n\nDiff:\n")
		}
		sb.WriteString(f.Patch)
		sb.WriteString("\n\n")
	}

	return sb.String()
}
