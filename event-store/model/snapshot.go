package model

import "time"

// Snapshot - Aggregate'in belirli bir andaki state'ini tutar
type Snapshot struct {
	ID          string    `json:"id" ch:"id"`
	AggregateID string    `json:"aggregate_id" ch:"aggregate_id"`
	Version     uint32    `json:"version" ch:"version"`
	State       string    `json:"state" ch:"state"` // JSON olarak serialize edilmiş state
	CreatedAt   time.Time `json:"created_at" ch:"created_at"`
}
