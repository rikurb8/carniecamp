package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/atotto/clipboard"
	"github.com/rikurb8/carnie/internal/config"
	"github.com/rikurb8/carnie/internal/prime"
	"github.com/rikurb8/carnie/internal/workorder"
	"github.com/spf13/cobra"
)

func newWorkOrderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workorder",
		Aliases: []string{"wo"},
		Short:   "Manage work orders for agents",
	}

	cmd.AddCommand(newWorkOrderCreateCommand())
	cmd.AddCommand(newWorkOrderListCommand())
	cmd.AddCommand(newWorkOrderShowCommand())
	cmd.AddCommand(newWorkOrderUpdateCommand())
	cmd.AddCommand(newWorkOrderPromptCommand())

	return cmd
}

func newWorkOrderCreateCommand() *cobra.Command {
	var title string
	var description string
	var beadID string
	var status string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new work order",
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if description == "" {
				return fmt.Errorf("--description is required")
			}

			statusValue := workorder.StatusReady
			if status != "" {
				parsed, err := workorder.ParseStatus(status)
				if err != nil {
					return err
				}
				statusValue = parsed
			}

			store, err := openWorkOrderStore()
			if err != nil {
				return err
			}
			defer store.Close()

			order, err := store.Create(context.Background(), workorder.CreateInput{
				Title:       title,
				Description: description,
				BeadID:      beadID,
				Status:      statusValue,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created work order %d (%s)\n", order.ID, order.Status)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Work order title")
	cmd.Flags().StringVar(&description, "description", "", "Work order description")
	cmd.Flags().StringVar(&beadID, "bead", "", "Associated bead ID")
	cmd.Flags().StringVar(&status, "status", "", "Initial status (default: ready)")

	return cmd
}

func newWorkOrderListCommand() *cobra.Command {
	var status string
	var beadID string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work orders",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openWorkOrderStore()
			if err != nil {
				return err
			}
			defer store.Close()

			var statusFilter *workorder.Status
			if status != "" {
				parsed, err := workorder.ParseStatus(status)
				if err != nil {
					return err
				}
				statusFilter = &parsed
			}

			orders, err := store.List(context.Background(), workorder.ListOptions{
				Status: statusFilter,
				BeadID: beadID,
				Limit:  limit,
			})
			if err != nil {
				return err
			}

			beadIndex, _ := workorder.LoadBeadIndex(mustGetwd())

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(writer, "ID\tStatus\tBead\tBead Description\tTitle\tUpdated")
			for _, order := range orders {
				beadDesc := ""
				if info, ok := beadIndex[order.BeadID]; ok {
					beadDesc = info.Description
					if beadDesc == "" {
						beadDesc = info.Title
					}
				}
				fmt.Fprintf(
					writer,
					"%d\t%s\t%s\t%s\t%s\t%s\n",
					order.ID,
					order.Status,
					order.BeadID,
					truncateASCII(beadDesc, 50),
					truncateASCII(order.Title, 60),
					order.UpdatedAt.Format("2006-01-02 15:04"),
				)
			}
			return writer.Flush()
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&beadID, "bead", "", "Filter by bead ID")
	cmd.Flags().IntVar(&limit, "limit", 200, "Limit number of work orders")

	return cmd
}

func newWorkOrderShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a work order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid work order id %q", args[0])
			}

			store, err := openWorkOrderStore()
			if err != nil {
				return err
			}
			defer store.Close()

			order, err := store.Get(context.Background(), id)
			if err != nil {
				return err
			}

			beadIndex, _ := workorder.LoadBeadIndex(mustGetwd())
			beadInfo := beadIndex[order.BeadID]

			fmt.Fprintf(cmd.OutOrStdout(), "ID: %d\n", order.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\n", order.Title)
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", order.Status)
			if order.BeadID != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Bead: %s\n", order.BeadID)
			}
			if beadInfo.Title != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Bead Title: %s\n", beadInfo.Title)
			}
			if beadInfo.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Bead Description: %s\n", beadInfo.Description)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", order.CreatedAt.Format(time.RFC3339))
			fmt.Fprintf(cmd.OutOrStdout(), "Updated: %s\n", order.UpdatedAt.Format(time.RFC3339))
			if order.StartedAt != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Started: %s\n", order.StartedAt.Format(time.RFC3339))
			}
			if order.CompletedAt != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Completed: %s\n", order.CompletedAt.Format(time.RFC3339))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nDescription:\n%s\n", order.Description)
			return nil
		},
	}

	return cmd
}

func newWorkOrderUpdateCommand() *cobra.Command {
	var status string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a work order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if status == "" {
				return fmt.Errorf("--status is required")
			}
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid work order id %q", args[0])
			}

			next, err := workorder.ParseStatus(status)
			if err != nil {
				return err
			}

			store, err := openWorkOrderStore()
			if err != nil {
				return err
			}
			defer store.Close()

			order, err := store.UpdateStatus(context.Background(), id, next)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated work order %d to %s\n", order.ID, order.Status)
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "New status")
	return cmd
}

func newWorkOrderPromptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompt <id>",
		Short: "Render a work order prompt",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid work order id %q", args[0])
			}

			store, err := openWorkOrderStore()
			if err != nil {
				return err
			}
			defer store.Close()

			order, err := store.Get(context.Background(), id)
			if err != nil {
				return err
			}

			rolePrompt, err := prime.LoadPrompt(prime.RoleCarnie)
			if err != nil {
				return err
			}

			beadIndex, _ := workorder.LoadBeadIndex(mustGetwd())
			beadInfo := beadIndex[order.BeadID]

			projectName, projectDesc := loadCampMetadata()
			prompt, err := workorder.RenderPrompt(workorder.PromptData{
				RolePrompt:         rolePrompt,
				WorkOrder:          order,
				BeadTitle:          beadInfo.Title,
				BeadDescription:    beadInfo.Description,
				ProjectName:        projectName,
				ProjectDescription: projectDesc,
			})
			if err != nil {
				return err
			}

			if err := clipboard.WriteAll(prompt); err != nil {
				fmt.Fprint(cmd.OutOrStdout(), prompt)
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Prompt copied to clipboard")
			return nil
		},
	}

	return cmd
}

func openWorkOrderStore() (*workorder.Store, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	dbPath, err := workorder.DefaultDBPath(cwd)
	if err != nil {
		return nil, err
	}
	return workorder.OpenStore(dbPath)
}

func loadCampMetadata() (string, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", ""
	}
	root, err := workorder.FindCampRoot(cwd)
	if err != nil {
		return "", ""
	}
	cfg, err := config.LoadCampConfig(filepath.Join(root, config.CampConfigFile))
	if err != nil {
		return "", ""
	}
	return cfg.Name, cfg.Description
}

func truncateASCII(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(value) <= width {
		return value
	}
	if width <= 3 {
		return value[:width]
	}
	return value[:width-3] + "..."
}

func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}
