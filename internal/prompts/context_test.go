package prompts

import (
	"strings"
	"testing"

	"github.com/rikurb8/carnie/internal/config"
)

func TestFormatContextSection_Empty(t *testing.T) {
	ctx := ProjectContext{}
	result := FormatContextSection(ctx)
	if result != "" {
		t.Errorf("expected empty string for empty context, got %q", result)
	}
}

func TestFormatContextSection_WithProject(t *testing.T) {
	ctx := ProjectContext{
		ProjectName:        "TestProject",
		ProjectDescription: "A test project",
	}
	result := FormatContextSection(ctx)

	if !strings.Contains(result, "TestProject") {
		t.Error("expected result to contain project name")
	}
	if !strings.Contains(result, "A test project") {
		t.Error("expected result to contain project description")
	}
}

func TestFormatContextSection_WithEpics(t *testing.T) {
	ctx := ProjectContext{
		ProjectName: "TestProject",
		Epics: []EpicSummary{
			{ID: "epic-1", Title: "First Epic", Status: "open", TaskCount: 3},
			{ID: "epic-2", Title: "Second Epic", Status: "closed", TaskCount: 5},
		},
	}
	result := FormatContextSection(ctx)

	if !strings.Contains(result, "epic-1") {
		t.Error("expected result to contain first epic ID")
	}
	if !strings.Contains(result, "First Epic") {
		t.Error("expected result to contain first epic title")
	}
	if !strings.Contains(result, "3 tasks") {
		t.Error("expected result to contain task count")
	}
	if !strings.Contains(result, "Existing Epics") {
		t.Error("expected result to contain Existing Epics header")
	}
}

func TestBuildSystemPrompt_IncludesBase(t *testing.T) {
	ctx := ProjectContext{
		ProjectName: "Test",
	}
	result := BuildSystemPrompt(ctx)

	// Should include the base prompt
	if !strings.Contains(result, "expert software project planner") {
		t.Error("expected result to include base planning prompt")
	}
	// Should include context
	if !strings.Contains(result, "Test") {
		t.Error("expected result to include project context")
	}
}

func TestBuildSystemPrompt_EmptyContext(t *testing.T) {
	ctx := ProjectContext{}
	result := BuildSystemPrompt(ctx)

	// Should just be the base prompt without context section
	if !strings.HasPrefix(result, "You are an expert") {
		t.Error("expected result to start with base prompt when no context")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestGatherContext_WithConfig(t *testing.T) {
	cfg := &config.CampConfig{
		Name:        "MyProject",
		Description: "Project description",
	}

	// Use a non-existent directory so beads won't be found
	ctx := GatherContext("/nonexistent", cfg)

	if ctx.ProjectName != "MyProject" {
		t.Errorf("expected ProjectName to be 'MyProject', got %q", ctx.ProjectName)
	}
	if ctx.ProjectDescription != "Project description" {
		t.Errorf("expected ProjectDescription to be set")
	}
}
