package operator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"text/template"
	"time"

	"github.com/rikurb8/carnie/internal/session"
	"github.com/rikurb8/carnie/internal/templates"
)

// GHUser represents a GitHub user (author, assignee, commenter)
type GHUser struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	IsBot bool   `json:"is_bot"`
}

// GHLabel represents a GitHub issue label
type GHLabel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// GHMilestone represents a GitHub milestone
type GHMilestone struct {
	ID          string    `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueOn       time.Time `json:"dueOn"`
}

// GHComment represents a comment on a GitHub issue
type GHComment struct {
	ID        string    `json:"id"`
	Author    GHUser    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	URL       string    `json:"url"`
}

// GHIssue represents a GitHub issue with all relevant fields
type GHIssue struct {
	ID          string       `json:"id"`
	Number      int          `json:"number"`
	Title       string       `json:"title"`
	Body        string       `json:"body"`
	State       string       `json:"state"`
	StateReason string       `json:"stateReason"`
	Author      GHUser       `json:"author"`
	Assignees   []GHUser     `json:"assignees"`
	Labels      []GHLabel    `json:"labels"`
	Milestone   *GHMilestone `json:"milestone"`
	Comments    []GHComment  `json:"comments"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	ClosedAt    *time.Time   `json:"closedAt"`
	Closed      bool         `json:"closed"`
	URL         string       `json:"url"`
}

// ghIssueJSONFields are the fields to request from gh issue view --json
var ghIssueJSONFields = "assignees,author,body,closed,closedAt,comments,createdAt,id,labels,milestone,number,state,stateReason,title,updatedAt,url"

// FetchGHIssue fetches a GitHub issue by number using the gh CLI
func FetchGHIssue(issueNumber string) (*GHIssue, error) {
	cmd := exec.Command("gh", "issue", "view", issueNumber, "--json", ghIssueJSONFields)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh issue view failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh issue view failed: %w", err)
	}

	var issue GHIssue
	if err := json.Unmarshal(output, &issue); err != nil {
		return nil, fmt.Errorf("parse issue JSON: %w", err)
	}

	return &issue, nil
}

// PrintIssue prints the parsed issue data in a readable format
func PrintIssue(issue *GHIssue) {
	fmt.Printf("Issue #%d: %s\n", issue.Number, issue.Title)
	fmt.Printf("State: %s\n", issue.State)
	fmt.Printf("Author: %s (%s)\n", issue.Author.Name, issue.Author.Login)
	fmt.Printf("URL: %s\n", issue.URL)
	fmt.Printf("Created: %s\n", issue.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", issue.UpdatedAt.Format(time.RFC3339))

	if len(issue.Assignees) > 0 {
		fmt.Printf("Assignees: ")
		for i, a := range issue.Assignees {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", a.Login)
		}
		fmt.Println()
	}

	if len(issue.Labels) > 0 {
		fmt.Printf("Labels: ")
		for i, l := range issue.Labels {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", l.Name)
		}
		fmt.Println()
	}

	if issue.Milestone != nil {
		fmt.Printf("Milestone: %s\n", issue.Milestone.Title)
	}

	fmt.Printf("\n--- Body ---\n%s\n", issue.Body)

	if len(issue.Comments) > 0 {
		fmt.Printf("\n--- Comments (%d) ---\n", len(issue.Comments))
		for _, c := range issue.Comments {
			fmt.Printf("\n[%s] %s:\n%s\n", c.CreatedAt.Format("2006-01-02 15:04"), c.Author.Login, c.Body)
		}
	}
}

// RenderIssueToBeadsPrompt renders the issue-to-beads template with issue data
func RenderIssueToBeadsPrompt(issue *GHIssue) (string, error) {
	tmplContent, err := templates.Load("issue-to-beads.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("load template: %w", err)
	}

	tmpl, err := template.New("issue-to-beads").Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, issue); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// IssueToBeadsCommand contains the command to start an issue-to-beads session
type IssueToBeadsCommand struct {
	Command string
	Tool    session.Tool
	Model   string
}

// BuildIssueToBeadsCommand builds a command to start opencode with the issue-to-beads prompt
func BuildIssueToBeadsCommand(issue *GHIssue, model string) (IssueToBeadsCommand, error) {
	prompt, err := RenderIssueToBeadsPrompt(issue)
	if err != nil {
		return IssueToBeadsCommand{}, err
	}

	tool := session.ToolOpencode
	if model == "" {
		model = "openai/gpt-5.2-codex"
	}
	model = session.NormalizeModel(tool, model)

	opts := session.Options{
		Tool:         tool,
		Model:        model,
		SystemPrompt: prompt,
		Interactive:  true,
	}

	command := session.Command(opts)
	if command == "" {
		return IssueToBeadsCommand{}, fmt.Errorf("build issue-to-beads command")
	}

	return IssueToBeadsCommand{Command: command, Tool: tool, Model: model}, nil
}
