package cli

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/review"
	"github.com/CrowdStrike/codestrike/internal/setup"
	"github.com/CrowdStrike/codestrike/internal/setup/logger"
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

			configPath, err := config.ResolvePath(configFlag)
			if err != nil {
				return fmt.Errorf("resolving config path: %w", err)
			}

			appConfig, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			ref, err := review.ParsePRURL(args[0])
			if err != nil {
				return fmt.Errorf("parsing PR URL: %w", err)
			}

			deps, err := setup.Wire(cmd.Context(), envCfg, appConfig, &log, ref.Owner, ref.Repo)
			if err != nil {
				return fmt.Errorf("wiring dependencies: %w", err)
			}

			pipeline := review.NewPipeline(deps.SCMClient, deps.LLMClient, deps.AppConfig, deps.Logger)

			return pipeline.Run(cmd.Context(), ref)
		},
	}

	return cmd
}
