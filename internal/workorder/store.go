package workorder

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/sqlite"
)

type Store struct {
	db *sql.DB
}

type CreateInput struct {
	Title       string
	Description string
	BeadID      string
	Status      Status
}

type ListOptions struct {
	Status *Status
	BeadID string
	Limit  int
}

func OpenStore(path string) (*Store, error) {
	db, err := openSQLite(path)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.ensureSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Create(ctx context.Context, input CreateInput) (WorkOrder, error) {
	if input.Title == "" {
		return WorkOrder{}, fmt.Errorf("title is required")
	}
	if input.Description == "" {
		return WorkOrder{}, fmt.Errorf("description is required")
	}
	if !input.Status.IsValid() {
		return WorkOrder{}, fmt.Errorf("invalid status %q", input.Status)
	}

	now := time.Now().UTC()
	order := WorkOrder{
		Title:       input.Title,
		Description: input.Description,
		BeadID:      input.BeadID,
		Status:      input.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if input.Status == StatusInProgress {
		order.StartedAt = &now
	}
	if input.Status == StatusDone {
		order.StartedAt = &now
		order.CompletedAt = &now
	}

	stmt := workOrders.INSERT(
		woTitle,
		woDescription,
		woBeadID,
		woStatus,
		woCreatedAt,
		woUpdatedAt,
		woStartedAt,
		woCompletedAt,
	).VALUES(
		order.Title,
		order.Description,
		nullString(order.BeadID),
		string(order.Status),
		formatTime(order.CreatedAt),
		formatTime(order.UpdatedAt),
		nullableTime(order.StartedAt),
		nullableTime(order.CompletedAt),
	)

	result, err := stmt.ExecContext(ctx, s.db)
	if err != nil {
		return WorkOrder{}, fmt.Errorf("insert work order: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return WorkOrder{}, fmt.Errorf("read work order id: %w", err)
	}
	order.ID = id
	return order, nil
}

func (s *Store) Get(ctx context.Context, id int64) (WorkOrder, error) {
	stmt := workOrders.SELECT(
		woID,
		woTitle,
		woDescription,
		woBeadID,
		woStatus,
		woCreatedAt,
		woUpdatedAt,
		woStartedAt,
		woCompletedAt,
	).WHERE(woID.EQ(sqlite.Int64(id)))

	rows, err := stmt.Rows(ctx, s.db)
	if err != nil {
		return WorkOrder{}, fmt.Errorf("select work order: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return WorkOrder{}, sql.ErrNoRows
	}
	order, err := scanWorkOrder(rows.Rows)
	if err != nil {
		return WorkOrder{}, err
	}
	return order, nil
}

func (s *Store) List(ctx context.Context, opts ListOptions) ([]WorkOrder, error) {
	stmt := workOrders.SELECT(
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

	conditions := make([]sqlite.BoolExpression, 0, 2)
	if opts.Status != nil {
		conditions = append(conditions, woStatus.EQ(sqlite.String(string(*opts.Status))))
	}
	if opts.BeadID != "" {
		conditions = append(conditions, woBeadID.EQ(sqlite.String(opts.BeadID)))
	}
	if len(conditions) > 0 {
		var combined sqlite.BoolExpression
		for _, condition := range conditions {
			if combined == nil {
				combined = condition
				continue
			}
			combined = combined.AND(condition)
		}
		stmt = stmt.WHERE(combined)
	}
	stmt = stmt.ORDER_BY(woUpdatedAt.DESC())
	if opts.Limit > 0 {
		stmt = stmt.LIMIT(int64(opts.Limit))
	}

	rows, err := stmt.Rows(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("list work orders: %w", err)
	}
	defer rows.Close()

	var orders []WorkOrder
	for rows.Next() {
		order, err := scanWorkOrder(rows.Rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate work orders: %w", err)
	}
	return orders, nil
}

func (s *Store) UpdateStatus(ctx context.Context, id int64, next Status) (WorkOrder, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return WorkOrder{}, err
	}

	updated, err := Transition(current, next, time.Now().UTC())
	if err != nil {
		return WorkOrder{}, err
	}

	stmt := workOrders.UPDATE(
		woStatus,
		woUpdatedAt,
		woStartedAt,
		woCompletedAt,
	).SET(
		string(updated.Status),
		formatTime(updated.UpdatedAt),
		nullableTime(updated.StartedAt),
		nullableTime(updated.CompletedAt),
	).WHERE(woID.EQ(sqlite.Int64(id)))

	if _, err := stmt.ExecContext(ctx, s.db); err != nil {
		return WorkOrder{}, fmt.Errorf("update work order: %w", err)
	}
	return updated, nil
}

func scanWorkOrder(rows *sql.Rows) (WorkOrder, error) {
	var order WorkOrder
	var status string
	var createdAt string
	var updatedAt string
	var beadID sql.NullString
	var startedAt sql.NullString
	var completedAt sql.NullString

	if err := rows.Scan(
		&order.ID,
		&order.Title,
		&order.Description,
		&beadID,
		&status,
		&createdAt,
		&updatedAt,
		&startedAt,
		&completedAt,
	); err != nil {
		return WorkOrder{}, fmt.Errorf("scan work order: %w", err)
	}

	parsedStatus, err := ParseStatus(status)
	if err != nil {
		return WorkOrder{}, err
	}
	order.Status = parsedStatus
	order.BeadID = beadID.String

	order.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return WorkOrder{}, fmt.Errorf("parse created_at: %w", err)
	}
	order.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return WorkOrder{}, fmt.Errorf("parse updated_at: %w", err)
	}

	if startedAt.Valid {
		parsed, err := parseTime(startedAt.String)
		if err != nil {
			return WorkOrder{}, fmt.Errorf("parse started_at: %w", err)
		}
		order.StartedAt = &parsed
	}
	if completedAt.Valid {
		parsed, err := parseTime(completedAt.String)
		if err != nil {
			return WorkOrder{}, fmt.Errorf("parse completed_at: %w", err)
		}
		order.CompletedAt = &parsed
	}
	return order, nil
}

func nullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullableTime(value *time.Time) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: formatTime(*value), Valid: true}
}
