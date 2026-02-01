package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	CampConfigFile       = "camp.yml"
	CurrentVersion       = 1
	DefaultOperatorModel = "openai/gpt-5.2-codex"
	DefaultAgentModel    = "openai/gpt-5.2-codex"
	DefaultPlanningTool  = "opencode"
)

type CampConfig struct {
	Version     int            `yaml:"version"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description,omitempty"`
	Operator    OperatorConfig `yaml:"operator,omitempty"`
	Defaults    Defaults       `yaml:"defaults,omitempty"`
}

type OperatorConfig struct {
	Model              string `yaml:"model,omitempty"`
	PlanningTool       string `yaml:"planning_tool,omitempty"`        // "claude" or "opencode"
	PlanningPromptFile string `yaml:"planning_prompt_file,omitempty"` // custom prompt file path
}

type Defaults struct {
	AgentModel string `yaml:"agent_model,omitempty"`
}

func NewCampConfig(name string) *CampConfig {
	return &CampConfig{
		Version: CurrentVersion,
		Name:    name,
		Operator: OperatorConfig{
			Model:        DefaultOperatorModel,
			PlanningTool: DefaultPlanningTool,
		},
		Defaults: Defaults{
			AgentModel: DefaultAgentModel,
		},
	}
}

func (c *CampConfig) Write(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func LoadCampConfig(path string) (*CampConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config CampConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &config, nil
}
