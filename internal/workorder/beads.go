package workorder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type BeadInfo struct {
	ID          string
	Title       string
	Description string
}

type beadIssue struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func LoadBeadIndex(startDir string) (map[string]BeadInfo, error) {
	root, err := findBeadsRoot(startDir)
	if err != nil {
		return nil, err
	}
	issues, err := loadBeadIssues(root)
	if err != nil {
		return nil, err
	}
	index := make(map[string]BeadInfo, len(issues))
	for _, issue := range issues {
		index[issue.ID] = BeadInfo{ID: issue.ID, Title: issue.Title, Description: issue.Description}
	}
	return index, nil
}

func findBeadsRoot(startDir string) (string, error) {
	dir := startDir
	for {
		beadsPath := filepath.Join(dir, ".beads")
		if info, err := os.Stat(beadsPath); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .beads directory found")
		}
		dir = parent
	}
}

func loadBeadIssues(root string) ([]beadIssue, error) {
	path := filepath.Join(root, ".beads", "issues.jsonl")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open beads issues: %w", err)
	}
	defer file.Close()

	var issues []beadIssue
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}
		var issue beadIssue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			return nil, fmt.Errorf("parse bead issue at line %d: %w", lineNum, err)
		}
		issues = append(issues, issue)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read beads issues: %w", err)
	}
	return issues, nil
}
