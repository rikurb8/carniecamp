package dashboard

import "github.com/charmbracelet/lipgloss"

type dashboardStyles struct {
	header             lipgloss.Style
	welcome            lipgloss.Style
	subheader          lipgloss.Style
	tag                lipgloss.Style
	columnTitle        lipgloss.Style
	columnTitleDim     lipgloss.Style
	panelTitle         lipgloss.Style
	item               lipgloss.Style
	itemSelected       lipgloss.Style
	itemSelectedIn     lipgloss.Style
	paneBorder         lipgloss.Style
	paneBorderActive   lipgloss.Style
	columnHeader       lipgloss.Style
	footer             lipgloss.Style
	helpBox            lipgloss.Style
	helpTitle          lipgloss.Style
	helpText           lipgloss.Style
	dimText            lipgloss.Style
	errorText          lipgloss.Style
	drawerItem         lipgloss.Style
	drawerItemSelected lipgloss.Style
	drawerEpic         lipgloss.Style
	badgePriority      lipgloss.Style
	badgeDefault       lipgloss.Style
	badgeReady         lipgloss.Style
	badgeProgress      lipgloss.Style
	badgeBlocked       lipgloss.Style
	badgeClosed        lipgloss.Style
	navbarBar          lipgloss.Style
	navbarTitle        lipgloss.Style
	navbarMeta         lipgloss.Style
	navbarSub          lipgloss.Style
	viewTabActive      lipgloss.Style
	viewTabInactive    lipgloss.Style
	tentTop            lipgloss.Style
	tentPost           lipgloss.Style
	tentBase           lipgloss.Style
	tentStripeA        lipgloss.Style
	tentStripeB        lipgloss.Style
}

func newDashboardStyles() dashboardStyles {
	border := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}

	return dashboardStyles{
		header:             lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		welcome:            lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true),
		subheader:          lipgloss.NewStyle().Foreground(lipgloss.Color("222")),
		tag:                lipgloss.NewStyle().Foreground(lipgloss.Color("52")).Background(lipgloss.Color("220")).Padding(0, 1).Bold(true),
		columnTitle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		columnTitleDim:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("130")),
		panelTitle:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		item:               lipgloss.NewStyle().Foreground(lipgloss.Color("254")),
		itemSelected:       lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("196")).Bold(true),
		itemSelectedIn:     lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("130")),
		paneBorder:         lipgloss.NewStyle().Foreground(lipgloss.Color("130")),
		paneBorderActive:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		columnHeader:       lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true).Underline(true),
		footer:             lipgloss.NewStyle().Foreground(lipgloss.Color("179")),
		helpBox:            lipgloss.NewStyle().Border(border).BorderForeground(lipgloss.Color("214")).Padding(1, 2).Foreground(lipgloss.Color("254")).Background(lipgloss.Color("52")),
		helpTitle:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")),
		helpText:           lipgloss.NewStyle().Foreground(lipgloss.Color("254")),
		dimText:            lipgloss.NewStyle().Foreground(lipgloss.Color("178")),
		errorText:          lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		drawerItem:         lipgloss.NewStyle().Foreground(lipgloss.Color("254")),
		drawerItemSelected: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("196")).Bold(true),
		drawerEpic:         lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		badgePriority:      lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("130")).Padding(0, 1).Bold(true),
		badgeDefault:       lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("238")).Padding(0, 1).Bold(true),
		badgeReady:         lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("70")).Padding(0, 1).Bold(true),
		badgeProgress:      lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("33")).Padding(0, 1).Bold(true),
		badgeBlocked:       lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("160")).Padding(0, 1).Bold(true),
		badgeClosed:        lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("28")).Padding(0, 1).Bold(true),
		navbarBar:          lipgloss.NewStyle().Background(lipgloss.Color("124")).Foreground(lipgloss.Color("230")).Bold(true),
		navbarTitle:        lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Bold(true),
		navbarMeta:         lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true),
		navbarSub:          lipgloss.NewStyle().Foreground(lipgloss.Color("229")),
		viewTabActive:      lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("196")).Bold(true).Padding(0, 1),
		viewTabInactive:    lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Padding(0, 1),
		tentTop:            lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Background(lipgloss.Color("124")).Bold(true),
		tentPost:           lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Background(lipgloss.Color("52")).Bold(true),
		tentBase:           lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Background(lipgloss.Color("124")).Bold(true),
		tentStripeA:        lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Background(lipgloss.Color("124")).Bold(true),
		tentStripeB:        lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("130")).Bold(true),
	}
}
