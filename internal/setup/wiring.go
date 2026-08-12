package setup

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/env"
	"github.com/CrowdStrike/codestrike/internal/llm"
	"github.com/CrowdStrike/codestrike/internal/llm/aws"
	"github.com/CrowdStrike/codestrike/internal/llm/ollama"
	"github.com/CrowdStrike/codestrike/internal/llm/openaiplatform"
	"github.com/CrowdStrike/codestrike/internal/scm"
	"github.com/CrowdStrike/codestrike/internal/scm/github"
)

type Config struct {
	GitHubToken string
	OpenAIURL   string
	OpenAIKey   string
	AWSRegion   string
	OllamaURL   string
	ModelFamily string
	ModelID     string
	LogLevel    string
}

type Dependencies struct {
	SCMClient scm.Client
	LLMClient llm.LLMClient
	AppConfig *config.Config
	Logger    *zerolog.Logger
}

func LoadConfig() *Config {
	return &Config{
		GitHubToken: env.GetString("GITHUB_TOKEN", ""),
		OpenAIURL:   env.GetString("OPEN_AI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIKey:   env.GetString("OPEN_AI_KEY", ""),
		AWSRegion:   env.GetString("AWS_REGION", "us-east-1"),
		OllamaURL:   env.GetString("OLLAMA_BASE_URL", "http://localhost:11434/v1"),
		ModelFamily: env.GetString("MODEL_FAMILY", "openai_platform"),
		ModelID:     env.GetString("MODEL_ID", "gpt-4o"),
		LogLevel:    env.GetString("LOG_LEVEL", "info"),
	}
}

func Wire(ctx context.Context, cfg *Config, appConfig *config.Config, logger *zerolog.Logger, owner, repo string) (*Dependencies, error) {
	if cfg.GitHubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	ghClient := github.New(github.Config{
		Owner:   owner,
		Repo:    repo,
		Token:   cfg.GitHubToken,
		BaseURL: appConfig.GitHub.BaseURL,
	})

	llmClient, err := createLLMClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating LLM client: %w", err)
	}

	return &Dependencies{
		SCMClient: ghClient,
		LLMClient: llmClient,
		AppConfig: appConfig,
		Logger:    logger,
	}, nil
}

func createLLMClient(ctx context.Context, cfg *Config) (llm.LLMClient, error) {
	family, err := llm.ParseFamily(cfg.ModelFamily)
	if err != nil {
		return nil, err
	}

	switch family {
	case llm.FamilyAnthropic:
		if cfg.AWSRegion == "" {
			return nil, fmt.Errorf("AWS_REGION is required for Anthropic/Bedrock provider")
		}
		return aws.NewClient(ctx, cfg.AWSRegion, cfg.ModelID)
	case llm.FamilyOpenAIPlatform:
		if cfg.OpenAIKey == "" {
			return nil, fmt.Errorf("OPEN_AI_KEY is required for OpenAI provider")
		}
		return openaiplatform.NewClient(ctx, cfg.OpenAIURL, cfg.OpenAIKey, cfg.ModelID)
	case llm.FamilyOllama:
		if cfg.OllamaURL == "" {
			return nil, fmt.Errorf("OLLAMA_BASE_URL is required for Ollama provider")
		}
		return ollama.NewClient(ctx, cfg.OllamaURL, cfg.ModelID)
	default:
		return nil, fmt.Errorf("unsupported model family: %s", family)
	}
}
