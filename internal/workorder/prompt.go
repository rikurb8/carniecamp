package workorder

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/rikurb8/carnie/internal/templates"
)

type PromptData struct {
	RolePrompt         string
	WorkOrder          WorkOrder
	BeadTitle          string
	BeadDescription    string
	ProjectName        string
	ProjectDescription string
}

func RenderPrompt(data PromptData) (string, error) {
	tmplContent, err := templates.Load("workorder.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("load workorder template: %w", err)
	}

	tmpl, err := template.New("workorder").Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("parse workorder template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute workorder template: %w", err)
	}
	return buf.String(), nil
}
