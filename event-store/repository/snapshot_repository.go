package repository

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/eyupaydin41/event-store/model"
)

type SnapshotRepository struct {
	conn driver.Conn
}

func NewSnapshotRepository(conn driver.Conn) *SnapshotRepository {
	return &SnapshotRepository{conn: conn}
}

// CreateTable - Snapshot tablosunu oluşturur
func (r *SnapshotRepository) CreateTable() error {
	ctx := context.Background()
	query := `
		CREATE TABLE IF NOT EXISTS snapshots (
			id String,
			aggregate_id String,
			version UInt32,
			state String,
			created_at DateTime
		) ENGINE = ReplacingMergeTree(created_at)
		ORDER BY (aggregate_id, version)
	`
	return r.conn.Exec(ctx, query)
}

// SaveSnapshot - Snapshot'ı kaydeder
func (r *SnapshotRepository) SaveSnapshot(snapshot *model.Snapshot) error {
	ctx := context.Background()
	query := `
		INSERT INTO snapshots (id, aggregate_id, version, state, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	if err := r.conn.Exec(ctx, query,
		snapshot.ID,
		snapshot.AggregateID,
		snapshot.Version,
		snapshot.State,
		snapshot.CreatedAt,
	); err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	return nil
}

// GetLatestSnapshot - Aggregate için en son snapshot'ı getirir
func (r *SnapshotRepository) GetLatestSnapshot(aggregateID string) (*model.Snapshot, error) {
	ctx := context.Background()
	query := `
		SELECT id, aggregate_id, version, state, created_at
		FROM snapshots
		WHERE aggregate_id = ?
		ORDER BY version DESC
		LIMIT 1
	`

	var snapshot model.Snapshot
	err := r.conn.QueryRow(ctx, query, aggregateID).Scan(
		&snapshot.ID,
		&snapshot.AggregateID,
		&snapshot.Version,
		&snapshot.State,
		&snapshot.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("snapshot not found for aggregate %s: %w", aggregateID, err)
	}

	return &snapshot, nil
}

// GetSnapshotAtVersion - Belirli bir version'daki snapshot'ı getirir
func (r *SnapshotRepository) GetSnapshotAtVersion(aggregateID string, version uint32) (*model.Snapshot, error) {
	ctx := context.Background()
	query := `
		SELECT id, aggregate_id, version, state, created_at
		FROM snapshots
		WHERE aggregate_id = ? AND version <= ?
		ORDER BY version DESC
		LIMIT 1
	`

	var snapshot model.Snapshot
	err := r.conn.QueryRow(ctx, query, aggregateID, version).Scan(
		&snapshot.ID,
		&snapshot.AggregateID,
		&snapshot.Version,
		&snapshot.State,
		&snapshot.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("snapshot not found for aggregate %s at version %d: %w", aggregateID, version, err)
	}

	return &snapshot, nil
}

// HasSnapshot - Aggregate için snapshot olup olmadığını kontrol eder
func (r *SnapshotRepository) HasSnapshot(aggregateID string) (bool, error) {
	ctx := context.Background()
	var count uint64

	query := "SELECT count() FROM snapshots WHERE aggregate_id = ?"
	err := r.conn.QueryRow(ctx, query, aggregateID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteOldSnapshots - Belirli bir version'dan eski snapshot'ları siler
func (r *SnapshotRepository) DeleteOldSnapshots(aggregateID string, keepLastN int) error {
	ctx := context.Background()

	// ClickHouse'da DELETE yerine ALTER TABLE ... DELETE kullanılır
	query := `
		ALTER TABLE snapshots DELETE
		WHERE aggregate_id = ? AND version NOT IN (
			SELECT version FROM snapshots
			WHERE aggregate_id = ?
			ORDER BY version DESC
			LIMIT ?
		)
	`

	return r.conn.Exec(ctx, query, aggregateID, aggregateID, keepLastN)
}
