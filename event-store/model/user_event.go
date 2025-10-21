package model

import "time"

type Event struct {
	ID          string    `json:"id"`
	EventType   string    `json:"event_type"`
	AggregateID string    `json:"aggregate_id"`
	Payload     string    `json:"payload"`
	Timestamp   time.Time `json:"timestamp"`
	Version     uint32    `json:"version"`
}

type EventFilter struct {
	EventType   string    `json:"event_type,omitempty"`
	AggregateID string    `json:"aggregate_id,omitempty"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Limit       int       `json:"limit,omitempty"`
	Offset      int       `json:"offset,omitempty"`
}
