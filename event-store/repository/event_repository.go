package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/eyupaydin41/event-store/model"
)

type EventRepository struct {
	conn driver.Conn
}

func NewEventRepository(conn driver.Conn) *EventRepository {
	return &EventRepository{conn: conn}
}

func (r *EventRepository) SaveEvent(event *model.Event) error {
	ctx := context.Background()
	query := `
		INSERT INTO events (id, event_type, aggregate_id, payload, timestamp, version)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	if err := r.conn.Exec(ctx, query,
		event.ID,
		event.EventType,
		event.AggregateID,
		event.Payload,
		event.Timestamp,
		event.Version,
	); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

func (r *EventRepository) GetEvents(filter model.EventFilter) ([]*model.Event, error) {
	ctx := context.Background()

	var conditions []string
	var args []interface{}

	if filter.EventType != "" {
		conditions = append(conditions, "event_type = ?")
		args = append(args, filter.EventType)
	}

	if filter.AggregateID != "" {
		conditions = append(conditions, "aggregate_id = ?")
		args = append(args, filter.AggregateID)
	}

	if !filter.StartTime.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, filter.StartTime)
	}

	if !filter.EndTime.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, filter.EndTime)
	}

	query := "SELECT id, event_type, aggregate_id, payload, timestamp, version FROM events"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY timestamp ASC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*model.Event
	for rows.Next() {
		var event model.Event
		var version uint32
		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.Payload,
			&event.Timestamp,
			&version,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		event.Version = version
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}

func (r *EventRepository) CountEvents() (uint64, error) {
	ctx := context.Background()
	var count uint64
	err := r.conn.QueryRow(ctx, "SELECT count() FROM events").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}
	return count, nil
}

func (r *EventRepository) GetLatestVersionForAggregate(aggregateID string) (uint32, error) {
	ctx := context.Background()
	var version uint32

	query := "SELECT max(version) FROM events WHERE aggregate_id = ?"
	err := r.conn.QueryRow(ctx, query, aggregateID).Scan(&version)

	if err != nil {
		return 0, nil
	}

	return version, nil
}

// GetEventsAfterVersion - Belirli bir version'dan sonraki event'leri getirir
// Snapshot'tan sonra sadece gerekli event'leri yüklemek için kullanılır
func (r *EventRepository) GetEventsAfterVersion(aggregateID string, afterVersion uint32) ([]*model.Event, error) {
	ctx := context.Background()

	query := `
		SELECT id, event_type, aggregate_id, payload, timestamp, version
		FROM events
		WHERE aggregate_id = ? AND version > ?
		ORDER BY version ASC
	`

	rows, err := r.conn.Query(ctx, query, aggregateID, afterVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to query events after version: %w", err)
	}
	defer rows.Close()

	var events []*model.Event
	for rows.Next() {
		var event model.Event
		var version uint32
		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.Payload,
			&event.Timestamp,
			&version,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		event.Version = version
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}

// SaveBatchEvents - Birden fazla event'i tek seferde kaydeder
func (r *EventRepository) SaveBatchEvents(events []*model.Event) error {
	if len(events) == 0 {
		return nil
	}

	ctx := context.Background()
	batch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO events (id, event_type, aggregate_id, payload, timestamp, version)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, event := range events {
		if err := batch.Append(
			event.ID,
			event.EventType,
			event.AggregateID,
			event.Payload,
			event.Timestamp,
			event.Version,
		); err != nil {
			return fmt.Errorf("failed to append event to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}
