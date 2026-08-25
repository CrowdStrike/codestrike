package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GitHub GitHubConfig `yaml:"github"`
	Review ReviewConfig `yaml:"review"`
}

type GitHubConfig struct {
	BaseURL string `yaml:"base_url"`
}

type ReviewConfig struct {
	SystemPrompt string        `yaml:"system_prompt"`
	Tone         string        `yaml:"tone"`
	ContextFiles []string      `yaml:"context_files"`
	Guardrails   Guardrails    `yaml:"guardrails"`
	Context      ContextConfig `yaml:"context"`
}

type ContextConfig struct {
	MaxInputRatio        float64               `yaml:"max_input_ratio"`
	ReservedOutputTokens int                   `yaml:"reserved_output_tokens"`
	TokenizerModel       string                `yaml:"tokenizer_model"`
	ModelLimits          map[string]ModelLimit `yaml:"model_limits"`
}

type ModelLimit struct {
	ContextWindow   int `yaml:"context_window"`
	MaxOutputTokens int `yaml:"max_output_tokens"`
}

type Guardrails struct {
	MaxPatchSizeBytes int      `yaml:"max_patch_size_bytes"`
	IgnoredPaths      []string `yaml:"ignored_paths"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("config file not found at %q; pass --config to specify a different path", path)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	configDir := filepath.Dir(path)
	cfg.Review.SystemPrompt, err = resolveText(configDir, "prompts", cfg.Review.SystemPrompt)
	if err != nil {
		return nil, fmt.Errorf("loading review system prompt: %w", err)
	}
	cfg.Review.Tone, err = resolveText(configDir, "tones", cfg.Review.Tone)
	if err != nil {
		return nil, fmt.Errorf("loading review tone: %w", err)
	}

	return &cfg, nil
}

// resolveText treats a value as a predefined file name when a matching file
// exists in the relevant config subdirectory. Otherwise it remains inline text.
func resolveText(configDir, subdir, value string) (string, error) {
	for _, name := range []string{value, value + ".md"} {
		if name == "" || filepath.Base(name) != name {
			continue
		}
		path := filepath.Join(configDir, subdir, name)
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("reading %q: %w", path, err)
		}
	}
	return value, nil
}
