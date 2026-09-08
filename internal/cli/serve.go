package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CrowdStrike/codestrike/internal/api"
	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/setup"
	"github.com/CrowdStrike/codestrike/internal/setup/logger"
	"github.com/emicklei/go-restful/v3"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run CodeStrike as a persistent HTTP service",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = godotenv.Load()

			envCfg := setup.LoadConfig()
			log := logger.New(envCfg.LogLevel)

			appConfig, err := getAppConfig(cmd)
			if err != nil {
				return err
			}

			if envCfg.GitHubToken == "" {
				return fmt.Errorf("GITHUB_TOKEN environment variable is required")
			}

			llmClient, err := setup.CreateLLMClient(cmd.Context(), envCfg)
			if err != nil {
				return fmt.Errorf("creating LLM client: %w", err)
			}

			addr, _ := cmd.Flags().GetString("addr")
			if addr == "" {
				addr = ":" + envCfg.ListenHttpPort
			}

			handler := api.NewHandler(llmClient, appConfig, envCfg.GitHubToken, &log)
			container := restful.NewContainer()
			api.RegisterRoutes(container, handler)

			srv := &http.Server{
				Addr:              addr,
				Handler:           container,
				ReadHeaderTimeout: 10 * time.Second,
			}

			return listenAndShutdown(cmd.Context(), srv, &log)
		},
	}

	cmd.Flags().String("addr", "", "Bind address (default :8080, or PORT env var)")

	return cmd
}

func getAppConfig(cmd *cobra.Command) (*config.Config, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, fmt.Errorf("reading --config flag: %w", err)
	}

	resolvedPath, err := config.ResolvePath(configPath)
	if err != nil {
		return nil, fmt.Errorf("resolving config path: %w", err)
	}

	appConfig, err := config.Load(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return appConfig, nil
}

func listenAndShutdown(ctx context.Context, srv *http.Server, logger *zerolog.Logger) error {
	errCh := make(chan error, 1)

	go func() {
		logger.Info().Str("addr", srv.Addr).Msg("server listening")
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-sigCh:
		logger.Info().Str("signal", sig.String()).Msg("shutting down")
	case <-ctx.Done():
		logger.Info().Msg("context cancelled, shutting down")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
