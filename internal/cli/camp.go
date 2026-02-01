package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rikurb8/carnie/internal/config"
	"github.com/spf13/cobra"
)

func newCampCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "camp",
		Short: "Manage carnie camp config",
		Long:  "Commands for managing the camp.yml configuration inside your project.",
	}

	cmd.AddCommand(newCampInitCommand())

	return cmd
}

func newCampInitCommand() *cobra.Command {
	var name string
	var description string
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new camp",
		Long:  "Creates a camp.yml configuration file in the current project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := filepath.Join(".", config.CampConfigFile)

			if _, err := os.Stat(configPath); err == nil && !force {
				return fmt.Errorf("%s already exists (use --force to overwrite)", config.CampConfigFile)
			}

			if name == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
				name = filepath.Base(cwd)
			}

			cfg := config.NewCampConfig(name)
			if description != "" {
				cfg.Description = description
			}

			if err := cfg.Write(configPath); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", config.CampConfigFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Camp name (defaults to directory name)")
	cmd.Flags().StringVar(&description, "description", "", "Camp description")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing camp.yml")

	return cmd
}
