package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

// newReviewCmd builds the `codestrike review <pr-url>` command.
func newReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review <pr-url>",
		Short: "Run an AI review on a pull request and post the result as a comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("review: not implemented yet")
		},
	}

	return cmd
}
