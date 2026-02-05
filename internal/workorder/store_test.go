package workorder

import (
	"context"
	"path/filepath"
	"testing"
)

func TestStoreCreateAndUpdate(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "workorders.db")

	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	created, err := store.Create(context.Background(), CreateInput{
		Title:       "Implement work orders",
		Description: "Add persistence and CLI",
		BeadID:      "cn-ta1.1",
		Status:      StatusReady,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected id to be set")
	}
	if created.Status != StatusReady {
		t.Fatalf("expected status ready, got %s", created.Status)
	}

	updated, err := store.UpdateStatus(context.Background(), created.ID, StatusInProgress)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.Status != StatusInProgress {
		t.Fatalf("expected status in_progress, got %s", updated.Status)
	}
	if updated.StartedAt == nil {
		t.Fatal("expected started_at to be set")
	}
}

func TestStoreListFilters(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "workorders.db")

	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	_, err = store.Create(context.Background(), CreateInput{
		Title:       "Ready work",
		Description: "Ready description",
		Status:      StatusReady,
	})
	if err != nil {
		t.Fatalf("create ready work order: %v", err)
	}
	_, err = store.Create(context.Background(), CreateInput{
		Title:       "Blocked work",
		Description: "Blocked description",
		Status:      StatusBlocked,
	})
	if err != nil {
		t.Fatalf("create blocked work order: %v", err)
	}

	status := StatusBlocked
	orders, err := store.List(context.Background(), ListOptions{Status: &status})
	if err != nil {
		t.Fatalf("list work orders: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 blocked work order, got %d", len(orders))
	}
	if orders[0].Status != StatusBlocked {
		t.Fatalf("expected blocked status, got %s", orders[0].Status)
	}
}
