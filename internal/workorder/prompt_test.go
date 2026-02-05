package workorder

import (
	"strings"
	"testing"
)

func TestRenderPrompt(t *testing.T) {
	order := WorkOrder{
		ID:          42,
		Title:       "Implement work orders",
		Description: "Add persistence and CLI",
		Status:      StatusReady,
		BeadID:      "cn-ta1.1",
	}

	prompt, err := RenderPrompt(PromptData{
		RolePrompt:      "Role content",
		WorkOrder:       order,
		BeadTitle:       "Bead title",
		BeadDescription: "Bead description",
	})
	if err != nil {
		t.Fatalf("render prompt: %v", err)
	}
	if prompt == "" {
		t.Fatal("expected prompt to be rendered")
	}
	if !strings.Contains(prompt, order.Title) {
		t.Fatal("expected prompt to include work order title")
	}
	if !strings.Contains(prompt, "Role content") {
		t.Fatal("expected prompt to include role content")
	}
}
