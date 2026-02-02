package dashboard

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{loadDashboardDataCmd(m.limit), m.tickCmd()}
	for i := range m.homeSpinners {
		cmds = append(cmds, m.homeSpinners[i].Tick)
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		m.applyDrawerLayout()
		return m, nil
	case tea.KeyMsg:
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
		case "left":
			m.setCollapseForSelected(true)
			drawerWidth, _, _, _ := drawerLayout(m.width, m.height)
			innerWidth := maxInt(1, drawerWidth-2)
			m.refreshDrawerList(m.activeColumn, true, innerWidth)
			return m, nil
		case "right":
			m.setCollapseForSelected(false)
			drawerWidth, _, _, _ := drawerLayout(m.width, m.height)
			innerWidth := maxInt(1, drawerWidth-2)
			m.refreshDrawerList(m.activeColumn, true, innerWidth)
			return m, nil
		case "tab", "l":
			m.activeColumn = (m.activeColumn + 1) % len(m.columns)
			return m, nil
		case "shift+tab":
			m.activeColumn = (m.activeColumn - 1 + len(m.columns)) % len(m.columns)
			return m, nil
		case "r":
			return m, loadDashboardDataCmd(m.limit)
		case "1":
			m.activeView = ViewHome
			return m, nil
		case "2":
			m.activeView = ViewTasks
			return m, nil
		}

		if len(m.lists) > 0 {
			listModel := m.lists[m.activeColumn]
			updated, cmd := listModel.Update(msg)
			m.lists[m.activeColumn] = updated
			return m, cmd
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
		drawerWidth, _, _, _ := drawerLayout(m.width, m.height)
		innerWidth := maxInt(1, drawerWidth-2)
		m.refreshDrawerLists(true, innerWidth)
		return m, nil
	case spinner.TickMsg:
		var cmds []tea.Cmd
		for i := range m.homeSpinners {
			var cmd tea.Cmd
			m.homeSpinners[i], cmd = m.homeSpinners[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
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

func (m *Model) setCollapseForSelected(collapse bool) {
	if len(m.columns) == 0 {
		return
	}
	issue := m.selectedIssue()
	if issue == nil {
		return
	}
	if issue.IssueType != "epic" {
		return
	}
	if collapse {
		m.collapsed[issue.ID] = true
		return
	}
	delete(m.collapsed, issue.ID)
}

func (m *Model) updateColumns(data dataState) {
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
}

func (m Model) selectedIssue() *Issue {
	if len(m.columns) == 0 || len(m.lists) == 0 {
		return nil
	}
	item := m.lists[m.activeColumn].SelectedItem()
	entry, ok := item.(drawerItem)
	if !ok {
		return nil
	}
	issue := entry.Entry.Issue
	return &issue
}
