package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rikurb8/bordertown/internal/config"
	"github.com/spf13/cobra"
)

func newTownCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "town",
		Short: "Manage bordertown town config",
		Long:  "Commands for managing the town.yml configuration inside your project.",
	}

	cmd.AddCommand(newTownInitCommand())

	return cmd
}

func newTownInitCommand() *cobra.Command {
	var name string
	var description string
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new town",
		Long:  "Creates a town.yml configuration file in the current project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := filepath.Join(".", config.TownConfigFile)

			if _, err := os.Stat(configPath); err == nil && !force {
				return fmt.Errorf("%s already exists (use --force to overwrite)", config.TownConfigFile)
			}

			if name == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
				name = filepath.Base(cwd)
			}

			cfg := config.NewTownConfig(name)
			if description != "" {
				cfg.Description = description
			}

			if err := cfg.Write(configPath); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", config.TownConfigFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Town name (defaults to directory name)")
	cmd.Flags().StringVar(&description, "description", "", "Town description")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing town.yml")

	return cmd
}
