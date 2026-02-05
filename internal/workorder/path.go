package workorder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rikurb8/carnie/internal/config"
)

const (
	workOrderDir    = ".carnie"
	workOrderDBFile = "carniecamp.db"
)

func FindCampRoot(startDir string) (string, error) {
	dir := startDir
	for {
		configPath := filepath.Join(dir, config.CampConfigFile)
		if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no %s found", config.CampConfigFile)
		}
		dir = parent
	}
}

func DefaultDBPath(startDir string) (string, error) {
	root, err := FindCampRoot(startDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, workOrderDir, workOrderDBFile), nil
}
