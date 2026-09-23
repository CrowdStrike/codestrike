package review

import (
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/llm"
	"github.com/CrowdStrike/codestrike/internal/scm"
	"github.com/CrowdStrike/codestrike/internal/tokenizer"
)

func TestIsIgnoredPath(t *testing.T) {
	pipeline := &Pipeline{}
	patterns := []string{
		"vendor/",
		"*.lock",
		"web/*.min.js",
	}

	tests := []struct {
		filename string
		want     bool
	}{
		{filename: "vendor/package/source.go", want: true},
		{filename: "nested/vendor/source.go", want: false},
		{filename: "go.sum", want: false},
		{filename: "frontend/package.lock", want: true},
		{filename: "web/app.min.js", want: true},
		{filename: "nested/web/app.min.js", want: false},
		{filename: "web/app.js", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := pipeline.isIgnoredPath(tt.filename, patterns); got != tt.want {
				t.Errorf("isIgnoredPath(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

type fakeLLMClient struct {
	response *llm.LLMResponse
}

func (f *fakeLLMClient) InvokeModel(_ context.Context, _ llm.LLMRequest) (*llm.LLMResponse, error) {
	return f.response, nil
}

func (f *fakeLLMClient) InvokeModelWithRetry(_ context.Context, _ llm.LLMRequest) (*llm.LLMResponse, error) {
	return f.response, nil
}

func newTestRunPipeline(client *mockClient, llmClient llm.LLMClient, dryRun bool) *Pipeline {
	log := zerolog.Nop()
	return NewPipeline(client, llmClient, &config.Config{}, tokenizer.New(), &log, Options{
		DryRun: dryRun,
	})
}

func TestRun_DryRun_ReturnsBodyWithoutPublishing(t *testing.T) {
	client := &mockClient{
		files: []scm.PullRequestFile{
			{Filename: "a.go", Status: "modified", Patch: "@@ -1,1 +1,1 @@\n-old\n+new"},
		},
	}
	llmClient := &fakeLLMClient{
		response: &llm.LLMResponse{Content: `[{"file":"a.go","line":1,"body":"looks fine"}]`},
	}
	p := newTestRunPipeline(client, llmClient, true)

	ref := PRReference{Provider: ProviderGitHub, Owner: "o", Repo: "r", Number: 42}
	body, err := p.Run(context.Background(), ref)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !strings.Contains(body, "looks fine") {
		t.Errorf("expected returned body to contain review content, got %q", body)
	}
	if !strings.Contains(body, "<!-- codestrike:review -->") {
		t.Errorf("expected returned body to contain codestrike marker, got %q", body)
	}
	if len(client.publishCalls) != 0 {
		t.Errorf("expected PublishComment to not be called in dry-run mode, got %d calls", len(client.publishCalls))
	}
}

func TestRun_Publishes_ReturnsEmptyBody(t *testing.T) {
	client := &mockClient{
		files: []scm.PullRequestFile{
			{Filename: "a.go", Status: "modified", Patch: "@@ -1,1 +1,1 @@\n-old\n+new"},
		},
	}
	llmClient := &fakeLLMClient{
		response: &llm.LLMResponse{Content: `[{"file":"a.go","line":1,"body":"looks fine"}]`},
	}
	p := newTestRunPipeline(client, llmClient, false)

	ref := PRReference{Provider: ProviderGitHub, Owner: "o", Repo: "r", Number: 42}
	body, err := p.Run(context.Background(), ref)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if body != "" {
		t.Errorf("expected empty returned body when not dry-run, got %q", body)
	}
	if len(client.publishCalls) != 1 {
		t.Fatalf("expected PublishComment to be called once, got %d calls", len(client.publishCalls))
	}
	if client.publishCalls[0].number != ref.Number {
		t.Errorf("expected PublishComment to be called with PR number %d, got %d", ref.Number, client.publishCalls[0].number)
	}
	if !strings.Contains(client.publishCalls[0].body, "looks fine") {
		t.Errorf("expected published body to contain review content, got %q", client.publishCalls[0].body)
	}
}
