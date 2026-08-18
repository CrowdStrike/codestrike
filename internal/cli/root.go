package cli

import (
	"github.com/spf13/cobra"

	"github.com/CrowdStrike/codestrike/internal/version"
)

// NewRootCmd constructs the root codestrike command with all subcommands attached.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "codestrike",
		Short:         "codestrike is an AI-driven pull request review tool",
		Version:       version.Version,
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.SetVersionTemplate(root.Name() + " " + version.Short() + "\n")

	root.PersistentFlags().String("config", "", "Path to app config file (default: OS user config dir, e.g. ~/.config/codestrike/default.yaml on Linux)")

	root.AddCommand(newReviewCmd())
	root.AddCommand(newInitCmd())
	root.AddCommand(newVersionCmd())

	return root
}

// Execute runs the codestrike CLI and returns any error encountered.
func Execute() error {
	return NewRootCmd().Execute()
}
