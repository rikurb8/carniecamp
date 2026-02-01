package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading dashboard..."
	}

	styles := newDashboardStyles()
	inner := m
	if m.width > 2 && m.height > 3 {
		inner.width = m.width - 2
		inner.height = m.height - 3
	}

	navbar := renderDashboardNavbar(inner, styles)
	stats := renderDashboardStats(inner, styles)
	body := renderDashboardBody(inner, styles)
	footer := renderDashboardFooter()

	output := lipgloss.JoinVertical(lipgloss.Left, navbar, stats, body)
	if m.showHelp {
		output = renderHelpOverlay(output, inner, styles)
	}
	if m.showImprover {
		output = renderImproverOverlay(output, inner, styles)
	}
	if m.width > 2 && m.height > 2 {
		return renderTentFrame(output, footer, styles, m.width, m.height)
	}
	if footer != "" {
		return lipgloss.JoinVertical(lipgloss.Left, output, styles.footer.Render(footer))
	}
	return output
}

func renderDashboardNavbar(m Model, styles dashboardStyles) string {
	width := m.width
	if width <= 0 {
		return ""
	}

	leftSubText := ""
	updated := "Stand by..."
	if m.errMessage != "" {
		updated = m.errMessage
	} else if !m.lastUpdated.IsZero() {
		updated = fmt.Sprintf("Updated %s", m.lastUpdated.Format("15:04:05"))
	}
	leftSubText, updated = clampNavbarSegments(width, leftSubText, updated)
	leftSub := styles.navbarSub.Render(leftSubText)
	rightSub := styles.navbarMeta.Render(updated)
	line2 := renderNavbarLine(width, leftSub, rightSub, styles.navbarBar)

	return line2
}

func renderNavbarLine(width int, left string, right string, style lipgloss.Style) string {
	if width <= 0 {
		return ""
	}
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + right
	return style.Width(width).Render(line)
}

func clampNavbarSegments(width int, left string, right string) (string, string) {
	if width <= 0 {
		return left, right
	}
	maxRight := width / 3
	if maxRight < 10 {
		maxRight = minInt(10, width)
	}
	if maxRight > width-1 {
		maxRight = width - 1
	}
	maxLeft := width - maxRight - 1
	if maxLeft < 5 {
		maxLeft = width - 1
	}
	return truncateASCII(left, maxLeft), truncateASCII(right, maxRight)
}

func renderDashboardStats(m Model, styles dashboardStyles) string {
	if m.lastUpdated.IsZero() {
		return styles.dimText.Render("Loading beads data...")
	}
	status := m.summary
	tags := []string{
		styles.tag.Render(fmt.Sprintf("Total %d", status.TotalIssues)),
		styles.tag.Render(fmt.Sprintf("Open %d", status.OpenIssues)),
		styles.tag.Render(fmt.Sprintf("Ready %d", status.ReadyIssues)),
		styles.tag.Render(fmt.Sprintf("In Progress %d", status.InProgressIssues)),
		styles.tag.Render(fmt.Sprintf("Blocked %d", status.BlockedIssues)),
		styles.tag.Render(fmt.Sprintf("Deferred %d", status.DeferredIssues)),
		styles.tag.Render(fmt.Sprintf("Closed %d", status.ClosedIssues)),
	}
	statsLine := styles.subheader.Render(lipgloss.JoinHorizontal(lipgloss.Left, tags...))
	return lipgloss.JoinVertical(lipgloss.Left, statsLine)
}

func renderDashboardBody(m Model, styles dashboardStyles) string {
	return renderMasterDetail(m, styles)
}

func renderMasterDetail(m Model, styles dashboardStyles) string {
	width := m.width
	if width <= 0 {
		return ""
	}
	height := availableListHeight(m.height)
	if height < 3 {
		height = 3
	}

	minDrawer := 26
	maxDrawer := 40
	drawerWidth := width / 3
	if drawerWidth < minDrawer {
		drawerWidth = minDrawer
	}
	if drawerWidth > maxDrawer {
		drawerWidth = maxDrawer
	}
	if drawerWidth > width-12 {
		drawerWidth = width - 12
	}
	if drawerWidth < 20 {
		drawerWidth = minInt(20, width)
	}
	gap := 2
	if drawerWidth+gap >= width {
		return renderDrawer(m, width, height, styles)
	}

	left := renderDrawer(m, drawerWidth, height, styles)
	right := renderDetailPanel(m, width-drawerWidth-gap, height, styles)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
}

func renderDrawer(m Model, width int, height int, styles dashboardStyles) string {
	rows := make([]string, 0, height)
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	for colIndex, column := range m.columns {
		active := colIndex == m.activeColumn
		entries := buildDrawerEntries(column, m.collapsed)
		title := fmt.Sprintf("%s (%d)", column.Title, len(entries))
		headerStyle := styles.columnHeader
		borderStyle := styles.paneBorder
		if active {
			headerStyle = styles.columnTitle
			borderStyle = styles.paneBorderActive
		}
		rows = append(rows, renderPaneRule(innerWidth, borderStyle))
		if len(rows) >= height {
			break
		}
		rows = append(rows, renderPaneHeaderRow(title, innerWidth, headerStyle, borderStyle))
		if len(rows) >= height {
			break
		}
		rows = append(rows, renderPaneRow("", innerWidth, styles.item, borderStyle))
		if len(rows) >= height {
			break
		}

		if len(entries) == 0 {
			rows = append(rows, renderPaneRow("(none)", innerWidth, styles.dimText, borderStyle))
			if len(rows) < height {
				rows = append(rows, renderPaneRule(innerWidth, borderStyle))
			}
			continue
		}

		start := column.Offset
		if start < 0 {
			start = 0
		}
		visible := height - len(rows) - 1
		if visible < 1 {
			rows = append(rows, renderPaneRule(innerWidth, borderStyle))
			break
		}
		end := start + visible
		if end > len(entries) {
			end = len(entries)
		}

		for i := start; i < end; i++ {
			entry := entries[i]
			issue := entry.Issue
			line := fmt.Sprintf("%s %s", issue.ID, issue.Title)
			if issue.IssueType == "epic" {
				indicator := "[-]"
				if m.collapsed[issue.ID] {
					indicator = "[+]"
				}
				line = indicator + " EPIC " + line
			}
			if entry.Level > 0 {
				line = strings.Repeat("  ", entry.Level) + "- " + line
			}
			line = truncateASCII(line, innerWidth)
			style := styles.item
			if issue.IssueType == "epic" {
				style = styles.panelTitle
			}
			if i == column.Selected {
				if active {
					style = styles.itemSelected
				} else {
					style = styles.itemSelectedIn
				}
			}
			rows = append(rows, renderPaneRow(line, innerWidth, style, borderStyle))
			if len(rows) >= height-1 {
				break
			}
		}
		if len(rows) < height {
			rows = append(rows, renderPaneRule(innerWidth, borderStyle))
		}
		if len(rows) >= height {
			break
		}
	}

	for len(rows) < height {
		rows = append(rows, "")
	}

	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func renderPaneRule(innerWidth int, borderStyle lipgloss.Style) string {
	if innerWidth < 1 {
		return ""
	}
	line := "+" + strings.Repeat("-", innerWidth) + "+"
	return borderStyle.Render(line)
}

func renderPaneHeaderRow(text string, innerWidth int, textStyle lipgloss.Style, borderStyle lipgloss.Style) string {
	text = truncateASCII(text, innerWidth)
	pad := innerWidth - len(text)
	if pad < 0 {
		pad = 0
	}
	leftPad := pad / 2
	rightPad := pad - leftPad
	content := strings.Repeat(" ", leftPad) + textStyle.Render(text) + strings.Repeat(" ", rightPad)
	return borderStyle.Render("|") + content + borderStyle.Render("|")
}

func renderPaneRow(text string, innerWidth int, textStyle lipgloss.Style, borderStyle lipgloss.Style) string {
	text = truncateASCII(text, innerWidth)
	pad := innerWidth - len(text)
	if pad < 0 {
		pad = 0
	}
	content := textStyle.Render(text) + strings.Repeat(" ", pad)
	return borderStyle.Render("|") + content + borderStyle.Render("|")
}

func renderDetailPanel(m Model, width int, height int, styles dashboardStyles) string {
	selected := m.selectedIssue()
	lines := []string{}
	if selected == nil {
		lines = append(lines, styles.dimText.Render(truncateASCII("Select an issue to see details.", width)))
		return renderPanel("Issue Details", lines, width, height, styles)
	}

	lines = append(lines, styles.panelTitle.Render(truncateASCII(fmt.Sprintf("%s %s", selected.ID, selected.Title), width)))
	lines = append(lines, styles.dimText.Render(truncateASCII(fmt.Sprintf("Status: %s", selected.Status), width)))
	lines = append(lines, styles.dimText.Render(truncateASCII(fmt.Sprintf("Priority: P%d", selected.Priority), width)))
	if selected.Owner != "" {
		lines = append(lines, styles.dimText.Render(truncateASCII(fmt.Sprintf("Owner: %s", selected.Owner), width)))
	}
	if selected.UpdatedAt != "" {
		lines = append(lines, styles.dimText.Render(truncateASCII(fmt.Sprintf("Updated: %s", formatTimestamp(selected.UpdatedAt)), width)))
	}
	lines = append(lines, "")
	if selected.Description != "" {
		lines = append(lines, styles.panelTitle.Render(truncateASCII("Notes", width)))
		for _, line := range wrapLines(selected.Description, width) {
			lines = append(lines, styles.item.Render(truncateASCII(line, width)))
		}
	}

	return renderPanel("Issue Details", lines, width, height, styles)
}

func renderIssuePreviewLines(title string, issues []Issue, limit int, width int, styles dashboardStyles) []string {
	lines := []string{styles.panelTitle.Render(truncateASCII(title, width))}
	if len(issues) == 0 {
		lines = append(lines, styles.dimText.Render(truncateASCII("(none)", width)))
		return lines
	}
	if limit <= 0 {
		return lines
	}
	if limit > len(issues) {
		limit = len(issues)
	}
	for i := 0; i < limit; i++ {
		issue := issues[i]
		line := fmt.Sprintf("P%d %s %s", issue.Priority, issue.ID, issue.Title)
		lines = append(lines, styles.item.Render(truncateASCII(line, width)))
	}
	return lines
}

func renderPanel(title string, lines []string, width int, height int, styles dashboardStyles) string {
	if width <= 0 {
		return ""
	}
	rows := make([]string, 0, height)
	rows = append(rows, styles.panelTitle.Render(truncateASCII(title, width)))

	bodyHeight := height - 1
	if bodyHeight < 0 {
		bodyHeight = 0
	}
	for i := 0; i < bodyHeight && i < len(lines); i++ {
		rows = append(rows, lines[i])
	}
	for len(rows) < height {
		rows = append(rows, "")
	}
	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func renderColumn(column issueColumn, width int, height int, active bool, styles dashboardStyles) string {
	if width <= 0 {
		return ""
	}

	title := fmt.Sprintf("%s (%d)", column.Title, len(column.Issues))
	titleStyle := styles.columnTitle
	if !active {
		titleStyle = styles.columnTitleDim
	}

	rows := make([]string, 0, height+1)
	rows = append(rows, titleStyle.Render(truncateASCII(title, width)))

	bodyHeight := height - 1
	if bodyHeight < 0 {
		bodyHeight = 0
	}

	if len(column.Issues) == 0 {
		for i := 0; i < bodyHeight; i++ {
			rows = append(rows, styles.dimText.Render(truncateASCII("(no issues)", width)))
		}
		return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
	}

	start := column.Offset
	if start < 0 {
		start = 0
	}
	end := start + bodyHeight
	if end > len(column.Issues) {
		end = len(column.Issues)
	}

	for i := start; i < end; i++ {
		issue := column.Issues[i]
		line := fmt.Sprintf("P%d %s %s", issue.Priority, issue.ID, issue.Title)
		line = truncateASCII(line, width)

		style := styles.item
		if i == column.Selected {
			if active {
				style = styles.itemSelected
			} else {
				style = styles.itemSelectedIn
			}
		}

		rows = append(rows, style.Render(line))
	}

	for len(rows) < height {
		rows = append(rows, "")
	}

	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func renderDashboardFooter() string {
	return "h/? help  o improve  tab switch  j/k move  left/right collapse  r refresh  q quit"
}

func renderHelpOverlay(base string, m Model, styles dashboardStyles) string {
	width := m.width
	height := m.height
	if width <= 0 || height <= 0 {
		return base
	}

	dialogWidth := minInt(width-6, 70)
	if dialogWidth < 40 {
		dialogWidth = minInt(width-2, 40)
	}
	if dialogWidth < 20 {
		dialogWidth = width
	}

	keysLine := fmt.Sprintf("%-6s %s", "Keys:", "q quit  o improve  tab switch section  j/k move  left/right collapse  r refresh  h/? close")
	tipsLine := fmt.Sprintf("%-6s %s", "Tips:", "Use left/right to fold epics; tab switches Future/Completed; j/k moves selection.")
	help := []string{
		styles.helpTitle.Render("Dashboard Help"),
		styles.helpText.Render(keysLine),
		styles.helpText.Render(tipsLine),
		"",
	}

	box := styles.helpBox.Width(dialogWidth).Render(lipgloss.JoinVertical(lipgloss.Left, help...))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func renderImproverOverlay(base string, m Model, styles dashboardStyles) string {
	width := m.width
	height := m.height
	if width <= 0 || height <= 0 {
		return base
	}

	dialogWidth := minInt(width-6, 80)
	if dialogWidth < 50 {
		dialogWidth = minInt(width-2, 50)
	}
	if dialogWidth < 20 {
		dialogWidth = width
	}
	selected := m.selectedIssue()
	issueLine := "No issue selected"
	if selected != nil {
		issueLine = fmt.Sprintf("%s %s", selected.ID, selected.Title)
	}
	lines := []string{
		styles.helpTitle.Render("Epic Improver / Task Splitter"),
		styles.dimText.Render("Enter to launch opencode  |  esc close"),
		"",
		styles.panelTitle.Render("Active Issue"),
		styles.helpText.Render(truncateASCII(issueLine, dialogWidth-4)),
		"",
		styles.panelTitle.Render("Optional Instructions"),
		m.improveInput.View(),
	}

	box := styles.helpBox.Width(dialogWidth).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) applyImproverLayout() {
	width := m.width
	if width <= 0 {
		return
	}
	dialogWidth := minInt(width-6, 80)
	if dialogWidth < 50 {
		dialogWidth = minInt(width-2, 50)
	}
	if dialogWidth < 20 {
		dialogWidth = width
	}
	innerWidth := dialogWidth - 4
	if innerWidth < 10 {
		innerWidth = 10
	}
	m.improveInput.Width = innerWidth
}

func renderTentFrame(content string, footer string, styles dashboardStyles, width int, height int) string {
	frameHeight := 3
	if width < 4 || height <= frameHeight {
		return content
	}

	innerWidth := width - 2
	innerHeight := height - frameHeight
	inner := lipgloss.NewStyle().Width(innerWidth).Height(innerHeight).Render(content)
	lines := strings.Split(inner, "\n")
	if len(lines) < innerHeight {
		for len(lines) < innerHeight {
			lines = append(lines, "")
		}
	}
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	top := styles.tentTop.Render("/" + strings.Repeat("^", width-2) + "\\")
	stripe := renderTentStripeLine(styles, width)
	base := styles.tentBase.Render(renderTentBaseLine(footer, width))
	rows := make([]string, 0, height)
	rows = append(rows, top)
	rows = append(rows, stripe)
	for _, line := range lines {
		rows = append(rows, styles.tentPost.Render("|")+line+styles.tentPost.Render("|"))
	}
	rows = append(rows, base)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderTentBaseLine(footer string, width int) string {
	if width < 2 {
		return ""
	}
	inner := width - 2
	text := truncateASCII(footer, inner)
	if text == "" {
		text = strings.Repeat("_", inner)
		return "\\" + text + "/"
	}
	pad := inner - len(text)
	if pad < 0 {
		pad = 0
	}
	return "\\" + text + strings.Repeat("_", pad) + "/"
}

func renderTentStripeLine(styles dashboardStyles, width int) string {
	if width < 2 {
		return ""
	}
	inner := width - 2
	title := " CARNIE CAMP "
	if inner <= len(title) {
		title = ""
	}
	var b strings.Builder
	b.WriteString(styles.tentPost.Render("|"))
	start := -1
	end := -1
	if title != "" {
		start = (inner - len(title)) / 2
		end = start + len(title)
	}
	for i := 0; i < inner; i++ {
		if i >= start && i < end {
			b.WriteString(styles.tentStripeB.Render(string(title[i-start])))
			continue
		}
		if i%2 == 0 {
			b.WriteString(styles.tentStripeA.Render("^"))
			continue
		}
		b.WriteString(styles.tentStripeB.Render("~"))
	}
	b.WriteString(styles.tentPost.Render("|"))
	return b.String()
}
