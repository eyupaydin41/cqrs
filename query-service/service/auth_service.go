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
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(eventData, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	var eventPayload struct {
		AggregateID  string    `json:"aggregate_id"`
		Email        string    `json:"email"`
		PasswordHash string    `json:"password_hash"`
		Timestamp    time.Time `json:"timestamp"`
		Version      uint32    `json:"version"`
	}

	if err := json.Unmarshal(envelope.Data, &eventPayload); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	// Auth projection oluştur
	auth := &model.AuthProjection{
		ID:           eventPayload.AggregateID,
		Email:        eventPayload.Email,
		PasswordHash: eventPayload.PasswordHash,
		Status:       "active",
		UpdatedAt:    eventPayload.Timestamp,
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
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(eventData, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	var eventPayload struct {
		AggregateID     string `json:"aggregate_id"`
		NewPasswordHash string `json:"new_password_hash"`
	}

	if err := json.Unmarshal(envelope.Data, &eventPayload); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	if err := s.authRepo.UpdatePassword(eventPayload.AggregateID, eventPayload.NewPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	log.Printf("Auth projection password updated: id=%s", eventPayload.AggregateID)
	return nil
}

// HandleEmailChangedEvent - user.email.changed event'ini işler
func (s *AuthService) HandleEmailChangedEvent(eventData []byte) error {
	var envelope struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(eventData, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	var eventPayload struct {
		AggregateID string `json:"aggregate_id"`
		NewEmail    string `json:"new_email"`
	}

	if err := json.Unmarshal(envelope.Data, &eventPayload); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	if err := s.authRepo.UpdateEmail(eventPayload.AggregateID, eventPayload.NewEmail); err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	log.Printf("Auth projection email updated: id=%s, new_email=%s", eventPayload.AggregateID, eventPayload.NewEmail)
	return nil
}

// HandleUserDeactivatedEvent - user.deactivated event'ini işler
func (s *AuthService) HandleUserDeactivatedEvent(eventData []byte) error {
	var envelope struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(eventData, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	var eventPayload struct {
		AggregateID string `json:"aggregate_id"`
	}

	if err := json.Unmarshal(envelope.Data, &eventPayload); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	if err := s.authRepo.UpdateStatus(eventPayload.AggregateID, "deactivated"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	log.Printf("Auth projection deactivated: id=%s", eventPayload.AggregateID)
	return nil
}

// FindByEmail - Email'e göre auth projection bulur
func (s *AuthService) FindByEmail(email string) (*model.AuthProjection, error) {
	return s.authRepo.FindByEmail(email)
}
