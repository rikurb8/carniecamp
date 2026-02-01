package rig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	rigsDirName       = "rigs"
	registryFileName  = "registry.yml"
	bordertownDirName = ".bordertown"
)

var ErrRigNotFound = errors.New("rig not found")

type Registry struct {
	Rigs []Rig `yaml:"rigs"`
}

type Rig struct {
	Name      string `yaml:"name"`
	Remote    string `yaml:"remote"`
	LocalPath string `yaml:"local_path"`
}

func RigsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	return filepath.Join(home, bordertownDirName, rigsDirName), nil
}

func EnsureRigsDir() (string, error) {
	dir, err := RigsDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create rigs dir: %w", err)
	}

	return dir, nil
}

func RegistryPath() (string, error) {
	dir, err := EnsureRigsDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, registryFileName), nil
}

func LoadRegistry() (*Registry, error) {
	registryPath, err := RegistryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(registryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Registry{}, nil
		}
		return nil, fmt.Errorf("read registry: %w", err)
	}

	var registry Registry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}

	return &registry, nil
}

func SaveRegistry(registry *Registry) error {
	if registry == nil {
		return fmt.Errorf("registry is required")
	}

	registryPath, err := RegistryPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(registry)
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(registryPath, data, 0644); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}

	return nil
}

func ListRigs() ([]Rig, error) {
	registry, err := LoadRegistry()
	if err != nil {
		return nil, err
	}

	if registry.Rigs == nil {
		return []Rig{}, nil
	}

	return registry.Rigs, nil
}

func LoadRig(name string) (*Rig, error) {
	if name == "" {
		return nil, fmt.Errorf("rig name is required")
	}

	rigs, err := ListRigs()
	if err != nil {
		return nil, err
	}

	for _, rig := range rigs {
		if rig.Name == name {
			match := rig
			return &match, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrRigNotFound, name)
}
