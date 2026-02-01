package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rikurb8/carnie/internal/beads"
	"github.com/spf13/cobra"
)

type bdIssue struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Status       string         `json:"status"`
	Priority     int            `json:"priority"`
	IssueType    string         `json:"issue_type"`
	Owner        string         `json:"owner"`
	UpdatedAt    string         `json:"updated_at"`
	CreatedAt    string         `json:"created_at"`
	Dependencies []bdDependency `json:"dependencies,omitempty"`
}

type bdDependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
}

type dashboardData struct {
	Status     bdStatus
	Ready      []bdIssue
	InProgress []bdIssue
	Blocked    []bdIssue
	Closed     []bdIssue
}

type dashboardDataMsg struct {
	Data dashboardData
	Err  error
}

type dashboardTickMsg struct{}

type issueColumn struct {
	Title      string
	Issues     []bdIssue
	Selected   int
	Offset     int
	ParentByID map[string]string
	LevelByID  map[string]int
}

type dashboardModel struct {
	width        int
	height       int
	columns      []issueColumn
	activeColumn int
	refresh      time.Duration
	limit        int
	lastUpdated  time.Time
	summary      bdStatusSummary
	errMessage   string
	showHelp     bool
	collapsed    map[string]bool
}

type dashboardStyles struct {
	header         lipgloss.Style
	welcome        lipgloss.Style
	subheader      lipgloss.Style
	tag            lipgloss.Style
	columnTitle    lipgloss.Style
	columnTitleDim lipgloss.Style
	panelTitle     lipgloss.Style
	item           lipgloss.Style
	itemSelected   lipgloss.Style
	itemSelectedIn lipgloss.Style
	footer         lipgloss.Style
	helpBox        lipgloss.Style
	helpTitle      lipgloss.Style
	helpText       lipgloss.Style
	dimText        lipgloss.Style
	errorText      lipgloss.Style
	navbarBar      lipgloss.Style
	navbarTitle    lipgloss.Style
	navbarMeta     lipgloss.Style
	navbarSub      lipgloss.Style
	tentTop        lipgloss.Style
	tentPost       lipgloss.Style
	tentBase       lipgloss.Style
	tentStripeA    lipgloss.Style
	tentStripeB    lipgloss.Style
}

type drawerEntry struct {
	Issue bdIssue
	Level int
}

func newDashboardCommand() *cobra.Command {
	var refresh time.Duration
	var limit int

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Launch the Carnie dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			model := newDashboardModel(refresh, limit)
			program := tea.NewProgram(model, tea.WithAltScreen())
			_, err := program.Run()
			return err
		},
	}

	cmd.Flags().DurationVar(&refresh, "refresh", 6*time.Second, "Auto-refresh interval (0 to disable)")
	cmd.Flags().IntVar(&limit, "limit", 200, "Max issues per column (0 for unlimited)")

	return cmd
}

func newDashboardModel(refresh time.Duration, limit int) dashboardModel {
	return dashboardModel{
		refresh:   refresh,
		limit:     limit,
		collapsed: map[string]bool{},
		columns: []issueColumn{
			{Title: "Future Work"},
			{Title: "Completed Work"},
		},
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return tea.Batch(loadDashboardDataCmd(m.limit), m.tickCmd())
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		m.ensureVisible()
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
	case dashboardTickMsg:
		return m, tea.Batch(loadDashboardDataCmd(m.limit), m.tickCmd())
	case dashboardDataMsg:
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
	}

	return m, nil
}

func (m dashboardModel) View() string {
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
	header := renderDashboardHeader(styles)
	stats := renderDashboardStats(inner, styles)
	body := renderDashboardBody(inner, styles)
	footer := renderDashboardFooter(inner, styles)

	output := lipgloss.JoinVertical(lipgloss.Left, navbar, header, stats, body, footer)
	if m.showHelp {
		output = renderHelpOverlay(output, inner, styles)
	}
	if m.width > 2 && m.height > 2 {
		return renderTentFrame(output, styles, m.width, m.height)
	}
	return output
}

func (m dashboardModel) tickCmd() tea.Cmd {
	if m.refresh <= 0 {
		return nil
	}
	return tea.Tick(m.refresh, func(time.Time) tea.Msg {
		return dashboardTickMsg{}
	})
}

func (m *dashboardModel) moveSelection(delta int) {
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

func (m *dashboardModel) setCollapseForSelected(collapse bool) {
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

func (m *dashboardModel) ensureVisible() {
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

func (m *dashboardModel) updateColumns(data dashboardData) {
	previous := make(map[string]string, len(m.columns))
	for _, column := range m.columns {
		selectedID := selectedIssueID(column, m.collapsed)
		if selectedID == "" {
			continue
		}
		previous[column.Title] = selectedID
	}

	future := append([]bdIssue{}, data.Ready...)
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

func orderIssuesWithParents(issues []bdIssue) ([]bdIssue, map[string]string, map[string]int) {
	ordered := make([]bdIssue, 0, len(issues))
	parentByID := make(map[string]string)
	levelByID := make(map[string]int)
	added := make(map[string]bool)

	issueByID := make(map[string]bdIssue, len(issues))
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

func renderDashboardHeader(styles dashboardStyles) string {
	welcome := styles.welcome.Render("Welcome to Carnie Camp")
	return welcome
}

func renderDashboardNavbar(m dashboardModel, styles dashboardStyles) string {
	width := m.width
	if width <= 0 {
		return ""
	}

	leftText := " CARNIE CAMP "
	meta := "Beads loading..."
	if !m.lastUpdated.IsZero() {
		meta = fmt.Sprintf("Beads %d", m.summary.TotalIssues)
	}
	leftText, meta = clampNavbarSegments(width, leftText, meta)
	left := styles.navbarTitle.Render(leftText)
	right := styles.navbarMeta.Render(meta)
	line1 := renderNavbarLine(width, left, right, styles.navbarBar)

	leftSubText := " Layout: Master-Detail  |  tab switch  j/k move  left/right collapse  r refresh "
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

	return lipgloss.JoinVertical(lipgloss.Left, line1, line2)
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

func renderDashboardStats(m dashboardModel, styles dashboardStyles) string {
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
	viewLine := styles.dimText.Render("Layout: Master-Detail")
	return lipgloss.JoinVertical(lipgloss.Left, statsLine, viewLine)
}

func renderDashboardColumns(m dashboardModel, styles dashboardStyles) string {
	gap := 2
	columnCount := len(m.columns)
	width := m.width
	if width <= 0 || columnCount == 0 {
		return ""
	}

	available := width - gap*(columnCount-1)
	if available < columnCount {
		available = columnCount
	}
	columnWidth := available / columnCount
	height := availableListHeight(m.height)

	rendered := make([]string, 0, columnCount)
	for idx, column := range m.columns {
		active := idx == m.activeColumn
		rendered = append(rendered, renderColumn(column, columnWidth, height, active, styles))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, joinWithGap(gap, rendered)...)
}

func renderDashboardBody(m dashboardModel, styles dashboardStyles) string {
	return renderMasterDetail(m, styles)
}

func renderMasterDetail(m dashboardModel, styles dashboardStyles) string {
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

func renderDrawer(m dashboardModel, width int, height int, styles dashboardStyles) string {
	rows := make([]string, 0, height)
	for colIndex, column := range m.columns {
		active := colIndex == m.activeColumn
		entries := buildDrawerEntries(column, m.collapsed)
		title := fmt.Sprintf("%s (%d)", column.Title, len(entries))
		titleStyle := styles.columnTitleDim
		if active {
			titleStyle = styles.columnTitle
		}
		rows = append(rows, titleStyle.Render(truncateASCII(title, width)))

		if len(entries) == 0 {
			rows = append(rows, styles.dimText.Render(truncateASCII("(none)", width)))
			continue
		}

		start := column.Offset
		if start < 0 {
			start = 0
		}
		visible := height - len(rows)
		if visible < 1 {
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
			line = truncateASCII(line, width)
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
			rows = append(rows, style.Render(line))
			if len(rows) >= height {
				break
			}
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
	issueByID := make(map[string]bdIssue, len(column.Issues))
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

func isHiddenByCollapse(id string, parentByID map[string]string, issueByID map[string]bdIssue, collapsed map[string]bool) bool {
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

func renderDetailPanel(m dashboardModel, width int, height int, styles dashboardStyles) string {
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

func renderIssuePreviewLines(title string, issues []bdIssue, limit int, width int, styles dashboardStyles) []string {
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

func renderDashboardFooter(m dashboardModel, styles dashboardStyles) string {
	hint := "h help  tab switch  j/k move  left/right collapse  r refresh  q quit"
	updated := ""
	updatedStyle := styles.dimText
	if !m.lastUpdated.IsZero() {
		updated = fmt.Sprintf("Updated %s", m.lastUpdated.Format("15:04:05"))
	}
	if m.errMessage != "" {
		updated = truncateASCII(m.errMessage, m.width)
		updatedStyle = styles.errorText
	}

	line := lipgloss.JoinHorizontal(lipgloss.Left, styles.footer.Render(hint), "  ", updatedStyle.Render(updated))
	return line
}

func renderHelpOverlay(base string, m dashboardModel, styles dashboardStyles) string {
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

	help := []string{
		styles.helpTitle.Render("Dashboard Help"),
		styles.helpText.Render("Keys: q quit  tab switch section  j/k move  left/right collapse  r refresh  h close"),
		styles.helpText.Render("Layout: Master-Detail"),
		"",
	}

	selected := m.selectedIssue()
	if selected != nil {
		help = append(help,
			styles.helpTitle.Render("Selected Issue"),
			styles.helpText.Render(fmt.Sprintf("ID: %s", selected.ID)),
			styles.helpText.Render(fmt.Sprintf("Title: %s", selected.Title)),
			styles.helpText.Render(fmt.Sprintf("Status: %s", selected.Status)),
			styles.helpText.Render(fmt.Sprintf("Priority: P%d", selected.Priority)),
			styles.helpText.Render(fmt.Sprintf("Owner: %s", selected.Owner)),
		)
		if selected.UpdatedAt != "" {
			help = append(help, styles.helpText.Render(fmt.Sprintf("Updated: %s", formatTimestamp(selected.UpdatedAt))))
		}
		if selected.Description != "" {
			help = append(help, "", styles.helpText.Render("Notes:"))
			for _, line := range wrapLines(selected.Description, dialogWidth-4) {
				help = append(help, styles.helpText.Render(line))
			}
		}
	} else {
		help = append(help, styles.helpText.Render("No issue selected."))
	}

	box := styles.helpBox.Width(dialogWidth).Render(lipgloss.JoinVertical(lipgloss.Left, help...))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func renderTentFrame(content string, styles dashboardStyles, width int, height int) string {
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
	base := styles.tentBase.Render("\\" + strings.Repeat("_", width-2) + "/")
	rows := make([]string, 0, height)
	rows = append(rows, top)
	rows = append(rows, stripe)
	for _, line := range lines {
		rows = append(rows, styles.tentPost.Render("|")+line+styles.tentPost.Render("|"))
	}
	rows = append(rows, base)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderTentStripeLine(styles dashboardStyles, width int) string {
	if width < 2 {
		return ""
	}
	inner := width - 2
	var b strings.Builder
	b.WriteString(styles.tentPost.Render("|"))
	for i := 0; i < inner; i++ {
		if i%2 == 0 {
			b.WriteString(styles.tentStripeA.Render("^"))
			continue
		}
		b.WriteString(styles.tentStripeB.Render("~"))
	}
	b.WriteString(styles.tentPost.Render("|"))
	return b.String()
}

func (m dashboardModel) selectedIssue() *bdIssue {
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

func loadDashboardDataCmd(limit int) tea.Cmd {
	return func() tea.Msg {
		data, err := fetchDashboardData(limit)
		return dashboardDataMsg{Data: data, Err: err}
	}
}

func fetchDashboardData(limit int) (dashboardData, error) {
	statusOutput, err := runBdJSON("status", "--json")
	if err != nil {
		return dashboardData{}, err
	}

	var status bdStatus
	if err := json.Unmarshal(statusOutput, &status); err != nil {
		return dashboardData{}, fmt.Errorf("parse bd status: %w", err)
	}

	ready, inProgress, blocked, closed, err := fetchIssuesFromBeads(limit)
	if err != nil {
		return dashboardData{}, err
	}

	return dashboardData{
		Status:     status,
		Ready:      ready,
		InProgress: inProgress,
		Blocked:    blocked,
		Closed:     closed,
	}, nil
}

func fetchIssuesFromBeads(limit int) ([]bdIssue, []bdIssue, []bdIssue, []bdIssue, error) {
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

func mapBeadsIssues(issues []beads.Issue) []bdIssue {
	output := make([]bdIssue, 0, len(issues))
	for _, issue := range issues {
		deps := make([]bdDependency, 0, len(issue.Dependencies))
		for _, dep := range issue.Dependencies {
			deps = append(deps, bdDependency{IssueID: dep.IssueID, DependsOnID: dep.DependsOnID, Type: dep.Type})
		}
		output = append(output, bdIssue{
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

func fetchIssues(args []string, limit int) ([]bdIssue, error) {
	if limit > 0 {
		args = append(args, "--limit", strconv.Itoa(limit))
	}
	output, err := runBdJSON(args...)
	if err != nil {
		return nil, err
	}
	return parseBdIssues(output)
}

func parseBdIssues(data []byte) ([]bdIssue, error) {
	var issues []bdIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("parse bd issues: %w", err)
	}
	return issues, nil
}

func newDashboardStyles() dashboardStyles {
	border := lipgloss.Border{
		Top:         "-",
		Bottom:      "-",
		Left:        "|",
		Right:       "|",
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
	}

	return dashboardStyles{
		header:         lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		welcome:        lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true),
		subheader:      lipgloss.NewStyle().Foreground(lipgloss.Color("222")),
		tag:            lipgloss.NewStyle().Foreground(lipgloss.Color("52")).Background(lipgloss.Color("220")).Padding(0, 1).Bold(true),
		columnTitle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		columnTitleDim: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("130")),
		panelTitle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		item:           lipgloss.NewStyle().Foreground(lipgloss.Color("254")),
		itemSelected:   lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("196")).Bold(true),
		itemSelectedIn: lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("130")),
		footer:         lipgloss.NewStyle().Foreground(lipgloss.Color("179")),
		helpBox:        lipgloss.NewStyle().Border(border).BorderForeground(lipgloss.Color("214")).Padding(1, 2).Foreground(lipgloss.Color("254")).Background(lipgloss.Color("52")),
		helpTitle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")),
		helpText:       lipgloss.NewStyle().Foreground(lipgloss.Color("254")),
		dimText:        lipgloss.NewStyle().Foreground(lipgloss.Color("178")),
		errorText:      lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		navbarBar:      lipgloss.NewStyle().Background(lipgloss.Color("124")).Foreground(lipgloss.Color("230")).Bold(true),
		navbarTitle:    lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Bold(true),
		navbarMeta:     lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true),
		navbarSub:      lipgloss.NewStyle().Foreground(lipgloss.Color("229")),
		tentTop:        lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Background(lipgloss.Color("124")).Bold(true),
		tentPost:       lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Background(lipgloss.Color("52")).Bold(true),
		tentBase:       lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Background(lipgloss.Color("124")).Bold(true),
		tentStripeA:    lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Background(lipgloss.Color("124")).Bold(true),
		tentStripeB:    lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("130")).Bold(true),
	}
}

func availableListHeight(height int) int {
	listHeight := height - 10
	if listHeight < 3 {
		return 3
	}
	return listHeight
}

func truncateASCII(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(value) <= width {
		return value
	}
	if width <= 3 {
		return value[:width]
	}
	return value[:width-3] + "..."
}

func joinWithGap(gap int, columns []string) []string {
	if gap <= 0 {
		return columns
	}
	output := make([]string, 0, len(columns)*2-1)
	spacer := strings.Repeat(" ", gap)
	for idx, col := range columns {
		if idx > 0 {
			output = append(output, spacer)
		}
		output = append(output, col)
	}
	return output
}

func formatTimestamp(value string) string {
	if value == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, value)
	}
	if err != nil {
		return value
	}
	return parsed.Format("Jan 02 15:04")
}

func wrapLines(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var current string
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if len(current)+1+len(word) > width {
			lines = append(lines, current)
			current = word
			continue
		}
		current = current + " " + word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
