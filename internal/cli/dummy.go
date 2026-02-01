package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDummyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hello",
		Short: "Print a sample greeting",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Hello from Carnie.")
			return nil
		},
	}

	return cmd
}
