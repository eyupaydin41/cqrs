package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/eyupaydin41/query-service/model"
	"github.com/eyupaydin41/query-service/repository"
)

// AuthService - Authentication projection'ları yöneten servis
type AuthService struct {
	authRepo *repository.AuthProjectionRepository
}

func NewAuthService(authRepo *repository.AuthProjectionRepository) *AuthService {
	return &AuthService{
		authRepo: authRepo,
	}
}

// HandleUserCreatedEvent - user.created event'ini işler
func (s *AuthService) HandleUserCreatedEvent(eventData []byte) error {
	var envelope struct {
		EventType   string          `json:"event_type"`
		AggregateID string          `json:"aggregate_id"`
		Timestamp   time.Time       `json:"timestamp"`
		Version     uint32          `json:"version"`
		Data        json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(eventData, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	var eventPayload struct {
		AggregateID  string    `json:"aggregate_id"`
		Email        string    `json:"email"`
		PasswordHash string    `json:"password_hash"`
		Timestamp    time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(envelope.Data, &eventPayload); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	// Auth projection oluştur
	auth := &model.AuthProjection{
		ID:           envelope.AggregateID,
		Email:        eventPayload.Email,
		PasswordHash: eventPayload.PasswordHash,
		Status:       "active",
		UpdatedAt:    envelope.Timestamp,
	}

	if err := s.authRepo.Upsert(auth); err != nil {
		return fmt.Errorf("failed to upsert auth projection: %w", err)
	}

	log.Printf("Auth projection created: id=%s, email=%s", auth.ID, auth.Email)
	return nil
}

// HandlePasswordChangedEvent - user.password.changed event'ini işler
func (s *AuthService) HandlePasswordChangedEvent(eventData []byte) error {
	var envelope struct {
		EventType   string          `json:"event_type"`
		AggregateID string          `json:"aggregate_id"`
		Timestamp   time.Time       `json:"timestamp"`
		Version     uint32          `json:"version"`
		Data        json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(eventData, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	var eventPayload struct {
		NewPasswordHash string `json:"new_password_hash"`
	}

	if err := json.Unmarshal(envelope.Data, &eventPayload); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	if err := s.authRepo.UpdatePassword(envelope.AggregateID, eventPayload.NewPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	log.Printf("Auth projection password updated: id=%s", envelope.AggregateID)
	return nil
}

// HandleEmailChangedEvent - user.email.changed event'ini işler
func (s *AuthService) HandleEmailChangedEvent(eventData []byte) error {
	var envelope struct {
		EventType   string          `json:"event_type"`
		AggregateID string          `json:"aggregate_id"`
		Timestamp   time.Time       `json:"timestamp"`
		Version     uint32          `json:"version"`
		Data        json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(eventData, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	var eventPayload struct {
		NewEmail string `json:"new_email"`
	}

	if err := json.Unmarshal(envelope.Data, &eventPayload); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	if err := s.authRepo.UpdateEmail(envelope.AggregateID, eventPayload.NewEmail); err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	log.Printf("Auth projection email updated: id=%s, new_email=%s", envelope.AggregateID, eventPayload.NewEmail)
	return nil
}

// HandleUserDeactivatedEvent - user.deactivated event'ini işler
func (s *AuthService) HandleUserDeactivatedEvent(eventData []byte) error {
	var envelope struct {
		EventType   string          `json:"event_type"`
		AggregateID string          `json:"aggregate_id"`
		Timestamp   time.Time       `json:"timestamp"`
		Version     uint32          `json:"version"`
		Data        json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(eventData, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	if err := s.authRepo.UpdateStatus(envelope.AggregateID, "deactivated"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	log.Printf("Auth projection deactivated: id=%s", envelope.AggregateID)
	return nil
}

// FindByEmail - Email'e göre auth projection bulur
func (s *AuthService) FindByEmail(email string) (*model.AuthProjection, error) {
	return s.authRepo.FindByEmail(email)
}
