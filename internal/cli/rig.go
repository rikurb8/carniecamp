package cli

import (
	"fmt"

	"github.com/rikurb8/bordertown/internal/rig"
	"github.com/spf13/cobra"
)

func newRigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rig",
		Short: "Manage rigs",
		Long:  "Commands for adding and managing project rigs.",
	}

	cmd.AddCommand(newRigAddCommand())

	return cmd
}

func newRigAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <ssh-url>",
		Short: "Add a rig by cloning a repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			remote := args[1]

			newRig, err := rig.AddRig(name, remote)
			if err != nil {
				return fmt.Errorf("%s", rig.UserMessage(err))
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added rig %q at %s\n", newRig.Name, newRig.LocalPath)
			return nil
		},
	}

	return cmd
}
