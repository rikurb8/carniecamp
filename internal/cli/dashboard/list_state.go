package dashboard

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

func (m *Model) applyDrawerLayout() {
	if len(m.lists) == 0 {
		return
	}
	drawerWidth, _, listHeight, _ := drawerLayout(m.width, m.height)
	innerWidth := maxInt(1, drawerWidth-2)
	m.refreshDrawerLists(true, innerWidth)
	for idx := range m.lists {
		m.lists[idx].SetSize(innerWidth, listHeight)
	}
}

func (m *Model) refreshDrawerLists(keepSelection bool, innerWidth int) {
	for idx := range m.lists {
		m.refreshDrawerList(idx, keepSelection, innerWidth)
	}
}

func (m *Model) refreshDrawerList(index int, keepSelection bool, innerWidth int) {
	if index < 0 || index >= len(m.lists) || index >= len(m.columns) {
		return
	}
	model := m.lists[index]
	selectedID := ""
	if keepSelection {
		selectedID = selectedIssueIDFromList(model)
	}
	items := buildDrawerItems(m.columns[index], m.collapsed)
	model.SetItems(items)
	styles := newDashboardStyles()
	delegateHeight := maxInt(1, maxDrawerItemHeight(items, innerWidth, styles))
	model.SetDelegate(newDrawerDelegate(styles, delegateHeight))
	if selectedID != "" {
		for i, item := range items {
			entry, ok := item.(drawerItem)
			if ok && entry.Entry.Issue.ID == selectedID {
				model.Select(i)
				break
			}
		}
	} else if len(items) > 0 {
		model.Select(0)
	}
	m.lists[index] = model
}

func buildDrawerItems(column issueColumn, collapsed map[string]bool) []list.Item {
	entries := buildDrawerEntries(column, collapsed)
	parentByID := column.ParentByID
	if parentByID == nil {
		parentByID = map[string]string{}
	}
	childCounts := make(map[string]int, len(column.Issues))
	for _, issue := range column.Issues {
		parentID := parentByID[issue.ID]
		childCounts[parentID]++
	}
	seenByParent := make(map[string]int)
	isLastByID := make(map[string]bool, len(entries))
	for _, entry := range entries {
		parentID := parentByID[entry.Issue.ID]
		seenByParent[parentID]++
		isLastByID[entry.Issue.ID] = seenByParent[parentID] >= childCounts[parentID]
	}
	items := make([]list.Item, 0, len(entries))
	for _, entry := range entries {
		hasChildren := childCounts[entry.Issue.ID] > 0
		prefix := buildTreePrefix(entry.Issue.ID, entry.Level, parentByID, isLastByID)
		items = append(items, drawerItem{Entry: entry, Collapsed: collapsed[entry.Issue.ID], HasChildren: hasChildren, Prefix: prefix})
	}
	return items
}

func selectedIssueIDFromList(model list.Model) string {
	item := model.SelectedItem()
	entry, ok := item.(drawerItem)
	if !ok {
		return ""
	}
	return entry.Entry.Issue.ID
}

func buildTreePrefix(id string, level int, parentByID map[string]string, isLastByID map[string]bool) string {
	if level <= 0 {
		return ""
	}
	ancestors := make([]string, 0, level)
	parentID := parentByID[id]
	for parentID != "" {
		ancestors = append(ancestors, parentID)
		parentID = parentByID[parentID]
	}
	var b strings.Builder
	if len(ancestors) > 1 {
		for i := len(ancestors) - 1; i > 0; i-- {
			ancestorID := ancestors[i]
			if isLastByID[ancestorID] {
				b.WriteString("   ")
			} else {
				b.WriteString("│  ")
			}
		}
	}
	if isLastByID[id] {
		b.WriteString("└─ ")
	} else {
		b.WriteString("├─ ")
	}
	return b.String()
}

func maxDrawerItemHeight(items []list.Item, width int, styles dashboardStyles) int {
	maxLines := 1
	if width <= 0 {
		return maxLines
	}
	for _, item := range items {
		entry, ok := item.(drawerItem)
		if !ok {
			continue
		}
		lines := drawerItemLineCount(entry, width, styles)
		if lines > maxLines {
			maxLines = lines
		}
	}
	return maxLines
}
