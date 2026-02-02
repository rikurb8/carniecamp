package cli

import (
	"fmt"
	"os"

	"github.com/rikurb8/carnie/internal/operator"
	"github.com/spf13/cobra"
)

func newOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "operator",
		Aliases: []string{"op"},
		Short:   "Print the operator command",
		Long:    "Outputs a ready-to-paste command to start an operator planning session.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			planning, err := operator.BuildPlanningCommand(cwd, "", "")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), planning.Command)
			return nil
		},
	}

	return cmd
}
