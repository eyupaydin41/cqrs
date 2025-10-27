package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// UserAggregate - User'ın anlık durumunu temsil eder
type UserAggregate struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   uint32    `json:"version"`

	// Event history (debugging için)
	EventCount int `json:"event_count"`
}

// NewUserAggregate - Boş user aggregate oluşturur
func NewUserAggregate() *UserAggregate {
	return &UserAggregate{
		Status: "unknown",
	}
}

// ApplyEvent - Bir event'i state'e uygular
func (u *UserAggregate) ApplyEvent(event *Event) error {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(event.Payload), &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	u.Version = event.Version
	u.EventCount++

	// Event type'a göre state'i güncelle
	switch event.EventType {
	case "user.created":
		return u.applyUserCreated(eventData, event.Timestamp)
	case "user.updated":
		return u.applyUserUpdated(eventData, event.Timestamp)
	case "user.deleted":
		return u.applyUserDeleted(eventData, event.Timestamp)
	case "user.email.changed":
		return u.applyEmailChanged(eventData, event.Timestamp)
	default:
		// Unknown event - skip
		return nil
	}
}

func (u *UserAggregate) applyUserCreated(data map[string]interface{}, timestamp time.Time) error {
	if id, ok := data["aggregate_id"].(string); ok {
		u.ID = id
	} else if id, ok := data["id"].(string); ok {
		u.ID = id
	}
	if email, ok := data["email"].(string); ok {
		u.Email = email
	}

	u.Status = "active"
	u.CreatedAt = timestamp
	u.UpdatedAt = timestamp

	return nil
}

func (u *UserAggregate) applyUserUpdated(data map[string]interface{}, timestamp time.Time) error {
	if email, ok := data["email"].(string); ok && email != "" {
		u.Email = email
	}

	u.UpdatedAt = timestamp
	return nil
}

func (u *UserAggregate) applyUserDeleted(data map[string]interface{}, timestamp time.Time) error {
	u.Status = "deleted"
	u.UpdatedAt = timestamp
	return nil
}

func (u *UserAggregate) applyEmailChanged(data map[string]interface{}, timestamp time.Time) error {
	if newEmail, ok := data["new_email"].(string); ok {
		u.Email = newEmail
	}

	u.UpdatedAt = timestamp
	return nil
}
