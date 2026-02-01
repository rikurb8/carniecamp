package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rikurb8/carnie/internal/cli/bd"
	"github.com/spf13/cobra"
)

type bdInfo struct {
	Config struct {
		IssuePrefix string `json:"issue_prefix"`
	} `json:"config"`
	DaemonConnected bool   `json:"daemon_connected"`
	DaemonStatus    string `json:"daemon_status"`
	DaemonVersion   string `json:"daemon_version"`
	DatabasePath    string `json:"database_path"`
	IssueCount      int    `json:"issue_count"`
	Mode            string `json:"mode"`
}

type kvRow struct {
	Label string
	Value string
}

func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project details and beads summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			infoOutput, err := bd.RunJSON("info", "--json")
			if err != nil {
				return err
			}

			statusOutput, err := bd.RunJSON("status", "--json")
			if err != nil {
				return err
			}

			var info bdInfo
			if err := json.Unmarshal(infoOutput, &info); err != nil {
				return fmt.Errorf("parse bd info: %w", err)
			}

			var status bd.Status
			if err := json.Unmarshal(statusOutput, &status); err != nil {
				return fmt.Errorf("parse bd status: %w", err)
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve working directory: %w", err)
			}

			projectName := filepath.Base(cwd)
			daemonState := info.DaemonStatus
			if !info.DaemonConnected {
				daemonState = "disconnected"
			}

			projectRows := []kvRow{
				{Label: "Project", Value: projectName},
				{Label: "Path", Value: cwd},
				{Label: "Issue Prefix", Value: info.Config.IssuePrefix},
				{Label: "Database", Value: info.DatabasePath},
				{Label: "Mode", Value: info.Mode},
				{Label: "Daemon", Value: fmt.Sprintf("%s (%s)", daemonState, info.DaemonVersion)},
				{Label: "Issue Count", Value: fmt.Sprintf("%d", info.IssueCount)},
			}

			issueRows := []kvRow{
				{Label: "Total", Value: fmt.Sprintf("%d", status.Summary.TotalIssues)},
				{Label: "Open", Value: fmt.Sprintf("%d", status.Summary.OpenIssues)},
				{Label: "In Progress", Value: fmt.Sprintf("%d", status.Summary.InProgressIssues)},
				{Label: "Blocked", Value: fmt.Sprintf("%d", status.Summary.BlockedIssues)},
				{Label: "Ready", Value: fmt.Sprintf("%d", status.Summary.ReadyIssues)},
				{Label: "Deferred", Value: fmt.Sprintf("%d", status.Summary.DeferredIssues)},
				{Label: "Closed", Value: fmt.Sprintf("%d", status.Summary.ClosedIssues)},
				{Label: "Pinned", Value: fmt.Sprintf("%d", status.Summary.PinnedIssues)},
				{Label: "Tombstones", Value: fmt.Sprintf("%d", status.Summary.TombstoneIssues)},
				{Label: "Epics to Close", Value: fmt.Sprintf("%d", status.Summary.EpicsEligibleForClosure)},
				{Label: "Avg Lead Time", Value: fmt.Sprintf("%.1f hrs", status.Summary.AverageLeadTimeHours)},
			}

			activityRows := []kvRow{
				{Label: "Hours Tracked", Value: fmt.Sprintf("%d", status.RecentActivity.HoursTracked)},
				{Label: "Commits", Value: fmt.Sprintf("%d", status.RecentActivity.CommitCount)},
				{Label: "Issues Created", Value: fmt.Sprintf("%d", status.RecentActivity.IssuesCreated)},
				{Label: "Issues Closed", Value: fmt.Sprintf("%d", status.RecentActivity.IssuesClosed)},
				{Label: "Issues Updated", Value: fmt.Sprintf("%d", status.RecentActivity.IssuesUpdated)},
				{Label: "Issues Reopened", Value: fmt.Sprintf("%d", status.RecentActivity.IssuesReopened)},
				{Label: "Total Changes", Value: fmt.Sprintf("%d", status.RecentActivity.TotalChanges)},
			}

			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
			sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("36"))
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))

			renderSection := func(title, body string) string {
				return lipgloss.JoinVertical(lipgloss.Left, sectionStyle.Render(title), body)
			}

			renderRows := func(rows []kvRow) string {
				maxLabel := 0
				for _, row := range rows {
					if len(row.Label) > maxLabel {
						maxLabel = len(row.Label)
					}
				}

				lines := make([]string, 0, len(rows))
				for _, row := range rows {
					label := labelStyle.Render(padRight(row.Label, maxLabel))
					value := valueStyle.Render(row.Value)
					lines = append(lines, fmt.Sprintf("%s  %s", label, value))
				}

				return lipgloss.JoinVertical(lipgloss.Left, lines...)
			}

			output := lipgloss.JoinVertical(
				lipgloss.Left,
				titleStyle.Render("Carnie Status"),
				"",
				renderSection("Project", renderRows(projectRows)),
				"",
				renderSection("Issues", renderRows(issueRows)),
				"",
				renderSection("Recent Activity", renderRows(activityRows)),
			)

			fmt.Fprintln(cmd.OutOrStdout(), output)
			return nil
		},
	}

	return cmd
}

func padRight(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}
