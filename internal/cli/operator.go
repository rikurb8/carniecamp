package cli

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/rikurb8/carnie/internal/operator"
	"github.com/spf13/cobra"
)

func newOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "operator",
		Aliases: []string{"op"},
		Short:   "Operator commands for planning and issue management",
		Long:    "Commands to help operators plan work and manage issues.",
	}

	cmd.AddCommand(newOperatorPlanCommand())
	cmd.AddCommand(newOperatorIssueToBeadsCommand())

	return cmd
}

func newOperatorPlanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Print the operator planning command",
		Long:  "Outputs a ready-to-paste command to start an operator planning session.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			planning, err := operator.BuildPlanningCommand(cwd, "", "")
			if err != nil {
				return err
			}

			if err := clipboard.WriteAll(planning.Command); err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), planning.Command)
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Command copied to clipboard")
			return nil
		},
	}
}

func newOperatorIssueToBeadsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "issue-to-beads <issue-number>",
		Short: "Convert a GitHub issue to beads tasks",
		Long:  "Fetches a GitHub issue and copies a planning command to clipboard.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issueNumber := args[0]

			issue, err := operator.FetchGHIssue(issueNumber)
			if err != nil {
				return err
			}

			planning, err := operator.BuildIssueToBeadsCommand(issue, "")
			if err != nil {
				return err
			}

			if err := clipboard.WriteAll(planning.Command); err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), planning.Command)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Command for issue #%d copied to clipboard\n", issue.Number)
			return nil
		},
	}
}
