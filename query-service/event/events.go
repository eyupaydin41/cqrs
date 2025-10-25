package event

import "time"

// UserLoginRecordedEvent - Login kaydedildiğinde publish edilir
type UserLoginRecordedEvent struct {
	EventType   string    `json:"event_type"`
	AggregateID string    `json:"aggregate_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     uint32    `json:"version"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
}
