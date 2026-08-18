package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/CrowdStrike/codestrike/internal/config"
)

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Install the bundled configuration in the user config directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := config.Install(force)
			if err != nil {
				return fmt.Errorf("installing config: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Configuration installed in %s\n", dir)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing bundled configuration files")
	return cmd
}
