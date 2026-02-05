package workorder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBeadIndex(t *testing.T) {
	dir := t.TempDir()
	beadsDir := filepath.Join(dir, ".beads")
	if err := os.MkdirAll(beadsDir, 0755); err != nil {
		t.Fatalf("create beads dir: %v", err)
	}

	data := []byte("{\"id\":\"cn-ta1.2\",\"title\":\"Add WorkOrder domain\",\"description\":\"Define status transitions\"}\n")
	if err := os.WriteFile(filepath.Join(beadsDir, "issues.jsonl"), data, 0644); err != nil {
		t.Fatalf("write issues file: %v", err)
	}

	index, err := LoadBeadIndex(dir)
	if err != nil {
		t.Fatalf("load bead index: %v", err)
	}

	info, ok := index["cn-ta1.2"]
	if !ok {
		t.Fatal("expected bead info for cn-ta1.2")
	}
	if info.Description == "" {
		t.Fatal("expected bead description to be populated")
	}
}
