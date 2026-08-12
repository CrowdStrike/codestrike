package config

import (
	"fmt"
	"os"

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
	SystemPrompt string     `yaml:"system_prompt"`
	Tone         string     `yaml:"tone"`
	Guardrails   Guardrails `yaml:"guardrails"`
}

type Guardrails struct {
	MaxFileSize  int      `yaml:"max_file_size"`
	IgnoredPaths []string `yaml:"ignored_paths"`
	IgnoredFiles []string `yaml:"ignored_files"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}
