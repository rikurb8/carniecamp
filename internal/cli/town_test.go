package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rikurb8/bordertown/internal/config"
)

func TestTownInitCreatesConfig(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"town", "init", "--name", "test-town"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	configPath := filepath.Join(dir, config.TownConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("expected %s to be created", config.TownConfigFile)
	}

	cfg, err := config.LoadTownConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Name != "test-town" {
		t.Errorf("expected name 'test-town', got %q", cfg.Name)
	}

	if cfg.Version != config.CurrentVersion {
		t.Errorf("expected version %d, got %d", config.CurrentVersion, cfg.Version)
	}
}

func TestTownInitDefaultsToDirectoryName(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"town", "init"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cfg, err := config.LoadTownConfig(filepath.Join(dir, config.TownConfigFile))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	expectedName := filepath.Base(dir)
	if cfg.Name != expectedName {
		t.Errorf("expected name %q (dir name), got %q", expectedName, cfg.Name)
	}
}

func TestTownInitErrorsIfExists(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	configPath := filepath.Join(dir, config.TownConfigFile)
	os.WriteFile(configPath, []byte("existing"), 0644)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"town", "init"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when town.yml exists")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestTownInitForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	configPath := filepath.Join(dir, config.TownConfigFile)
	os.WriteFile(configPath, []byte("old content"), 0644)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"town", "init", "--name", "new-town", "--force"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error with --force, got %v", err)
	}

	cfg, err := config.LoadTownConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Name != "new-town" {
		t.Errorf("expected name 'new-town', got %q", cfg.Name)
	}
}

func TestTownInitSetsDescription(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"town", "init", "--name", "my-town", "--description", "A test workspace"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cfg, err := config.LoadTownConfig(filepath.Join(dir, config.TownConfigFile))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Description != "A test workspace" {
		t.Errorf("expected description 'A test workspace', got %q", cfg.Description)
	}
}

func TestTownInitConfigHasDefaults(t *testing.T) {
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"town", "init", "--name", "test"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cfg, err := config.LoadTownConfig(filepath.Join(dir, config.TownConfigFile))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Mayor.Model != config.DefaultMayorModel {
		t.Errorf("expected mayor.model %q, got %q", config.DefaultMayorModel, cfg.Mayor.Model)
	}

	if cfg.Defaults.AgentModel != config.DefaultAgentModel {
		t.Errorf("expected defaults.agent_model %q, got %q", config.DefaultAgentModel, cfg.Defaults.AgentModel)
	}
}
