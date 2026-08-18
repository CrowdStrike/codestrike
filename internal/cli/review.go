package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/review"
	"github.com/CrowdStrike/codestrike/internal/setup"
	"github.com/CrowdStrike/codestrike/internal/setup/logger"
	"github.com/CrowdStrike/codestrike/internal/tokenizer"
)

func newReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review <pr-url>",
		Short: "Run an AI review on a pull request and post the result as a comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = godotenv.Load()

			envCfg := setup.LoadConfig()
			log := logger.New(envCfg.LogLevel)

			configFlag, err := cmd.Flags().GetString("config")
			if err != nil {
				return fmt.Errorf("reading --config flag: %w", err)
			}

			fullContext, err := cmd.Flags().GetBool("full-context")
			if err != nil {
				return fmt.Errorf("reading --full-context flag: %w", err)
			}

			persona, err := cmd.Flags().GetString("persona")
			if err != nil {
				return fmt.Errorf("reading --persona flag: %w", err)
			}

			configPath, err := config.ResolvePath(configFlag)
			if err != nil {
				return fmt.Errorf("resolving config path: %w", err)
			}

			appConfig, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if persona != "" {
				promptPath := filepath.Join(filepath.Dir(configPath), "prompts", persona+".md")
				data, err := os.ReadFile(promptPath)
				if err != nil {
					return fmt.Errorf("persona %q not found (expected at %s): %w", persona, promptPath, err)
				}
				appConfig.Review.SystemPrompt = string(data)
			}

			ref, err := review.ParsePRURL(args[0])
			if err != nil {
				return fmt.Errorf("parsing PR URL: %w", err)
			}

			deps, err := setup.Wire(cmd.Context(), envCfg, appConfig, &log, ref.Owner, ref.Repo)
			if err != nil {
				return fmt.Errorf("wiring dependencies: %w", err)
			}

			tok := tokenizer.NewForModel(appConfig.Review.Context.TokenizerModel)
			pipeline := review.NewPipeline(deps.SCMClient, deps.LLMClient, deps.AppConfig, tok, deps.Logger, review.Options{
				FullContext: fullContext,
			})

			return pipeline.Run(cmd.Context(), ref)
		},
	}

	cmd.Flags().Bool("full-context", false, "Fetch full file content for richer reviews (slower, uses more tokens)")
	cmd.Flags().String("persona", "", "Review persona — maps to a prompt file in prompts/ (e.g., security, performance)")

	return cmd
}
