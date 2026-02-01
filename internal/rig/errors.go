package rig

import (
	"errors"
	"fmt"
	"os"
)

type RigAlreadyExistsError struct {
	Name string
}

func (e RigAlreadyExistsError) Error() string {
	return fmt.Sprintf("rig %q already exists", e.Name)
}

type RigDirExistsError struct {
	Path string
}

func (e RigDirExistsError) Error() string {
	return fmt.Sprintf("rig directory already exists at %s", e.Path)
}

type RigCloneError struct {
	Remote string
	Path   string
	Err    error
}

func (e RigCloneError) Error() string {
	return fmt.Sprintf("clone rig %q to %s: %v", e.Remote, e.Path, e.Err)
}

func (e RigCloneError) Unwrap() error {
	return e.Err
}

func EnsureRigNameAvailable(name string) error {
	if name == "" {
		return fmt.Errorf("rig name is required")
	}

	rigs, err := ListRigs()
	if err != nil {
		return err
	}

	for _, rig := range rigs {
		if rig.Name == name {
			return RigAlreadyExistsError{Name: name}
		}
	}

	return nil
}

func EnsureRigPathAvailable(path string) error {
	if path == "" {
		return fmt.Errorf("rig path is required")
	}

	if _, err := os.Stat(path); err == nil {
		return RigDirExistsError{Path: path}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check rig path: %w", err)
	}

	return nil
}

func UserMessage(err error) string {
	if errors.Is(err, ErrRigNotFound) {
		return "Rig not found in registry. Add it with `bordertown rig add`."
	}

	var exists RigAlreadyExistsError
	if errors.As(err, &exists) {
		return fmt.Sprintf("Rig %q already exists. Choose a different name or remove it from the registry.", exists.Name)
	}

	var dirExists RigDirExistsError
	if errors.As(err, &dirExists) {
		return fmt.Sprintf("Directory %q already exists. Remove it or pick a different name.", dirExists.Path)
	}

	var cloneErr RigCloneError
	if errors.As(err, &cloneErr) {
		return fmt.Sprintf("Could not clone %q into %q. Check the remote URL and your network access.", cloneErr.Remote, cloneErr.Path)
	}

	return err.Error()
}
