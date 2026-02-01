package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rikurb8/carnie/internal/beads"
	"github.com/rikurb8/carnie/internal/config"
	"github.com/rikurb8/carnie/internal/prompts"
	"github.com/rikurb8/carnie/internal/session"
	"github.com/spf13/cobra"
)

func newOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "operator",
		Aliases: []string{"op"},
		Short:   "Project oversight and planning",
		Long:    "Commands for reviewing project status, epic planning, and work prioritization.",
	}

	cmd.AddCommand(newOperatorReviewCommand())
	cmd.AddCommand(newOperatorNewEpicCommand())

	return cmd
}

func newOperatorReviewCommand() *cobra.Command {
	var showClosed bool

	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review epics and their planning status",
		Long:  "Analyzes beads issues grouped by epic and indicates which epics need more planning.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			root, err := beads.FindBeadsRoot(cwd)
			if err != nil {
				return fmt.Errorf("find beads: %w", err)
			}

			issues, err := beads.LoadIssues(root)
			if err != nil {
				return fmt.Errorf("load issues: %w", err)
			}

			grouped := beads.GroupByEpic(issues)

			// Styles
			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
			epicStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("36"))
			closedEpicStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			taskStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("230"))
			closedTaskStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			orphanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("203"))

			var output strings.Builder

			output.WriteString(titleStyle.Render("Operator Review"))
			output.WriteString("\n\n")

			// Show epics
			epicsToShow := grouped.Epics
			if !showClosed {
				epicsToShow = grouped.OpenEpics()
			}

			if len(epicsToShow) == 0 {
				output.WriteString(dimStyle.Render("No epics found."))
				output.WriteString("\n")
			}

			for _, eg := range epicsToShow {
				summary := beads.SummarizeEpic(eg)

				// Epic header
				var epicHeader string
				statusIndicator := "○"
				if !eg.Epic.IsOpen() {
					statusIndicator = "●"
					epicHeader = closedEpicStyle.Render(fmt.Sprintf("%s %s [%s] - %s",
						statusIndicator, eg.Epic.ID, "closed", eg.Epic.Title))
				} else {
					epicHeader = epicStyle.Render(fmt.Sprintf("%s %s - %s",
						statusIndicator, eg.Epic.ID, eg.Epic.Title))
				}
				output.WriteString(epicHeader)
				output.WriteString("\n")

				// Task count
				taskCount := dimStyle.Render(fmt.Sprintf("  Tasks: %d open, %d closed",
					summary.OpenTasks, summary.ClosedTasks))
				output.WriteString(taskCount)
				output.WriteString("\n")

				// Planning status
				if summary.Planning.NeedsPlanning {
					warning := warningStyle.Render(fmt.Sprintf("  ⚠ Needs planning: %s",
						strings.Join(summary.Planning.Reasons, ", ")))
					output.WriteString(warning)
					output.WriteString("\n")
				}

				// List tasks (indented)
				for _, task := range eg.Children {
					var taskLine string
					if task.IsOpen() {
						taskLine = taskStyle.Render(fmt.Sprintf("    ○ %s - %s", task.ID, task.Title))
					} else {
						taskLine = closedTaskStyle.Render(fmt.Sprintf("    ● %s - %s", task.ID, task.Title))
					}
					output.WriteString(taskLine)
					output.WriteString("\n")
				}

				output.WriteString("\n")
			}

			// Show orphans
			orphans := grouped.Orphans
			if !showClosed {
				orphans = grouped.OpenOrphans()
			}

			if len(orphans) > 0 {
				output.WriteString(orphanStyle.Render("Orphan Issues (no epic)"))
				output.WriteString("\n")

				for _, issue := range orphans {
					var line string
					if issue.IsOpen() {
						line = taskStyle.Render(fmt.Sprintf("  ○ %s [%s] - %s",
							issue.ID, issue.IssueType, issue.Title))
					} else {
						line = closedTaskStyle.Render(fmt.Sprintf("  ● %s [%s] - %s",
							issue.ID, issue.IssueType, issue.Title))
					}
					output.WriteString(line)
					output.WriteString("\n")
				}
			}

			fmt.Fprint(cmd.OutOrStdout(), output.String())
			return nil
		},
	}

	cmd.Flags().BoolVar(&showClosed, "all", false, "Include closed epics and issues")

	return cmd
}

func newOperatorNewEpicCommand() *cobra.Command {
	var title string
	var tool string

	cmd := &cobra.Command{
		Use:     "new-epic",
		Aliases: []string{"plan-epic"},
		Short:   "Start a planning session for a new epic",
		Long:    "Spawns an AI session to help plan, refine, and create a new epic with tasks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			// Try to load camp config for tool preference and context
			var campCfg *config.CampConfig
			planningTool := config.DefaultPlanningTool
			model := config.DefaultOperatorModel
			configPath := filepath.Join(cwd, config.CampConfigFile)
			if cfg, err := config.LoadCampConfig(configPath); err == nil {
				campCfg = cfg
				if campCfg.Operator.PlanningTool != "" {
					planningTool = campCfg.Operator.PlanningTool
				}
				if campCfg.Operator.Model != "" {
					model = campCfg.Operator.Model
				}
			}

			// Command-line flag overrides config
			if tool != "" {
				planningTool = tool
			}

			selectedTool := session.ParseTool(planningTool)
			model = session.NormalizeModel(selectedTool, model)

			// Check if tmux is available
			if !session.TmuxAvailable() {
				return fmt.Errorf("tmux is not installed or not in PATH")
			}

			// Check if tool is available
			if !session.Available(selectedTool) {
				return fmt.Errorf("%s is not installed or not in PATH", selectedTool)
			}

			// Load custom prompt or use built-in default
			var promptFilePath string
			if campCfg != nil {
				promptFilePath = campCfg.Operator.PlanningPromptFile
			}
			basePrompt := prompts.LoadEpicPlanningPrompt(cwd, promptFilePath)

			// Gather project context and build system prompt
			ctx := prompts.GatherContext(cwd, campCfg)
			systemPrompt := prompts.BuildSystemPromptWithBase(ctx, basePrompt)

			// Build the session options
			sessionName := "cn-epic-planning"
			opts := session.Options{
				Tool:         selectedTool,
				Model:        model,
				SystemPrompt: systemPrompt,
				Prompt:       prompts.EpicPlanningInitialPrompt(title),
				WorkDir:      cwd,
				Interactive:  true,
				SessionName:  sessionName,
			}

			spawnCommand, err := session.SpawnCommand(opts)
			if err != nil {
				return fmt.Errorf("build spawn command: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Command: %s\n", spawnCommand)
			fmt.Fprintf(cmd.OutOrStdout(), "Starting epic planning session with %s in tmux...\n", selectedTool)
			fmt.Fprintf(cmd.OutOrStdout(), "Session: %s (reattach with: tmux attach -t %s)\n\n", sessionName, sessionName)

			return session.Spawn(opts)
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "Initial title or idea for the epic")
	cmd.Flags().StringVar(&tool, "tool", "", "Override planning tool (claude or opencode)")

	return cmd
}
