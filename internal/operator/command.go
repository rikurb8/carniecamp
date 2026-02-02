package operator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rikurb8/carnie/internal/config"
	"github.com/rikurb8/carnie/internal/session"
)

type PlanningCommand struct {
	Command string
	Tool    session.Tool
	Model   string
}

func BuildPlanningCommand(workDir string, title string, toolOverride string) (PlanningCommand, error) {
	if workDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return PlanningCommand{}, fmt.Errorf("get working directory: %w", err)
		}
		workDir = cwd
	}

	campCfg := loadCampConfig(workDir)

	planningTool := config.DefaultPlanningTool
	model := config.DefaultOperatorModel
	if campCfg != nil {
		if campCfg.Operator.PlanningTool != "" {
			planningTool = campCfg.Operator.PlanningTool
		}
		if campCfg.Operator.Model != "" {
			model = campCfg.Operator.Model
		}
	}
	if toolOverride != "" {
		planningTool = toolOverride
	}

	selectedTool := session.ParseTool(planningTool)
	model = session.NormalizeModel(selectedTool, model)

	var promptFilePath string
	if campCfg != nil {
		promptFilePath = campCfg.Operator.PlanningPromptFile
	}
	basePrompt := loadEpicPlanningPrompt(workDir, promptFilePath)
	systemPrompt := buildSystemPromptWithBase(campCfg, basePrompt)

	opts := session.Options{
		Tool:         selectedTool,
		Model:        model,
		SystemPrompt: systemPrompt,
		Prompt:       epicPlanningInitialPrompt(title),
		Interactive:  true,
	}

	command := session.Command(opts)
	if command == "" {
		return PlanningCommand{}, fmt.Errorf("build operator command")
	}

	return PlanningCommand{Command: command, Tool: selectedTool, Model: model}, nil
}

func loadCampConfig(workDir string) *config.CampConfig {
	path := filepath.Join(workDir, config.CampConfigFile)
	cfg, err := config.LoadCampConfig(path)
	if err != nil {
		return nil
	}
	return cfg
}
