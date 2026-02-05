package workorder

import (
	"testing"
	"time"
)

func TestParseStatus(t *testing.T) {
	status, err := ParseStatus("in_progress")
	if err != nil {
		t.Fatalf("expected valid status, got %v", err)
	}
	if status != StatusInProgress {
		t.Fatalf("expected %q, got %q", StatusInProgress, status)
	}

	if _, err := ParseStatus("nope"); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestTransitionRules(t *testing.T) {
	tests := []struct {
		from    Status
		to      Status
		allowed bool
	}{
		{StatusDraft, StatusReady, true},
		{StatusDraft, StatusInProgress, false},
		{StatusReady, StatusInProgress, true},
		{StatusReady, StatusDone, false},
		{StatusInProgress, StatusDone, true},
		{StatusBlocked, StatusReady, true},
		{StatusDone, StatusReady, false},
	}

	for _, tt := range tests {
		if CanTransition(tt.from, tt.to) != tt.allowed {
			t.Fatalf("expected transition %s -> %s allowed=%v", tt.from, tt.to, tt.allowed)
		}
	}
}

func TestTransitionSetsTimestamps(t *testing.T) {
	now := time.Now().UTC()
	order := WorkOrder{Status: StatusReady}

	updated, err := Transition(order, StatusInProgress, now)
	if err != nil {
		t.Fatalf("expected transition, got %v", err)
	}
	if updated.StartedAt == nil {
		t.Fatal("expected started_at to be set")
	}

	completed, err := Transition(updated, StatusDone, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("expected transition to done, got %v", err)
	}
	if completed.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}
}
