package config_test

import (
	"os"
	"path/filepath"
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
    max_file_size: 512000
    ignored_paths:
      - vendor/
      - dist/
    ignored_files:
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

	if cfg.Review.Guardrails.MaxFileSize != 512000 {
		t.Errorf("Guardrails.MaxFileSize = %d, want %d", cfg.Review.Guardrails.MaxFileSize, 512000)
	}

	if len(cfg.Review.Guardrails.IgnoredPaths) != 2 {
		t.Errorf("Guardrails.IgnoredPaths len = %d, want 2", len(cfg.Review.Guardrails.IgnoredPaths))
	}

	if len(cfg.Review.Guardrails.IgnoredFiles) != 1 {
		t.Errorf("Guardrails.IgnoredFiles len = %d, want 1", len(cfg.Review.Guardrails.IgnoredFiles))
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
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
