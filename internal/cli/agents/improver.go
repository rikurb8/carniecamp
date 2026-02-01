package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rikurb8/carnie/internal/config"
	"github.com/rikurb8/carnie/internal/prompts"
	"github.com/rikurb8/carnie/internal/session"
)

type Issue struct {
	ID          string
	Title       string
	Description string
	Status      string
	Priority    int
	IssueType   string
	Owner       string
}

type ImproverSessionMsg struct {
	Err error
}

const improverBasePrompt = `You are an epic improver and task splitter.
Your job is to refine the current issue into a crisp epic, identify missing details, and split it into actionable tasks.
Ask clarifying questions when needed, then propose a concrete plan with tasks and acceptance criteria.`

func CreateImproverSessionCmd(issue Issue, instructions string) tea.Cmd {
	return func() tea.Msg {
		cwd, err := os.Getwd()
		if err != nil {
			return ImproverSessionMsg{Err: fmt.Errorf("get working directory: %w", err)}
		}

		var campCfg *config.CampConfig
		model := config.DefaultOperatorModel
		configPath := filepath.Join(cwd, config.CampConfigFile)
		if cfg, err := config.LoadCampConfig(configPath); err == nil {
			campCfg = cfg
			if campCfg.Operator.Model != "" {
				model = campCfg.Operator.Model
			}
		}

		if !session.TmuxAvailable() {
			return ImproverSessionMsg{Err: fmt.Errorf("tmux is not installed or not in PATH")}
		}
		if !session.Available(session.ToolOpencode) {
			return ImproverSessionMsg{Err: fmt.Errorf("opencode is not installed or not in PATH")}
		}

		ctx := prompts.GatherContext(cwd, campCfg)
		systemPrompt := prompts.BuildSystemPromptWithBase(ctx, improverBasePrompt)
		prompt := buildImproverInitialPrompt(issue, instructions)

		sessionName := sanitizeSessionName(fmt.Sprintf("cn-epic-improver-%s", issue.ID))
		if sessionName == "" {
			sessionName = "cn-epic-improver"
		}
		if session.SessionExists(sessionName) {
			sessionName = fmt.Sprintf("%s-%s", sessionName, time.Now().Format("150405"))
		}

		opts := session.Options{
			Tool:         session.ToolOpencode,
			Model:        session.NormalizeModel(session.ToolOpencode, model),
			SystemPrompt: systemPrompt,
			Prompt:       prompt,
			WorkDir:      cwd,
			Interactive:  true,
			SessionName:  sessionName,
		}

		if err := session.Spawn(opts); err != nil {
			return ImproverSessionMsg{Err: err}
		}
		return ImproverSessionMsg{}
	}
}

func buildImproverInitialPrompt(issue Issue, instructions string) string {
	var b strings.Builder
	b.WriteString("Active issue:\n")
	b.WriteString(fmt.Sprintf("ID: %s\n", issue.ID))
	b.WriteString(fmt.Sprintf("Title: %s\n", issue.Title))
	if issue.IssueType != "" {
		b.WriteString(fmt.Sprintf("Type: %s\n", issue.IssueType))
	}
	if issue.Status != "" {
		b.WriteString(fmt.Sprintf("Status: %s\n", issue.Status))
	}
	if issue.Priority != 0 {
		b.WriteString(fmt.Sprintf("Priority: P%d\n", issue.Priority))
	}
	if issue.Owner != "" {
		b.WriteString(fmt.Sprintf("Owner: %s\n", issue.Owner))
	}
	if issue.Description != "" {
		b.WriteString("Description:\n")
		b.WriteString(issue.Description)
		b.WriteString("\n")
	}
	if strings.TrimSpace(instructions) != "" {
		b.WriteString("\nUser instructions:\n")
		b.WriteString(strings.TrimSpace(instructions))
		b.WriteString("\n")
	}
	b.WriteString("\nPlease improve this epic and split into tasks.")
	return b.String()
}

func sanitizeSessionName(name string) string {
	if name == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
