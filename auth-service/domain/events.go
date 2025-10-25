package domain

import "time"

// DomainEvent interface - Tüm domain event'ları bunu implement eder
type DomainEvent interface {
	GetEventType() string
	GetAggregateID() string
	GetTimestamp() time.Time
	GetVersion() uint32
}

// BaseEvent - Tüm event'ların ortak alanları
type BaseEvent struct {
	AggregateID string    `json:"aggregate_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     uint32    `json:"version"`
}

func (e BaseEvent) GetAggregateID() string {
	return e.AggregateID
}

func (e BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e BaseEvent) GetVersion() uint32 {
	return e.Version
}

// UserCreatedEvent - User oluşturulduğunda
type UserCreatedEvent struct {
	BaseEvent
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

func (e UserCreatedEvent) GetEventType() string {
	return "user.created"
}

// PasswordChangedEvent - Şifre değiştirildiğinde
type PasswordChangedEvent struct {
	BaseEvent
	NewPasswordHash string `json:"new_password_hash"`
}

func (e PasswordChangedEvent) GetEventType() string {
	return "user.password.changed"
}

// EmailChangedEvent - Email değiştirildiğinde
type EmailChangedEvent struct {
	BaseEvent
	OldEmail string `json:"old_email"`
	NewEmail string `json:"new_email"`
}

func (e EmailChangedEvent) GetEventType() string {
	return "user.email.changed"
}

// UserDeactivatedEvent - User deaktive edildiğinde
type UserDeactivatedEvent struct {
	BaseEvent
	Reason string `json:"reason"`
}

func (e UserDeactivatedEvent) GetEventType() string {
	return "user.deactivated"
}

// UserLoginRecordedEvent - Login kaydedildiğinde
type UserLoginRecordedEvent struct {
	BaseEvent
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

func (e UserLoginRecordedEvent) GetEventType() string {
	return "user.login.recorded"
}
