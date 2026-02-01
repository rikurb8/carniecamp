package dashboard

import "testing"

func TestParseBdIssues(t *testing.T) {
	input := []byte(`[
		{
			"id": "cn-1",
			"title": "Test issue",
			"description": "Something to do",
			"status": "open",
			"priority": 2,
			"issue_type": "task",
			"owner": "tester@example.com",
			"updated_at": "2026-02-01T10:00:00Z",
			"created_at": "2026-01-31T10:00:00Z"
		}
	]`)

	issues, err := parseBdIssues(input)
	if err != nil {
		t.Fatalf("parseBdIssues error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	issue := issues[0]
	if issue.ID != "cn-1" {
		t.Fatalf("expected id cn-1, got %s", issue.ID)
	}
	if issue.Priority != 2 {
		t.Fatalf("expected priority 2, got %d", issue.Priority)
	}
}
