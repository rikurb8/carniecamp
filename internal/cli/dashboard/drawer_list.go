package dashboard

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type drawerItem struct {
	Entry       drawerEntry
	Collapsed   bool
	HasChildren bool
	Prefix      string
}

func (d drawerItem) FilterValue() string {
	return fmt.Sprintf("%s %s", d.Entry.Issue.ID, d.Entry.Issue.Title)
}

type drawerItemDelegate struct {
	styles dashboardStyles
	height int
}

func newDrawerDelegate(styles dashboardStyles, height int) list.ItemDelegate {
	return &drawerItemDelegate{styles: styles, height: height}
}

func newDrawerList(delegate list.ItemDelegate) list.Model {
	model := list.New([]list.Item{}, delegate, 0, 0)
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetFilteringEnabled(false)
	model.DisableQuitKeybindings()
	return model
}

func (d *drawerItemDelegate) Height() int {
	if d.height < 1 {
		return 1
	}
	return d.height
}

func (d *drawerItemDelegate) Spacing() int {
	return 0
}

func (d *drawerItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d *drawerItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(drawerItem)
	if !ok {
		return
	}
	width := m.Width()
	lines := drawerItemLines(entry, width, d.styles)
	for len(lines) < d.Height() {
		lines = append(lines, "")
	}
	if len(lines) > d.Height() {
		lines = lines[:d.Height()]
	}
	line := strings.Join(lines, "\n")
	style := d.styles.drawerItem
	if entry.Entry.Issue.IssueType == "epic" {
		style = d.styles.drawerEpic
	}
	if index == m.Index() {
		style = d.styles.drawerItemSelected
	}
	_, _ = fmt.Fprint(w, style.Render(line))
}

func drawerItemLines(item drawerItem, width int, styles dashboardStyles) []string {
	if width <= 0 {
		return []string{""}
	}
	issue := item.Entry.Issue
	badge := renderDrawerBadges(issue, styles)
	badgeWidth := lipgloss.Width(badge)
	available := width
	if badge != "" {
		available = width - badgeWidth - 1
		if available < 1 {
			available = 1
		}
	}
	prefix := renderDrawerPrefix(item)
	prefixWidth := lipgloss.Width(prefix)
	title := issue.Title
	if issue.IssueType == "epic" && item.HasChildren && item.Collapsed {
		title = "◆ " + title
	}
	firstWidth := maxInt(1, available-prefixWidth)
	remainingWidth := maxInt(1, width-prefixWidth)
	titleLines := wrapTitle(title, firstWidth, remainingWidth)
	if len(titleLines) == 0 {
		titleLines = []string{""}
	}
	lines := make([]string, 0, len(titleLines))
	firstLeft := prefix + titleLines[0]
	lines = append(lines, renderLineWithBadge(firstLeft, badge, width))
	if len(titleLines) > 1 {
		indent := strings.Repeat(" ", prefixWidth)
		for _, line := range titleLines[1:] {
			lines = append(lines, lipgloss.NewStyle().Width(width).Render(indent+line))
		}
	}
	return lines
}

func drawerItemLineCount(item drawerItem, width int, styles dashboardStyles) int {
	return len(drawerItemLines(item, width, styles))
}

func renderDrawerPrefix(item drawerItem) string {
	issue := item.Entry.Issue
	indicator := ""
	if issue.IssueType == "epic" && item.HasChildren {
		if item.Collapsed {
			indicator = "▸ "
		} else {
			indicator = "▾ "
		}
	}
	return item.Prefix + indicator
}

func renderLineWithBadge(left string, badge string, width int) string {
	if badge == "" {
		return lipgloss.NewStyle().Width(width).Render(left)
	}
	gap := width - lipgloss.Width(left) - lipgloss.Width(badge)
	if gap < 1 {
		left = truncateASCII(left, maxInt(1, width-lipgloss.Width(badge)-1))
		gap = width - lipgloss.Width(left) - lipgloss.Width(badge)
		if gap < 1 {
			gap = 1
		}
	}
	return left + strings.Repeat(" ", gap) + badge
}

func renderDrawerBadges(issue Issue, styles dashboardStyles) string {
	badges := make([]string, 0, 2)
	if issue.Priority > 0 {
		badges = append(badges, styles.badgePriority.Render(fmt.Sprintf("P%d", issue.Priority)))
	}
	status := mapStatusBadge(issue.Status)
	if status != "" {
		style := styles.badgeDefault
		switch status {
		case "OPEN", "READY":
			style = styles.badgeReady
		case "WIP":
			style = styles.badgeProgress
		case "BLKD":
			style = styles.badgeBlocked
		case "DONE":
			style = styles.badgeClosed
		}
		badges = append(badges, style.Render(status))
	}
	if len(badges) == 0 {
		return ""
	}
	return strings.Join(badges, " ")
}

func mapStatusBadge(status string) string {
	switch status {
	case "open":
		return "OPEN"
	case "ready":
		return "READY"
	case "in_progress":
		return "WIP"
	case "blocked":
		return "BLKD"
	case "closed":
		return "DONE"
	default:
		return ""
	}
}
