package dashboard

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/rikurb8/carnie/internal/cli/bd"
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
	Selected   int
	Offset     int
	ParentByID map[string]string
	LevelByID  map[string]int
}

type Model struct {
	width        int
	height       int
	columns      []issueColumn
	activeColumn int
	refresh      time.Duration
	limit        int
	lastUpdated  time.Time
	summary      bd.StatusSummary
	errMessage   string
	showHelp     bool
	showImprover bool
	improveInput textinput.Model
	collapsed    map[string]bool
}

type drawerEntry struct {
	Issue Issue
	Level int
}

func NewModel(refresh time.Duration, limit int) Model {
	improve := textinput.New()
	improve.Placeholder = "Optional instructions"
	improve.Prompt = ""
	improve.CharLimit = 500

	return Model{
		refresh:      refresh,
		limit:        limit,
		collapsed:    map[string]bool{},
		improveInput: improve,
		columns: []issueColumn{
			{Title: "Future Work"},
			{Title: "Completed Work"},
		},
	}
}
