package workorder

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-jet/jet/v2/sqlite"
	_ "modernc.org/sqlite"
)

const (
	workOrdersTable = "work_orders"
)

var (
	woID          = sqlite.IntegerColumn("id")
	woTitle       = sqlite.StringColumn("title")
	woDescription = sqlite.StringColumn("description")
	woBeadID      = sqlite.StringColumn("bead_id")
	woStatus      = sqlite.StringColumn("status")
	woCreatedAt   = sqlite.StringColumn("created_at")
	woUpdatedAt   = sqlite.StringColumn("updated_at")
	woStartedAt   = sqlite.StringColumn("started_at")
	woCompletedAt = sqlite.StringColumn("completed_at")

	workOrders = sqlite.NewTable("", workOrdersTable, "",
		woID,
		woTitle,
		woDescription,
		woBeadID,
		woStatus,
		woCreatedAt,
		woUpdatedAt,
		woStartedAt,
		woCompletedAt,
	)
)

func openSQLite(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create workorder directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return db, nil
}

func (s *Store) ensureSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS work_orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    bead_id TEXT,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT
);
CREATE INDEX IF NOT EXISTS work_orders_status_idx ON work_orders(status);
CREATE INDEX IF NOT EXISTS work_orders_bead_idx ON work_orders(bead_id);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("ensure work order schema: %w", err)
	}
	return nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return parsed, nil
	}
	parsed, err = time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}
