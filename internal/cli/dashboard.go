package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type bdIssue struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	IssueType   string `json:"issue_type"`
	Owner       string `json:"owner"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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
	Title    string
	Issues   []bdIssue
	Selected int
	Offset   int
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
}

type dashboardStyles struct {
	header         lipgloss.Style
	subheader      lipgloss.Style
	tag            lipgloss.Style
	columnTitle    lipgloss.Style
	columnTitleDim lipgloss.Style
	item           lipgloss.Style
	itemSelected   lipgloss.Style
	itemSelectedIn lipgloss.Style
	footer         lipgloss.Style
	helpBox        lipgloss.Style
	helpTitle      lipgloss.Style
	helpText       lipgloss.Style
	dimText        lipgloss.Style
	errorText      lipgloss.Style
}

func newDashboardCommand() *cobra.Command {
	var refresh time.Duration
	var limit int

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Launch the Bordertown dashboard",
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
		refresh: refresh,
		limit:   limit,
		columns: []issueColumn{
			{Title: "Ready"},
			{Title: "In Progress"},
			{Title: "Blocked"},
			{Title: "Closed"},
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
		case "tab", "right", "l":
			m.activeColumn = (m.activeColumn + 1) % len(m.columns)
			m.ensureVisible()
			return m, nil
		case "shift+tab", "left":
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
	header := renderDashboardHeader(styles)
	stats := renderDashboardStats(m, styles)
	body := renderDashboardColumns(m, styles)
	footer := renderDashboardFooter(m, styles)

	output := lipgloss.JoinVertical(lipgloss.Left, header, stats, body, footer)
	if m.showHelp {
		output = renderHelpOverlay(output, m, styles)
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
	if len(column.Issues) == 0 {
		column.Selected = 0
		column.Offset = 0
		return
	}

	next := column.Selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(column.Issues) {
		next = len(column.Issues) - 1
	}
	column.Selected = next
	m.ensureVisible()
}

func (m *dashboardModel) ensureVisible() {
	height := availableListHeight(m.height)
	for idx := range m.columns {
		column := &m.columns[idx]
		if len(column.Issues) == 0 {
			column.Selected = 0
			column.Offset = 0
			continue
		}

		if column.Selected < 0 {
			column.Selected = 0
		}
		if column.Selected >= len(column.Issues) {
			column.Selected = len(column.Issues) - 1
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
		if column.Offset > len(column.Issues)-1 {
			column.Offset = len(column.Issues) - 1
		}
	}
}

func (m *dashboardModel) updateColumns(data dashboardData) {
	previous := make(map[string]string, len(m.columns))
	for _, column := range m.columns {
		if len(column.Issues) == 0 || column.Selected >= len(column.Issues) {
			continue
		}
		previous[column.Title] = column.Issues[column.Selected].ID
	}

	m.columns[0].Issues = data.Ready
	m.columns[1].Issues = data.InProgress
	m.columns[2].Issues = data.Blocked
	m.columns[3].Issues = data.Closed

	for idx := range m.columns {
		selectedID := previous[m.columns[idx].Title]
		if selectedID == "" || len(m.columns[idx].Issues) == 0 {
			m.columns[idx].Selected = 0
			m.columns[idx].Offset = 0
			continue
		}
		for issueIndex, issue := range m.columns[idx].Issues {
			if issue.ID == selectedID {
				m.columns[idx].Selected = issueIndex
				break
			}
		}
	}
}

func renderDashboardHeader(styles dashboardStyles) string {
	art := []string{
		" ____                  _               _                      ",
		"|  _ \\ ___  __ _ _   _| | ___  _ __ __| | ___  _ __ ___  _ __ ",
		"| |_) / _ \\ / _` | | | | |/ _ \\| '__/ _` |/ _ \\| '_ ` _ \\| '_ \\",
		"|  _ <  __/ (_| | |_| | | (_) | | | (_| | (_) | | | | | | |_) |",
		"|_| \\_\\___|\\__, |\\__,_|_|\\___/|_|  \\__,_|\\___/|_| |_| |_| .__/ ",
		"            |___/                                        |_|    ",
	}

	return styles.header.Render(strings.Join(art, "\n"))
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
	return styles.subheader.Render(lipgloss.JoinHorizontal(lipgloss.Left, tags...))
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
	hint := "h help  tab switch  j/k move  r refresh  q quit"
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
		styles.helpText.Render("Keys: q quit  tab switch columns  j/k move  r refresh  h close"),
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

func (m dashboardModel) selectedIssue() *bdIssue {
	if len(m.columns) == 0 {
		return nil
	}
	column := m.columns[m.activeColumn]
	if len(column.Issues) == 0 || column.Selected >= len(column.Issues) {
		return nil
	}
	return &column.Issues[column.Selected]
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

	ready, err := fetchIssues([]string{"list", "--ready", "--json", "--sort", "priority"}, limit)
	if err != nil {
		return dashboardData{}, err
	}
	inProgress, err := fetchIssues([]string{"list", "--status", "in_progress", "--json", "--sort", "priority"}, limit)
	if err != nil {
		return dashboardData{}, err
	}
	blocked, err := fetchIssues([]string{"list", "--status", "blocked", "--json", "--sort", "priority"}, limit)
	if err != nil {
		return dashboardData{}, err
	}
	closed, err := fetchIssues([]string{"list", "--status", "closed", "--json", "--sort", "updated", "--reverse"}, limit)
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
		header:         lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Bold(true),
		subheader:      lipgloss.NewStyle().Foreground(lipgloss.Color("223")),
		tag:            lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Background(lipgloss.Color("215")).Padding(0, 1),
		columnTitle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("221")),
		columnTitleDim: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("241")),
		item:           lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		itemSelected:   lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("60")).Bold(true),
		itemSelectedIn: lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("236")),
		footer:         lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		helpBox:        lipgloss.NewStyle().Border(border).Padding(1, 2).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("235")),
		helpTitle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("222")),
		helpText:       lipgloss.NewStyle().Foreground(lipgloss.Color("230")),
		dimText:        lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
		errorText:      lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
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
