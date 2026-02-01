package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	TownConfigFile      = "town.yml"
	CurrentVersion      = 1
	DefaultRigsDir      = "./rigs"
	DefaultMayorModel   = "openai/gpt-5.2-codex"
	DefaultAgentModel   = "openai/gpt-5.2-codex"
	DefaultPlanningTool = "opencode"
)

type TownConfig struct {
	Version     int         `yaml:"version"`
	Name        string      `yaml:"name"`
	Description string      `yaml:"description,omitempty"`
	RigsDir     string      `yaml:"rigs_dir,omitempty"`
	Mayor       MayorConfig `yaml:"mayor,omitempty"`
	Defaults    Defaults    `yaml:"defaults,omitempty"`
}

type MayorConfig struct {
	Model              string `yaml:"model,omitempty"`
	PlanningTool       string `yaml:"planning_tool,omitempty"`        // "claude" or "opencode"
	PlanningPromptFile string `yaml:"planning_prompt_file,omitempty"` // custom prompt file path
}

type Defaults struct {
	AgentModel string `yaml:"agent_model,omitempty"`
}

func NewTownConfig(name string) *TownConfig {
	return &TownConfig{
		Version: CurrentVersion,
		Name:    name,
		RigsDir: DefaultRigsDir,
		Mayor: MayorConfig{
			Model:        DefaultMayorModel,
			PlanningTool: DefaultPlanningTool,
		},
		Defaults: Defaults{
			AgentModel: DefaultAgentModel,
		},
	}
}

func (c *TownConfig) Write(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func LoadTownConfig(path string) (*TownConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config TownConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &config, nil
}
