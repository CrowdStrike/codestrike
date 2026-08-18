package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/CrowdStrike/codestrike/internal/version"
)

// newVersionCmd builds the `codestrike version` command.
func newVersionCmd() *cobra.Command {
	var long bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(c *cobra.Command, _ []string) {
			out := c.OutOrStdout()
			if long {
				fmt.Fprintln(out, "Version: "+version.Version)
				fmt.Fprintln(out)
				fmt.Fprintln(out, "Commit:      "+version.Commit)
				fmt.Fprintln(out, "Commit Date: "+version.CommitDate)
				fmt.Fprintln(out, "Built By:    "+version.BuiltBy)
				fmt.Fprintln(out, "Build Date:  "+version.Date)
				return
			}
			fmt.Fprintln(out, c.Root().Name()+" "+version.Short())
		},
	}

	cmd.Flags().BoolVar(&long, "long", false, "Print full build information")

	return cmd
}
