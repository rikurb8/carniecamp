package dashboard

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/rikurb8/carnie/internal/bd"
)

type Issue struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Status       string       `json:"status"`
	Priority     int          `json:"priority"`
	IssueType    string       `json:"issue_type"`
	Owner        string       `json:"owner"`
	UpdatedAt    string       `json:"updated_at"`
	CreatedAt    string       `json:"created_at"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
}

type Dependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
}

type dataState struct {
	Status     bd.Status
	Ready      []Issue
	InProgress []Issue
	Blocked    []Issue
	Closed     []Issue
}

type dataMsg struct {
	Data dataState
	Err  error
}

type tickMsg struct{}

type issueColumn struct {
	Title      string
	Issues     []Issue
	ParentByID map[string]string
	LevelByID  map[string]int
}

type ViewType int

const (
	ViewHome ViewType = iota
	ViewTasks
)

type Model struct {
	width           int
	height          int
	columns         []issueColumn
	activeColumn    int
	activeView      ViewType
	refresh         time.Duration
	limit           int
	lastUpdated     time.Time
	summary         bd.StatusSummary
	errMessage      string
	showHelp        bool
	collapsed       map[string]bool
	featureChildren map[string][]Issue
	lists           []list.Model
	homeSpinners    []spinner.Model
}

type drawerEntry struct {
	Issue Issue
	Level int
}

func NewModel(refresh time.Duration, limit int) Model {
	styles := newDashboardStyles()
	delegate := newDrawerDelegate(styles, 1)
	futureList := newDrawerList(delegate)

	// Multiple spinner styles for the border
	spinnerTypes := []spinner.Spinner{
		spinner.Dot,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
		spinner.Line,
		spinner.Jump,
	}
	colors := []string{"196", "220", "214", "226", "202", "198"}
	spinners := make([]spinner.Model, len(spinnerTypes))
	for i, st := range spinnerTypes {
		s := spinner.New()
		s.Spinner = st
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(colors[i]))
		spinners[i] = s
	}

	return Model{
		refresh:         refresh,
		limit:           limit,
		collapsed:       map[string]bool{},
		featureChildren: map[string][]Issue{},
		columns: []issueColumn{
			{Title: "Open Features"},
		},
		lists:        []list.Model{futureList},
		homeSpinners: spinners,
	}
}
