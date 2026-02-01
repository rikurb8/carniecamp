package dashboard

import "github.com/charmbracelet/lipgloss"

type dashboardStyles struct {
	header           lipgloss.Style
	welcome          lipgloss.Style
	subheader        lipgloss.Style
	tag              lipgloss.Style
	columnTitle      lipgloss.Style
	columnTitleDim   lipgloss.Style
	panelTitle       lipgloss.Style
	item             lipgloss.Style
	itemSelected     lipgloss.Style
	itemSelectedIn   lipgloss.Style
	paneBorder       lipgloss.Style
	paneBorderActive lipgloss.Style
	columnHeader     lipgloss.Style
	footer           lipgloss.Style
	helpBox          lipgloss.Style
	helpTitle        lipgloss.Style
	helpText         lipgloss.Style
	dimText          lipgloss.Style
	errorText        lipgloss.Style
	navbarBar        lipgloss.Style
	navbarTitle      lipgloss.Style
	navbarMeta       lipgloss.Style
	navbarSub        lipgloss.Style
	tentTop          lipgloss.Style
	tentPost         lipgloss.Style
	tentBase         lipgloss.Style
	tentStripeA      lipgloss.Style
	tentStripeB      lipgloss.Style
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
		header:           lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		welcome:          lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true),
		subheader:        lipgloss.NewStyle().Foreground(lipgloss.Color("222")),
		tag:              lipgloss.NewStyle().Foreground(lipgloss.Color("52")).Background(lipgloss.Color("220")).Padding(0, 1).Bold(true),
		columnTitle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		columnTitleDim:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("130")),
		panelTitle:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		item:             lipgloss.NewStyle().Foreground(lipgloss.Color("254")),
		itemSelected:     lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("196")).Bold(true),
		itemSelectedIn:   lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("130")),
		paneBorder:       lipgloss.NewStyle().Foreground(lipgloss.Color("130")),
		paneBorderActive: lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		columnHeader:     lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true).Underline(true),
		footer:           lipgloss.NewStyle().Foreground(lipgloss.Color("179")),
		helpBox:          lipgloss.NewStyle().Border(border).BorderForeground(lipgloss.Color("214")).Padding(1, 2).Foreground(lipgloss.Color("254")).Background(lipgloss.Color("52")),
		helpTitle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")),
		helpText:         lipgloss.NewStyle().Foreground(lipgloss.Color("254")),
		dimText:          lipgloss.NewStyle().Foreground(lipgloss.Color("178")),
		errorText:        lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		navbarBar:        lipgloss.NewStyle().Background(lipgloss.Color("124")).Foreground(lipgloss.Color("230")).Bold(true),
		navbarTitle:      lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Bold(true),
		navbarMeta:       lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true),
		navbarSub:        lipgloss.NewStyle().Foreground(lipgloss.Color("229")),
		tentTop:          lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Background(lipgloss.Color("124")).Bold(true),
		tentPost:         lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Background(lipgloss.Color("52")).Bold(true),
		tentBase:         lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Background(lipgloss.Color("124")).Bold(true),
		tentStripeA:      lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Background(lipgloss.Color("124")).Bold(true),
		tentStripeB:      lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("130")).Bold(true),
	}
}
