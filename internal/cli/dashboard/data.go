package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rikurb8/carnie/internal/beads"
	"github.com/rikurb8/carnie/internal/cli/bd"
)

func loadDashboardDataCmd(limit int) tea.Cmd {
	return func() tea.Msg {
		data, err := fetchDashboardData(limit)
		return dataMsg{Data: data, Err: err}
	}
}

func fetchDashboardData(limit int) (dataState, error) {
	statusOutput, err := bd.RunJSON("status", "--json")
	if err != nil {
		return dataState{}, err
	}

	var status bd.Status
	if err := json.Unmarshal(statusOutput, &status); err != nil {
		return dataState{}, fmt.Errorf("parse bd status: %w", err)
	}

	ready, inProgress, blocked, closed, err := fetchIssuesFromBeads(limit)
	if err != nil {
		return dataState{}, err
	}

	return dataState{
		Status:     status,
		Ready:      ready,
		InProgress: inProgress,
		Blocked:    blocked,
		Closed:     closed,
	}, nil
}

func fetchIssuesFromBeads(limit int) ([]Issue, []Issue, []Issue, []Issue, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("get working directory: %w", err)
	}
	root, err := beads.FindBeadsRoot(cwd)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("find beads: %w", err)
	}
	if err := exportBeadsJSONL(root); err != nil {
		return nil, nil, nil, nil, err
	}
	issues, err := beads.LoadIssues(root)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("load beads issues: %w", err)
	}

	readyIssues := filterIssuesByStatus(issues, "open")
	inProgressIssues := filterIssuesByStatus(issues, "in_progress")
	blockedIssues := filterIssuesByStatus(issues, "blocked")
	closedIssues := filterIssuesByStatus(issues, "closed")

	sort.SliceStable(readyIssues, func(i, j int) bool { return readyIssues[i].Priority < readyIssues[j].Priority })
	sort.SliceStable(inProgressIssues, func(i, j int) bool { return inProgressIssues[i].Priority < inProgressIssues[j].Priority })
	sort.SliceStable(blockedIssues, func(i, j int) bool { return blockedIssues[i].Priority < blockedIssues[j].Priority })
	sort.SliceStable(closedIssues, func(i, j int) bool { return closedIssues[i].UpdatedAt.After(closedIssues[j].UpdatedAt) })

	readyIssues = applyIssueLimit(readyIssues, limit)
	inProgressIssues = applyIssueLimit(inProgressIssues, limit)
	blockedIssues = applyIssueLimit(blockedIssues, limit)
	closedIssues = applyIssueLimit(closedIssues, limit)

	return mapBeadsIssues(readyIssues), mapBeadsIssues(inProgressIssues), mapBeadsIssues(blockedIssues), mapBeadsIssues(closedIssues), nil
}

func exportBeadsJSONL(root string) error {
	output := filepath.Join(root, beads.BeadsDir, beads.IssuesFile)
	command := exec.Command("bd", "export", "-o", output, "-q")
	if err := command.Run(); err != nil {
		return fmt.Errorf("bd export failed: %w", err)
	}
	return nil
}

func filterIssuesByStatus(issues []beads.Issue, status string) []beads.Issue {
	filtered := make([]beads.Issue, 0, len(issues))
	for _, issue := range issues {
		if issue.Status != status {
			continue
		}
		filtered = append(filtered, issue)
	}
	return filtered
}

func applyIssueLimit(issues []beads.Issue, limit int) []beads.Issue {
	if limit <= 0 || len(issues) <= limit {
		return issues
	}
	return issues[:limit]
}

func mapBeadsIssues(issues []beads.Issue) []Issue {
	output := make([]Issue, 0, len(issues))
	for _, issue := range issues {
		deps := make([]Dependency, 0, len(issue.Dependencies))
		for _, dep := range issue.Dependencies {
			deps = append(deps, Dependency{IssueID: dep.IssueID, DependsOnID: dep.DependsOnID, Type: dep.Type})
		}
		output = append(output, Issue{
			ID:           issue.ID,
			Title:        issue.Title,
			Description:  issue.Description,
			Status:       issue.Status,
			Priority:     issue.Priority,
			IssueType:    issue.IssueType,
			Owner:        issue.Owner,
			UpdatedAt:    issue.UpdatedAt.Format(time.RFC3339),
			CreatedAt:    issue.CreatedAt.Format(time.RFC3339),
			Dependencies: deps,
		})
	}
	return output
}

func fetchIssues(args []string, limit int) ([]Issue, error) {
	if limit > 0 {
		args = append(args, "--limit", strconv.Itoa(limit))
	}
	output, err := bd.RunJSON(args...)
	if err != nil {
		return nil, err
	}
	return parseBdIssues(output)
}

func parseBdIssues(data []byte) ([]Issue, error) {
	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("parse bd issues: %w", err)
	}
	return issues, nil
}

func orderIssuesWithParents(issues []Issue) ([]Issue, map[string]string, map[string]int) {
	ordered := make([]Issue, 0, len(issues))
	parentByID := make(map[string]string)
	levelByID := make(map[string]int)
	added := make(map[string]bool)

	issueByID := make(map[string]Issue, len(issues))
	for _, issue := range issues {
		issueByID[issue.ID] = issue
	}

	childrenByID := make(map[string][]string)
	for _, issue := range issues {
		for _, dep := range issue.Dependencies {
			childID := dep.DependsOnID
			if childID == "" {
				continue
			}
			if _, ok := issueByID[childID]; !ok {
				continue
			}
			if dep.Type == "parent-child" {
				parentByID[issue.ID] = childID
				childrenByID[childID] = append(childrenByID[childID], issue.ID)
				continue
			}
			childrenByID[issue.ID] = append(childrenByID[issue.ID], childID)
		}
	}

	for parentID := range childrenByID {
		children := childrenByID[parentID]
		if len(children) < 2 {
			continue
		}
		sorted := make([]string, 0, len(children))
		seen := make(map[string]bool)
		for _, issue := range issues {
			if !containsID(children, issue.ID) {
				continue
			}
			if seen[issue.ID] {
				continue
			}
			sorted = append(sorted, issue.ID)
			seen[issue.ID] = true
		}
		childrenByID[parentID] = sorted
	}

	visiting := make(map[string]bool)
	var visit func(id string, level int)
	visit = func(id string, level int) {
		if added[id] {
			return
		}
		if visiting[id] {
			return
		}
		issue, ok := issueByID[id]
		if !ok {
			return
		}
		visiting[id] = true
		ordered = append(ordered, issue)
		added[id] = true
		levelByID[id] = level
		for _, childID := range childrenByID[id] {
			if parentByID[childID] == "" {
				parentByID[childID] = id
			}
			visit(childID, level+1)
		}
		visiting[id] = false
	}

	for _, issue := range issues {
		if issue.IssueType != "epic" {
			continue
		}
		visit(issue.ID, 0)
	}

	for _, issue := range issues {
		if added[issue.ID] {
			continue
		}
		visit(issue.ID, 0)
	}

	return ordered, parentByID, levelByID
}

func containsID(ids []string, id string) bool {
	for _, value := range ids {
		if value == id {
			return true
		}
	}
	return false
}

func buildDrawerEntries(column issueColumn, collapsed map[string]bool) []drawerEntry {
	if len(column.Issues) == 0 {
		return nil
	}
	parentByID := column.ParentByID
	if parentByID == nil {
		parentByID = map[string]string{}
	}
	levelByID := column.LevelByID
	if levelByID == nil {
		levelByID = map[string]int{}
	}
	issueByID := make(map[string]Issue, len(column.Issues))
	for _, issue := range column.Issues {
		issueByID[issue.ID] = issue
	}
	entries := make([]drawerEntry, 0, len(column.Issues))
	for _, issue := range column.Issues {
		if isHiddenByCollapse(issue.ID, parentByID, issueByID, collapsed) {
			continue
		}
		entries = append(entries, drawerEntry{Issue: issue, Level: levelByID[issue.ID]})
	}
	return entries
}

func isHiddenByCollapse(id string, parentByID map[string]string, issueByID map[string]Issue, collapsed map[string]bool) bool {
	parent := parentByID[id]
	for parent != "" {
		issue, ok := issueByID[parent]
		if ok && issue.IssueType == "epic" && collapsed[parent] {
			return true
		}
		parent = parentByID[parent]
	}
	return false
}
