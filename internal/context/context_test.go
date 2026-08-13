package context_test

import (
	"strings"
	"testing"

	appcontext "github.com/CrowdStrike/codestrike/internal/context"
	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/scm"
	"github.com/CrowdStrike/codestrike/internal/tokenizer"
)

func TestBudget_AvailableInputTokens(t *testing.T) {
	tok := tokenizer.New()
	budget := appcontext.NewBudget(tok, 128000, 0.75, 4096)

	available := budget.AvailableInputTokens()
	want := 128000*75/100 - 4096
	if available != want {
		t.Errorf("AvailableInputTokens() = %d, want %d", available, want)
	}
}

func TestBudget_Allocate(t *testing.T) {
	tok := tokenizer.New()
	budget := appcontext.NewBudget(tok, 100000, 0.75, 4096)
	available := budget.AvailableInputTokens()

	alloc := budget.Allocate(1000, 5000, 3000, 50000)

	if alloc.SystemPrompt != 1000 {
		t.Errorf("SystemPrompt = %d, want 1000", alloc.SystemPrompt)
	}

	if alloc.ExistingComments != 5000 {
		t.Errorf("ExistingComments = %d, want 5000 (under 15%% cap)", alloc.ExistingComments)
	}

	if alloc.Memory != 3000 {
		t.Errorf("Memory = %d, want 3000 (under 10%% cap)", alloc.Memory)
	}

	remaining := available - 1000 - 5000 - 3000
	if alloc.FilePatches > remaining {
		t.Errorf("FilePatches = %d, exceeds remaining %d", alloc.FilePatches, remaining)
	}
}

func TestBuilder_SmallPR_FitsInBudget(t *testing.T) {
	tok := tokenizer.New()
	budget := appcontext.NewBudget(tok, 128000, 0.75, 4096)
	cfg := &config.Config{
		Review: config.ReviewConfig{
			SystemPrompt: "Review code.",
			Tone:         "constructive",
		},
	}

	builder := appcontext.NewBuilder(tok, budget, cfg)
	files := []scm.PullRequestFile{
		{Filename: "a.go", Status: "modified", Patch: "small patch"},
		{Filename: "b.go", Status: "added", Patch: "another small patch"},
	}

	result := builder.Build(files, "", "", "")
	if len(result.SkippedFiles) != 0 {
		t.Errorf("expected 0 skipped files, got %d", len(result.SkippedFiles))
	}
	if result.Prompt == "" {
		t.Error("expected non-empty prompt")
	}
	if result.TotalTokens == 0 {
		t.Error("expected non-zero total tokens")
	}
}

func TestBuilder_LargePR_SkipsFiles(t *testing.T) {
	tok := tokenizer.New()
	budget := appcontext.NewBudget(tok, 1000, 0.75, 200) // very small budget
	cfg := &config.Config{
		Review: config.ReviewConfig{
			SystemPrompt: "Review.",
			Tone:         "strict",
		},
	}

	builder := appcontext.NewBuilder(tok, budget, cfg)
	files := []scm.PullRequestFile{
		{Filename: "small.go", Status: "modified", Patch: "x := 1"},
		{Filename: "big.go", Status: "modified", Patch: strings.Repeat("word ", 500)},
	}

	result := builder.Build(files, "", "", "")
	if len(result.SkippedFiles) == 0 {
		t.Error("expected some files to be skipped due to tight budget")
	}
}

func TestBuilder_IncludesExistingComments(t *testing.T) {
	tok := tokenizer.New()
	budget := appcontext.NewBudget(tok, 128000, 0.75, 4096)
	cfg := &config.Config{
		Review: config.ReviewConfig{
			SystemPrompt: "Review.",
			Tone:         "constructive",
		},
	}

	builder := appcontext.NewBuilder(tok, budget, cfg)
	files := []scm.PullRequestFile{
		{Filename: "a.go", Status: "modified", Patch: "change"},
	}

	comments := "- [a.go:10] \"missing error check\" — reviewer: codestrike"
	result := builder.Build(files, comments, "", "")

	if !strings.Contains(result.Prompt, "do NOT repeat") {
		t.Error("expected prompt to contain dedup instructions")
	}
	if !strings.Contains(result.Prompt, "missing error check") {
		t.Error("expected prompt to contain existing comment")
	}
}
