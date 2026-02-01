package dashboard

import (
	"strings"
	"time"
)

func availableListHeight(height int) int {
	listHeight := height - 2
	if listHeight < 3 {
		return 3
	}
	return listHeight
}

func drawerLayout(totalWidth int, totalHeight int) (drawerWidth int, bodyHeight int, listHeight int, showSummary bool) {
	bodyHeight = availableListHeight(totalHeight)
	if bodyHeight < 3 {
		bodyHeight = 3
	}
	minDrawer := 32
	maxDrawer := 54
	drawerWidth = totalWidth * 9 / 20
	if drawerWidth < minDrawer {
		drawerWidth = minDrawer
	}
	if drawerWidth > maxDrawer {
		drawerWidth = maxDrawer
	}
	if drawerWidth > totalWidth-12 {
		drawerWidth = totalWidth - 12
	}
	if drawerWidth < 20 {
		drawerWidth = minInt(20, totalWidth)
	}
	showSummary = bodyHeight >= 6
	headerLines := 1
	if showSummary {
		headerLines = 2
	}
	listHeight = bodyHeight - headerLines - 2
	if listHeight < 1 {
		listHeight = 1
	}
	return drawerWidth, bodyHeight, listHeight, showSummary
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

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func wrapTitle(title string, firstWidth int, nextWidth int) []string {
	if firstWidth <= 0 {
		firstWidth = 1
	}
	if nextWidth <= 0 {
		nextWidth = 1
	}
	words := strings.Fields(title)
	if len(words) == 0 {
		return []string{""}
	}

	lines := []string{}
	current := ""
	currentWidth := firstWidth
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if len(current)+1+len(word) > currentWidth {
			lines = append(lines, current)
			current = word
			currentWidth = nextWidth
			continue
		}
		current = current + " " + word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
