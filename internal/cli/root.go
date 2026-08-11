package cli

import (
	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags.
var version = "dev"

// NewRootCmd constructs the root codestrike command with all subcommands attached.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "codestrike",
		Short:         "codestrike is an AI-driven pull request review tool",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	root.AddCommand(newReviewCmd())

	return root
}

// Execute runs the codestrike CLI and returns any error encountered.
func Execute() error {
	return NewRootCmd().Execute()
}
