package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupBeadsDir(t *testing.T, content string) string {
	dir := t.TempDir()
	beadsDir := filepath.Join(dir, ".beads")
	if err := os.Mkdir(beadsDir, 0755); err != nil {
		t.Fatalf("create .beads dir: %v", err)
	}

	issuesPath := filepath.Join(beadsDir, "issues.jsonl")
	if err := os.WriteFile(issuesPath, []byte(content), 0644); err != nil {
		t.Fatalf("write issues.jsonl: %v", err)
	}

	return dir
}

func TestMayorReviewEmptyBeads(t *testing.T) {
	dir := setupBeadsDir(t, "")
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"mayor", "review"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Mayor Review") {
		t.Error("expected output to contain 'Mayor Review'")
	}
	if !strings.Contains(result, "No epics found") {
		t.Error("expected output to contain 'No epics found'")
	}
}

func TestMayorReviewSingleEpic(t *testing.T) {
	content := `{"id":"epic-1","title":"Test Epic","status":"open","issue_type":"epic","description":"A test epic with enough description to pass the length check for planning analysis"}
{"id":"task-1","title":"Task One","status":"open","issue_type":"task"}`

	dir := setupBeadsDir(t, content)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"mayor", "review"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "epic-1") {
		t.Error("expected output to contain epic-1")
	}
	if !strings.Contains(result, "Test Epic") {
		t.Error("expected output to contain epic title")
	}
}

func TestMayorReviewEpicWithTasks(t *testing.T) {
	// Epic depends on task-1 (task-1 blocks the epic)
	content := `{"id":"epic-1","title":"Test Epic","status":"open","issue_type":"epic","description":"A test epic","dependencies":[{"issue_id":"epic-1","depends_on_id":"task-1","type":"blocks"}]}
{"id":"task-1","title":"Child Task","status":"open","issue_type":"task"}`

	dir := setupBeadsDir(t, content)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"mayor", "review"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "epic-1") {
		t.Error("expected output to contain epic-1")
	}
	if !strings.Contains(result, "task-1") {
		t.Error("expected output to contain task-1 as child")
	}
	if !strings.Contains(result, "1 open") {
		t.Error("expected output to show task count")
	}
}

func TestMayorReviewEpicNeedsPlanning(t *testing.T) {
	// Epic with no tasks should show needs planning
	content := `{"id":"epic-1","title":"Empty Epic","status":"open","issue_type":"epic","description":"Short"}`

	dir := setupBeadsDir(t, content)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"mayor", "review"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Needs planning") {
		t.Error("expected output to show 'Needs planning' warning")
	}
}

func TestMayorReviewOrphanIssues(t *testing.T) {
	// Task with no epic parent
	content := `{"id":"task-1","title":"Orphan Task","status":"open","issue_type":"task"}`

	dir := setupBeadsDir(t, content)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"mayor", "review"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Orphan Issues") {
		t.Error("expected output to show 'Orphan Issues' section")
	}
	if !strings.Contains(result, "task-1") {
		t.Error("expected output to contain orphan task-1")
	}
}

func TestMayorReviewHidesClosedByDefault(t *testing.T) {
	content := `{"id":"epic-1","title":"Closed Epic","status":"closed","issue_type":"epic"}
{"id":"epic-2","title":"Open Epic","status":"open","issue_type":"epic","description":"A long enough description for the planning check"}`

	dir := setupBeadsDir(t, content)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"mayor", "review"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result := output.String()
	if strings.Contains(result, "epic-1") {
		t.Error("expected closed epic-1 to be hidden by default")
	}
	if !strings.Contains(result, "epic-2") {
		t.Error("expected open epic-2 to be shown")
	}
}

func TestMayorReviewShowAllFlag(t *testing.T) {
	content := `{"id":"epic-1","title":"Closed Epic","status":"closed","issue_type":"epic"}
{"id":"epic-2","title":"Open Epic","status":"open","issue_type":"epic"}`

	dir := setupBeadsDir(t, content)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"mayor", "review", "--all"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "epic-1") {
		t.Error("expected closed epic-1 to be shown with --all")
	}
	if !strings.Contains(result, "epic-2") {
		t.Error("expected open epic-2 to be shown with --all")
	}
}

func TestMayorReviewNoBeadsDirectory(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"mayor", "review"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no .beads directory")
	}

	if !strings.Contains(err.Error(), "find beads") {
		t.Errorf("expected 'find beads' error, got: %v", err)
	}
}
