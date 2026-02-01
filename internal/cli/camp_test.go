package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rikurb8/carnie/internal/config"
)

func TestCampInitCreatesConfig(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"camp", "init", "--name", "test-camp"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	configPath := filepath.Join(dir, config.CampConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("expected %s to be created", config.CampConfigFile)
	}

	cfg, err := config.LoadCampConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Name != "test-camp" {
		t.Errorf("expected name 'test-camp', got %q", cfg.Name)
	}

	if cfg.Version != config.CurrentVersion {
		t.Errorf("expected version %d, got %d", config.CurrentVersion, cfg.Version)
	}
}

func TestCampInitDefaultsToDirectoryName(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"camp", "init"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cfg, err := config.LoadCampConfig(filepath.Join(dir, config.CampConfigFile))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	expectedName := filepath.Base(dir)
	if cfg.Name != expectedName {
		t.Errorf("expected name %q (dir name), got %q", expectedName, cfg.Name)
	}
}

func TestCampInitErrorsIfExists(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	configPath := filepath.Join(dir, config.CampConfigFile)
	os.WriteFile(configPath, []byte("existing"), 0644)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"camp", "init"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when camp.yml exists")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestCampInitForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	configPath := filepath.Join(dir, config.CampConfigFile)
	os.WriteFile(configPath, []byte("old content"), 0644)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"camp", "init", "--name", "new-camp", "--force"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error with --force, got %v", err)
	}

	cfg, err := config.LoadCampConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Name != "new-camp" {
		t.Errorf("expected name 'new-camp', got %q", cfg.Name)
	}
}

func TestCampInitSetsDescription(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"camp", "init", "--name", "my-camp", "--description", "A test workspace"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cfg, err := config.LoadCampConfig(filepath.Join(dir, config.CampConfigFile))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Description != "A test workspace" {
		t.Errorf("expected description 'A test workspace', got %q", cfg.Description)
	}
}

func TestCampInitConfigHasDefaults(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"camp", "init", "--name", "test"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cfg, err := config.LoadCampConfig(filepath.Join(dir, config.CampConfigFile))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Operator.Model != config.DefaultOperatorModel {
		t.Errorf("expected operator.model %q, got %q", config.DefaultOperatorModel, cfg.Operator.Model)
	}

	if cfg.Defaults.AgentModel != config.DefaultAgentModel {
		t.Errorf("expected defaults.agent_model %q, got %q", config.DefaultAgentModel, cfg.Defaults.AgentModel)
	}
}
