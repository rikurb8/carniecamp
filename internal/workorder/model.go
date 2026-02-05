package workorder

import (
	"fmt"
	"strings"
	"time"
)

type Status string

const (
	StatusDraft      Status = "draft"
	StatusReady      Status = "ready"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
	StatusCanceled   Status = "canceled"
)

var validStatuses = []Status{
	StatusDraft,
	StatusReady,
	StatusInProgress,
	StatusBlocked,
	StatusDone,
	StatusCanceled,
}

func ValidStatuses() []Status {
	return validStatuses
}

func ParseStatus(value string) (Status, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	for _, status := range validStatuses {
		if string(status) == trimmed {
			return status, nil
		}
	}
	return Status(""), fmt.Errorf("invalid status %q", value)
}

func (s Status) IsValid() bool {
	for _, status := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

type WorkOrder struct {
	ID          int64
	Title       string
	Description string
	BeadID      string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

func CanTransition(from Status, to Status) bool {
	if from == to {
		return false
	}
	switch from {
	case StatusDraft:
		return to == StatusReady || to == StatusCanceled
	case StatusReady:
		return to == StatusInProgress || to == StatusBlocked || to == StatusCanceled
	case StatusInProgress:
		return to == StatusBlocked || to == StatusDone || to == StatusCanceled
	case StatusBlocked:
		return to == StatusReady || to == StatusInProgress || to == StatusCanceled
	case StatusDone, StatusCanceled:
		return false
	default:
		return false
	}
}

func Transition(order WorkOrder, next Status, now time.Time) (WorkOrder, error) {
	if !order.Status.IsValid() {
		return WorkOrder{}, fmt.Errorf("invalid current status %q", order.Status)
	}
	if !next.IsValid() {
		return WorkOrder{}, fmt.Errorf("invalid next status %q", next)
	}
	if !CanTransition(order.Status, next) {
		return WorkOrder{}, fmt.Errorf("cannot transition from %q to %q", order.Status, next)
	}

	order.Status = next
	order.UpdatedAt = now
	if next == StatusInProgress && order.StartedAt == nil {
		started := now
		order.StartedAt = &started
	}
	if next == StatusDone && order.CompletedAt == nil {
		completed := now
		order.CompletedAt = &completed
		if order.StartedAt == nil {
			started := now
			order.StartedAt = &started
		}
	}
	return order, nil
}
