package prompts

import (
	"fmt"
	"strings"

	"github.com/rikurb8/carnie/internal/beads"
	"github.com/rikurb8/carnie/internal/config"
)

// ProjectContext holds information about the current project for context injection.
type ProjectContext struct {
	ProjectName        string
	ProjectDescription string
	Epics              []EpicSummary
}

// EpicSummary is a brief summary of an existing epic.
type EpicSummary struct {
	ID          string
	Title       string
	Status      string
	TaskCount   int
	Description string
}

// GatherContext collects project context from camp.yml and beads.
func GatherContext(workDir string, campCfg *config.CampConfig) ProjectContext {
	ctx := ProjectContext{}

	// Get project info from camp config
	if campCfg != nil {
		ctx.ProjectName = campCfg.Name
		ctx.ProjectDescription = campCfg.Description
	}

	// Try to load epics from beads
	if root, err := beads.FindBeadsRoot(workDir); err == nil {
		if issues, err := beads.LoadIssues(root); err == nil {
			for _, issue := range issues {
				if issue.IssueType == "epic" {
					summary := EpicSummary{
						ID:          issue.ID,
						Title:       issue.Title,
						Status:      issue.Status,
						Description: truncate(issue.Description, 100),
					}
					// Count tasks that depend on this epic
					for _, other := range issues {
						for _, dep := range other.Dependencies {
							if dep.DependsOnID == issue.ID {
								summary.TaskCount++
							}
						}
					}
					ctx.Epics = append(ctx.Epics, summary)
				}
			}
		}
	}

	return ctx
}

// FormatContextSection formats the project context as a section to prepend to the system prompt.
func FormatContextSection(ctx ProjectContext) string {
	if ctx.ProjectName == "" && len(ctx.Epics) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Project Context\n\n")

	if ctx.ProjectName != "" {
		fmt.Fprintf(&sb, "**Project:** %s\n", ctx.ProjectName)
		if ctx.ProjectDescription != "" {
			fmt.Fprintf(&sb, "**Description:** %s\n", ctx.ProjectDescription)
		}
		sb.WriteString("\n")
	}

	if len(ctx.Epics) > 0 {
		sb.WriteString("**Existing Epics:**\n")
		for _, epic := range ctx.Epics {
			status := "open"
			if epic.Status == "closed" {
				status = "closed"
			}
			fmt.Fprintf(&sb, "- %s: %s (%s, %d tasks)\n",
				epic.ID, epic.Title, status, epic.TaskCount)
		}
		sb.WriteString("\nConsider how this new epic relates to existing work.\n")
	}

	sb.WriteString("\n---\n\n")
	return sb.String()
}

// BuildSystemPrompt combines project context with the base planning prompt.
func BuildSystemPrompt(ctx ProjectContext) string {
	return BuildSystemPromptWithBase(ctx, EpicPlanningPrompt)
}

// BuildSystemPromptWithBase combines project context with a custom base prompt.
func BuildSystemPromptWithBase(ctx ProjectContext, basePrompt string) string {
	contextSection := FormatContextSection(ctx)
	return contextSection + basePrompt
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
