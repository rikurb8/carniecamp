package prompts

import (
	"os"
	"path/filepath"
)

const (
	// DefaultPromptDir is the default directory for custom prompts.
	DefaultPromptDir = ".carnie/prompts"
	// DefaultEpicPlanningFile is the default filename for epic planning prompts.
	DefaultEpicPlanningFile = "epic-planning.md"
)

// LoadEpicPlanningPrompt loads the epic planning prompt from a custom file or returns the built-in default.
// It checks in order:
// 1. The explicitly configured path (if provided)
// 2. The default location (.carnie/prompts/epic-planning.md)
// 3. Falls back to the built-in EpicPlanningPrompt
func LoadEpicPlanningPrompt(workDir string, configuredPath string) string {
	// Try explicitly configured path first
	if configuredPath != "" {
		path := configuredPath
		if !filepath.IsAbs(path) {
			path = filepath.Join(workDir, path)
		}
		if content, err := os.ReadFile(path); err == nil {
			return string(content)
		}
	}

	// Try default location
	defaultPath := filepath.Join(workDir, DefaultPromptDir, DefaultEpicPlanningFile)
	if content, err := os.ReadFile(defaultPath); err == nil {
		return string(content)
	}

	// Fall back to built-in
	return EpicPlanningPrompt
}

// PromptFileExists checks if a custom prompt file exists at the given path or default location.
func PromptFileExists(workDir string, configuredPath string) bool {
	if configuredPath != "" {
		path := configuredPath
		if !filepath.IsAbs(path) {
			path = filepath.Join(workDir, path)
		}
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	defaultPath := filepath.Join(workDir, DefaultPromptDir, DefaultEpicPlanningFile)
	_, err := os.Stat(defaultPath)
	return err == nil
}
