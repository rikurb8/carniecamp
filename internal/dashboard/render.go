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
	body := renderDashboardBody(inner, styles)
	footer := renderDashboardFooter()

	output := lipgloss.JoinVertical(lipgloss.Left, navbar, body)
	if m.showHelp {
		output = renderHelpOverlay(output, inner, styles)
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

	// View tabs
	homeTab := "1 Home"
	tasksTab := "2 Tasks"
	if m.activeView == ViewHome {
		homeTab = styles.viewTabActive.Render(homeTab)
		tasksTab = styles.viewTabInactive.Render(tasksTab)
	} else {
		homeTab = styles.viewTabInactive.Render(homeTab)
		tasksTab = styles.viewTabActive.Render(tasksTab)
	}
	viewTabs := homeTab + " " + tasksTab

	updated := "Stand by..."
	if m.errMessage != "" {
		updated = m.errMessage
	} else if !m.lastUpdated.IsZero() {
		updated = fmt.Sprintf("Updated %s", m.lastUpdated.Format("15:04:05"))
	}
	rightSub := styles.navbarMeta.Render(updated)
	line := renderNavbarLine(width, viewTabs, rightSub, styles.navbarBar)

	return line
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
	if m.activeView == ViewHome {
		return renderHomeView(m, styles)
	}
	stats := renderDashboardStats(m, styles)
	tasks := renderMasterDetail(m, styles)
	return lipgloss.JoinVertical(lipgloss.Left, stats, tasks)
}

func renderHomeView(m Model, styles dashboardStyles) string {
	width := m.width
	height := m.height - 2
	if width <= 0 || height <= 0 {
		return ""
	}

	// Styles
	tentColor := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	goldColor := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	welcomeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	tipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	starStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)

	// Get spinner views
	getSpinner := func(i int) string {
		if i < len(m.homeSpinners) {
			return m.homeSpinners[i].View()
		}
		return "*"
	}

	// Build spinners row content with fixed structure
	s0, s1, s2, s3, s4, s5 := getSpinner(0), getSpinner(1), getSpinner(2), getSpinner(3), getSpinner(4), getSpinner(5)

	// ASCII tent art - plain text, will be colored
	tentLines := []string{
		"      *     +     *      ",
		"           /\\           ",
		"          /||\\          ",
		"         /||||\\         ",
		"        /||||||\\        ",
		"       /||||||||\\       ",
		"      /||||||||||\\      ",
		"     [||||||||||||]     ",
		"     [    |  |    ]     ",
		"     [----+--+----]     ",
		"                         ",
		"    STEP RIGHT UP!       ",
	}

	// Build left panel - use lipgloss box style
	var tentArtLines []string
	for i, line := range tentLines {
		si := i % 6
		ls := getSpinner(si)
		rs := getSpinner((si + 3) % 6)

		// Color the tent characters
		var colored strings.Builder
		for _, ch := range line {
			switch ch {
			case '*', '+':
				colored.WriteString(starStyle.Render(string(ch)))
			case '/', '\\', '[', ']', '|', '-':
				colored.WriteString(tentColor.Render(string(ch)))
			default:
				colored.WriteString(string(ch))
			}
		}
		tentArtLines = append(tentArtLines, ls+" "+colored.String()+" "+rs)
	}

	// Top and bottom spinner rows
	topSpinners := s0 + "  " + s1 + "  " + s2 + "  " + s3 + "  " + s4 + "  " + s5
	botSpinners := s5 + "  " + s4 + "  " + s3 + "  " + s2 + "  " + s1 + "  " + s0

	leftContent := []string{topSpinners, ""}
	leftContent = append(leftContent, tentArtLines...)
	leftContent = append(leftContent, "", botSpinners)

	leftBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("220")).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Center, leftContent...))

	// Build right panel: Welcome & Tips
	rightLines := []string{
		titleStyle.Render("╔═════════════════════════════╗"),
		titleStyle.Render("║") + welcomeStyle.Render("   WELCOME & QUICK TIPS   ") + titleStyle.Render("║"),
		titleStyle.Render("╚═════════════════════════════╝"),
		"",
		goldColor.Render("       * CARNIE CAMP *"),
		dimStyle.Render("  Your carnival of tasks awaits"),
		"",
		goldColor.Render("  ┌───────────────────────┐"),
		goldColor.Render("  │") + tipStyle.Render("     NAVIGATION       ") + goldColor.Render("│"),
		goldColor.Render("  ├───────────────────────┤"),
		goldColor.Render("  │") + tipStyle.Render(" ") + keyStyle.Render("1") + tipStyle.Render(" Home  ") + keyStyle.Render("2") + tipStyle.Render(" Tasks    ") + goldColor.Render("│"),
		goldColor.Render("  │") + tipStyle.Render(" ") + keyStyle.Render("q") + tipStyle.Render(" Quit  ") + keyStyle.Render("?") + tipStyle.Render(" Help     ") + goldColor.Render("│"),
		goldColor.Render("  └───────────────────────┘"),
		"",
		goldColor.Render("  ┌───────────────────────┐"),
		goldColor.Render("  │") + tipStyle.Render("    TASK CONTROLS     ") + goldColor.Render("│"),
		goldColor.Render("  ├───────────────────────┤"),
		goldColor.Render("  │") + tipStyle.Render(" ") + keyStyle.Render("j/k") + tipStyle.Render(" Move up/down   ") + goldColor.Render("│"),
		goldColor.Render("  │") + tipStyle.Render(" ") + keyStyle.Render("tab") + tipStyle.Render(" Switch columns ") + goldColor.Render("│"),
		goldColor.Render("  │") + tipStyle.Render(" ") + keyStyle.Render("</>") + tipStyle.Render(" Collapse/expand") + goldColor.Render("│"),
		goldColor.Render("  │") + tipStyle.Render(" ") + keyStyle.Render("r") + tipStyle.Render("   Refresh        ") + goldColor.Render("│"),
		goldColor.Render("  └───────────────────────┘"),
		"",
		starStyle.Render(" *") + dimStyle.Render(" The show is about to begin! ") + starStyle.Render("*"),
	}

	rightPanel := lipgloss.JoinVertical(lipgloss.Left, rightLines...)

	// Join panels
	combined := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "    ", rightPanel)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, combined)
}

func renderMasterDetail(m Model, styles dashboardStyles) string {
	width := m.width
	if width <= 0 {
		return ""
	}
	drawerWidth, bodyHeight, listHeight, showSummary := drawerLayout(m.width, m.height)
	gap := 2
	if drawerWidth+gap >= width {
		return renderActiveDrawer(m, width, bodyHeight, listHeight, showSummary, styles)
	}

	left := renderActiveDrawer(m, drawerWidth, bodyHeight, listHeight, showSummary, styles)
	right := renderDetailPanel(m, width-drawerWidth-gap, bodyHeight, styles)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
}

func renderActiveDrawer(m Model, width int, height int, listHeight int, showSummary bool, styles dashboardStyles) string {
	rows := make([]string, 0, height)
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	if len(m.columns) == 0 || len(m.lists) == 0 {
		return lipgloss.NewStyle().Width(width).Render("")
	}
	activeColumn := m.columns[m.activeColumn]
	activeList := m.lists[m.activeColumn]
	borderStyle := styles.paneBorderActive
	entries := buildDrawerEntries(activeColumn, m.collapsed)
	title := fmt.Sprintf("%s (%d)", activeColumn.Title, len(entries))
	rows = append(rows, renderPaneTopRule(innerWidth, borderStyle))
	if len(rows) < height {
		rows = append(rows, renderPaneHeaderLine(title, innerWidth, styles.columnTitle, borderStyle))
	}
	if showSummary && len(m.columns) > 1 && len(rows) < height {
		otherIndex := (m.activeColumn + 1) % len(m.columns)
		other := m.columns[otherIndex]
		otherCount := len(buildDrawerEntries(other, m.collapsed))
		summary := fmt.Sprintf("Other: %s (%d)", other.Title, otherCount)
		rows = append(rows, renderPaneHeaderLine(summary, innerWidth, styles.dimText, borderStyle))
	}

	activeList.SetSize(innerWidth, listHeight)
	listView := activeList.View()
	listLines := strings.Split(listView, "\n")
	for len(listLines) < listHeight {
		listLines = append(listLines, "")
	}
	if len(listLines) > listHeight {
		listLines = listLines[:listHeight]
	}
	for _, line := range listLines {
		rows = append(rows, renderPaneRawRow(line, innerWidth, borderStyle))
		if len(rows) >= height-1 {
			break
		}
	}
	if len(rows) < height {
		rows = append(rows, renderPaneBottomRule(innerWidth, borderStyle))
	}

	for len(rows) < height {
		rows = append(rows, "")
	}

	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func renderPaneTopRule(innerWidth int, borderStyle lipgloss.Style) string {
	if innerWidth < 1 {
		return ""
	}
	line := "┌" + strings.Repeat("─", innerWidth) + "┐"
	return borderStyle.Render(line)
}

func renderPaneBottomRule(innerWidth int, borderStyle lipgloss.Style) string {
	if innerWidth < 1 {
		return ""
	}
	line := "└" + strings.Repeat("─", innerWidth) + "┘"
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
	return borderStyle.Render("│") + content + borderStyle.Render("│")
}

func renderPaneHeaderLine(text string, innerWidth int, textStyle lipgloss.Style, borderStyle lipgloss.Style) string {
	text = truncateASCII(text, innerWidth)
	pad := innerWidth - len(text)
	if pad < 0 {
		pad = 0
	}
	content := textStyle.Render(text) + strings.Repeat(" ", pad)
	return borderStyle.Render("│") + content + borderStyle.Render("│")
}

func renderPaneRow(text string, innerWidth int, textStyle lipgloss.Style, borderStyle lipgloss.Style) string {
	text = truncateASCII(text, innerWidth)
	pad := innerWidth - len(text)
	if pad < 0 {
		pad = 0
	}
	content := textStyle.Render(text) + strings.Repeat(" ", pad)
	return borderStyle.Render("│") + content + borderStyle.Render("│")
}

func renderPaneRawRow(text string, innerWidth int, borderStyle lipgloss.Style) string {
	content := lipgloss.NewStyle().Width(innerWidth).Render(text)
	return borderStyle.Render("│") + content + borderStyle.Render("│")
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

	children := m.featureChildren[selected.ID]
	lines = append(lines, "")
	if len(children) == 0 {
		lines = append(lines, styles.panelTitle.Render(truncateASCII("Child Tasks", width)))
		lines = append(lines, styles.dimText.Render(truncateASCII("(none)", width)))
	} else {
		header := fmt.Sprintf("Child Tasks (%d)", len(children))
		lines = append(lines, styles.panelTitle.Render(truncateASCII(header, width)))
		for _, child := range children {
			status := mapStatusBadge(child.Status)
			if status == "" {
				status = strings.ToUpper(child.Status)
			}
			line := fmt.Sprintf("P%d %s %s %s", child.Priority, status, child.ID, child.Title)
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

func renderDashboardFooter() string {
	return "1/2 views  h/? help  tab switch  j/k move  left/right collapse  r refresh  q quit"
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

	keysLine := fmt.Sprintf("%-6s %s", "Keys:", "q quit  tab switch section  j/k move  left/right collapse  r refresh  h/? close")
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
