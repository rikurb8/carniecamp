package dashboard

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rikurb8/carnie/internal/cli/agents"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadDashboardDataCmd(m.limit), m.tickCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		if m.showImprover {
			m.applyImproverLayout()
		}
		m.ensureVisible()
		return m, nil
	case tea.KeyMsg:
		if m.showImprover {
			return m, m.updateImproverInput(typed)
		}
		if m.showHelp {
			switch typed.String() {
			case "h", "esc":
				m.showHelp = false
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}

		switch typed.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "h", "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "o":
			m.showImprover = true
			m.improveInput.SetValue("")
			m.improveInput.Focus()
			m.applyImproverLayout()
			return m, nil
		case "left":
			m.setCollapseForSelected(true)
			m.ensureVisible()
			return m, nil
		case "right":
			m.setCollapseForSelected(false)
			m.ensureVisible()
			return m, nil
		case "tab", "l":
			m.activeColumn = (m.activeColumn + 1) % len(m.columns)
			m.ensureVisible()
			return m, nil
		case "shift+tab":
			m.activeColumn = (m.activeColumn - 1 + len(m.columns)) % len(m.columns)
			m.ensureVisible()
			return m, nil
		case "down", "j":
			m.moveSelection(1)
			return m, nil
		case "up", "k":
			m.moveSelection(-1)
			return m, nil
		case "r":
			return m, loadDashboardDataCmd(m.limit)
		}
	case tickMsg:
		return m, tea.Batch(loadDashboardDataCmd(m.limit), m.tickCmd())
	case dataMsg:
		if typed.Err != nil {
			m.errMessage = typed.Err.Error()
			return m, nil
		}
		m.errMessage = ""
		m.lastUpdated = time.Now()
		m.summary = typed.Data.Status.Summary
		m.updateColumns(typed.Data)
		m.ensureVisible()
		return m, nil
	case agents.ImproverSessionMsg:
		if typed.Err != nil {
			m.errMessage = typed.Err.Error()
			return m, nil
		}
		m.errMessage = ""
		return m, nil
	}

	return m, nil
}

func (m Model) tickCmd() tea.Cmd {
	if m.refresh <= 0 {
		return nil
	}
	return tea.Tick(m.refresh, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m *Model) moveSelection(delta int) {
	column := &m.columns[m.activeColumn]
	entries := buildDrawerEntries(*column, m.collapsed)
	if len(entries) == 0 {
		column.Selected = 0
		column.Offset = 0
		return
	}

	next := column.Selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(entries) {
		next = len(entries) - 1
	}
	column.Selected = next
	m.ensureVisible()
}

func (m *Model) setCollapseForSelected(collapse bool) {
	if len(m.columns) == 0 {
		return
	}
	column := m.columns[m.activeColumn]
	entries := buildDrawerEntries(column, m.collapsed)
	if len(entries) == 0 {
		return
	}
	if column.Selected < 0 || column.Selected >= len(entries) {
		return
	}
	issue := entries[column.Selected].Issue
	if issue.IssueType != "epic" {
		return
	}
	if collapse {
		m.collapsed[issue.ID] = true
		return
	}
	delete(m.collapsed, issue.ID)
}

func (m *Model) ensureVisible() {
	height := availableListHeight(m.height)
	for idx := range m.columns {
		column := &m.columns[idx]
		entries := buildDrawerEntries(*column, m.collapsed)
		if len(entries) == 0 {
			column.Selected = 0
			column.Offset = 0
			continue
		}

		if column.Selected < 0 {
			column.Selected = 0
		}
		if column.Selected >= len(entries) {
			column.Selected = len(entries) - 1
		}

		if column.Selected < column.Offset {
			column.Offset = column.Selected
		}
		if column.Selected >= column.Offset+height {
			column.Offset = column.Selected - height + 1
		}
		if column.Offset < 0 {
			column.Offset = 0
		}
		if column.Offset > len(entries)-1 {
			column.Offset = len(entries) - 1
		}
	}
}

func (m *Model) updateColumns(data dataState) {
	previous := make(map[string]string, len(m.columns))
	for _, column := range m.columns {
		selectedID := selectedIssueID(column, m.collapsed)
		if selectedID == "" {
			continue
		}
		previous[column.Title] = selectedID
	}

	future := append([]Issue{}, data.Ready...)
	future = append(future, data.InProgress...)
	future = append(future, data.Blocked...)

	orderedFuture, futureParents, futureLevels := orderIssuesWithParents(future)
	orderedClosed, closedParents, closedLevels := orderIssuesWithParents(data.Closed)

	m.columns[0].Issues = orderedFuture
	m.columns[0].ParentByID = futureParents
	m.columns[0].LevelByID = futureLevels
	m.columns[1].Issues = orderedClosed
	m.columns[1].ParentByID = closedParents
	m.columns[1].LevelByID = closedLevels

	for idx := range m.columns {
		selectedID := previous[m.columns[idx].Title]
		entries := buildDrawerEntries(m.columns[idx], m.collapsed)
		if selectedID == "" || len(entries) == 0 {
			m.columns[idx].Selected = 0
			m.columns[idx].Offset = 0
			continue
		}
		found := false
		for entryIndex, entry := range entries {
			if entry.Issue.ID == selectedID {
				m.columns[idx].Selected = entryIndex
				found = true
				break
			}
		}
		if !found {
			m.columns[idx].Selected = 0
			m.columns[idx].Offset = 0
		}
	}
}

func selectedIssueID(column issueColumn, collapsed map[string]bool) string {
	entries := buildDrawerEntries(column, collapsed)
	if len(entries) == 0 {
		return ""
	}
	if column.Selected < 0 || column.Selected >= len(entries) {
		return ""
	}
	return entries[column.Selected].Issue.ID
}

func (m Model) selectedIssue() *Issue {
	if len(m.columns) == 0 {
		return nil
	}
	column := m.columns[m.activeColumn]
	entries := buildDrawerEntries(column, m.collapsed)
	if len(entries) == 0 || column.Selected >= len(entries) {
		return nil
	}
	issue := entries[column.Selected].Issue
	return &issue
}

func (m *Model) updateImproverInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.showImprover = false
		m.improveInput.Blur()
		return nil
	case "ctrl+c", "q":
		return tea.Quit
	case "enter":
		selected := m.selectedIssue()
		if selected == nil {
			m.errMessage = "Select an issue to improve"
			return nil
		}
		instructions := strings.TrimSpace(m.improveInput.Value())
		m.showImprover = false
		m.improveInput.Blur()
		return agents.CreateImproverSessionCmd(toAgentIssue(*selected), instructions)
	}
	var cmd tea.Cmd
	m.improveInput, cmd = m.improveInput.Update(msg)
	return cmd
}

func toAgentIssue(issue Issue) agents.Issue {
	return agents.Issue{
		ID:          issue.ID,
		Title:       issue.Title,
		Description: issue.Description,
		Status:      issue.Status,
		Priority:    issue.Priority,
		IssueType:   issue.IssueType,
		Owner:       issue.Owner,
	}
}
