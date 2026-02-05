package dashboard

import "testing"

func TestFilterOpenFeatures(t *testing.T) {
	data := dataState{
		Ready: []Issue{
			{ID: "feat-2", IssueType: "feature", Status: "open", Priority: 2},
			{ID: "task-1", IssueType: "task", Status: "open", Priority: 1},
		},
		InProgress: []Issue{
			{ID: "feat-1", IssueType: "feature", Status: "in_progress", Priority: 1},
		},
		Blocked: []Issue{
			{ID: "feat-3", IssueType: "feature", Status: "blocked", Priority: 3},
		},
		Closed: []Issue{
			{ID: "feat-closed", IssueType: "feature", Status: "closed", Priority: 1},
		},
	}

	features := filterOpenFeatures(data)
	if len(features) != 3 {
		t.Fatalf("expected 3 open features, got %d", len(features))
	}
	if features[0].ID != "feat-1" || features[1].ID != "feat-2" || features[2].ID != "feat-3" {
		t.Fatalf("unexpected feature order: %#v", []string{features[0].ID, features[1].ID, features[2].ID})
	}
}

func TestBuildFeatureChildren(t *testing.T) {
	issues := []Issue{
		{ID: "feat-1", IssueType: "feature"},
		{ID: "feat-2", IssueType: "feature"},
		{ID: "task-1", IssueType: "task", Priority: 2, Dependencies: []Dependency{{DependsOnID: "feat-1", Type: "parent-child"}}},
		{ID: "task-2", IssueType: "task", Priority: 1, Dependencies: []Dependency{{DependsOnID: "feat-1", Type: "parent-child"}}},
		{ID: "task-3", IssueType: "task", Priority: 3, Dependencies: []Dependency{{DependsOnID: "feat-2", Type: "parent-child"}}},
		{ID: "bug-1", IssueType: "bug", Priority: 1, Dependencies: []Dependency{{DependsOnID: "feat-1", Type: "parent-child"}}},
	}

	children := buildFeatureChildren(issues)
	feat1Children := children["feat-1"]
	if len(feat1Children) != 2 {
		t.Fatalf("expected 2 child tasks for feat-1, got %d", len(feat1Children))
	}
	if feat1Children[0].ID != "task-2" || feat1Children[1].ID != "task-1" {
		t.Fatalf("unexpected child order for feat-1: %#v", []string{feat1Children[0].ID, feat1Children[1].ID})
	}
	feat2Children := children["feat-2"]
	if len(feat2Children) != 1 || feat2Children[0].ID != "task-3" {
		t.Fatalf("unexpected child list for feat-2: %#v", feat2Children)
	}
}
