package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CrowdStrike/codestrike/internal/config"
)

func TestLoad(t *testing.T) {
	yaml := `
github:
  base_url: https://github.example.com/api/v3

review:
  system_prompt: |
    You are a code reviewer.
    Be concise.
  tone: strict
  guardrails:
    max_patch_size_bytes: 512000
    ignored_paths:
      - vendor/
      - dist/
      - "*.lock"
`
	path := filepath.Join(t.TempDir(), "test.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.GitHub.BaseURL != "https://github.example.com/api/v3" {
		t.Errorf("GitHub.BaseURL = %q, want %q", cfg.GitHub.BaseURL, "https://github.example.com/api/v3")
	}

	if cfg.Review.Tone != "strict" {
		t.Errorf("Review.Tone = %q, want %q", cfg.Review.Tone, "strict")
	}

	wantPrompt := "You are a code reviewer.\nBe concise.\n"
	if cfg.Review.SystemPrompt != wantPrompt {
		t.Errorf("Review.SystemPrompt = %q, want %q", cfg.Review.SystemPrompt, wantPrompt)
	}

	if cfg.Review.Guardrails.MaxPatchSizeBytes != 512000 {
		t.Errorf("Guardrails.MaxPatchSizeBytes = %d, want %d", cfg.Review.Guardrails.MaxPatchSizeBytes, 512000)
	}

	if len(cfg.Review.Guardrails.IgnoredPaths) != 3 {
		t.Errorf("Guardrails.IgnoredPaths len = %d, want 3", len(cfg.Review.Guardrails.IgnoredPaths))
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "--config") {
		t.Errorf("error = %q, want it to mention --config", err.Error())
	}
	if !strings.Contains(err.Error(), "/nonexistent/path.yaml") {
		t.Errorf("error = %q, want it to mention the path", err.Error())
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(path, []byte(":\n  :\n  - [invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoad_ResolvesPredefinedPromptAndTone(t *testing.T) {
	dir := t.TempDir()
	for _, subdir := range []string{"prompts", "tones"} {
		if err := os.Mkdir(filepath.Join(dir, subdir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "prompts", "review.md"), []byte("file prompt\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tones", "concise.md"), []byte("file tone\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "default.yaml")
	if err := os.WriteFile(path, []byte("review:\n  system_prompt: review\n  tone: concise.md\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Review.SystemPrompt != "file prompt\n" {
		t.Errorf("SystemPrompt = %q", cfg.Review.SystemPrompt)
	}
	if cfg.Review.Tone != "file tone\n" {
		t.Errorf("Tone = %q", cfg.Review.Tone)
	}
}
