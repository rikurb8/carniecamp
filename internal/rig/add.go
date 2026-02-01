package rig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func AddRig(name string, remote string) (*Rig, error) {
	if err := ValidateRigName(name); err != nil {
		return nil, err
	}
	if err := ValidateRemote(remote); err != nil {
		return nil, err
	}
	if err := EnsureRigNameAvailable(name); err != nil {
		return nil, err
	}

	rigsDir, err := EnsureRigsDir()
	if err != nil {
		return nil, err
	}

	localPath := filepath.Join(rigsDir, name)
	if err := EnsureRigPathAvailable(localPath); err != nil {
		return nil, err
	}

	if err := cloneRig(remote, localPath); err != nil {
		return nil, RigCloneError{Remote: remote, Path: localPath, Err: err}
	}

	registry, err := LoadRegistry()
	if err != nil {
		return nil, err
	}

	newRig := Rig{
		Name:      name,
		Remote:    remote,
		LocalPath: localPath,
	}
	registry.Rigs = append(registry.Rigs, newRig)

	if err := SaveRegistry(registry); err != nil {
		return nil, err
	}

	return &newRig, nil
}

func ValidateRigName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("rig name is required")
	}
	if name != filepath.Base(name) || name == "." || name == ".." {
		return fmt.Errorf("rig name must be a directory name, not a path")
	}
	return nil
}

func ValidateRemote(remote string) error {
	if strings.TrimSpace(remote) == "" {
		return fmt.Errorf("rig remote is required")
	}
	if strings.Contains(remote, " ") {
		return fmt.Errorf("rig remote must not contain spaces")
	}
	return nil
}

func cloneRig(remote string, path string) error {
	cmd := exec.Command("git", "clone", remote, path)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return err
	}
	return nil
}
