package beads

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadIssuesFromFile(t *testing.T) {
	// Create temp file with test JSONL
	dir := t.TempDir()
	path := filepath.Join(dir, "issues.jsonl")

	content := `{"id":"cn-001","title":"Test issue","status":"open","priority":2,"issue_type":"task"}
{"id":"cn-002","title":"Test epic","status":"open","priority":1,"issue_type":"epic"}
{"id":"cn-003","title":"Closed task","status":"closed","priority":2,"issue_type":"task"}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	issues, err := LoadIssuesFromFile(path)
	if err != nil {
		t.Fatalf("LoadIssuesFromFile: %v", err)
	}

	if len(issues) != 3 {
		t.Errorf("expected 3 issues, got %d", len(issues))
	}

	// Check first issue
	if issues[0].ID != "cn-001" {
		t.Errorf("expected id cn-001, got %s", issues[0].ID)
	}
	if !issues[0].IsOpen() {
		t.Error("expected issue to be open")
	}
	if issues[0].IsEpic() {
		t.Error("expected issue to not be epic")
	}

	// Check epic
	if !issues[1].IsEpic() {
		t.Error("expected issue to be epic")
	}

	// Check closed
	if issues[2].IsOpen() {
		t.Error("expected issue to be closed")
	}
}

func TestLoadIssues_NotFound(t *testing.T) {
	_, err := LoadIssuesFromFile("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestFindBeadsRoot(t *testing.T) {
	// Create temp directory structure
	root := t.TempDir()
	beadsDir := filepath.Join(root, ".beads")
	subDir := filepath.Join(root, "sub", "deep")

	if err := os.Mkdir(beadsDir, 0755); err != nil {
		t.Fatalf("mkdir .beads: %v", err)
	}
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("mkdir sub/deep: %v", err)
	}

	// Find from subdirectory
	found, err := FindBeadsRoot(subDir)
	if err != nil {
		t.Fatalf("FindBeadsRoot: %v", err)
	}
	if found != root {
		t.Errorf("expected %s, got %s", root, found)
	}
}
