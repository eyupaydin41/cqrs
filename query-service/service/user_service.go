package service

import (
	"encoding/json"
	"log"
	"time"

	"github.com/eyupaydin41/query-service/model"
	"github.com/eyupaydin41/query-service/repository"
	"github.com/google/uuid"
)

type UserService struct {
	repo             *repository.UserRepository
	loginHistoryRepo *repository.LoginHistoryRepository
}

func NewUserService(repo *repository.UserRepository, loginHistoryRepo *repository.LoginHistoryRepository) *UserService {
	return &UserService{
		repo:             repo,
		loginHistoryRepo: loginHistoryRepo,
	}
}

func (s *UserService) HandleUserRegisteredEvent(eventData []byte) {
	var payload map[string]interface{}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		log.Println("failed to parse event:", err)
		return
	}

	dataField, ok := payload["data"].(map[string]interface{})
	if !ok {
		log.Println("missing 'data' field in event")
		return
	}

	id, ok := dataField["id"].(string)
	if !ok || id == "" {
		id = uuid.New().String()
	}

	email, ok := dataField["email"].(string)
	if !ok {
		email = "unknown@example.com"
	}

	createdAt := time.Now()
	if createdAtStr, ok := dataField["created_at"].(string); ok && createdAtStr != "" {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			createdAt = t
		} else {
			log.Println("invalid created_at format, using current time")
		}
	}

	user := &model.User{
		ID:        id,
		Email:     email,
		CreatedAt: createdAt,
	}

	if err := s.repo.Create(user); err != nil {
		log.Println("failed to insert user:", err)
		return
	}

	log.Printf("User inserted: id=%s, email=%s\n", user.ID, user.Email)
}

func (s *UserService) HandleUserLoggedInEvent(eventData []byte) {
	var payload map[string]interface{}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		log.Println("failed to parse event:", err)
		return
	}

	dataField, ok := payload["data"].(map[string]interface{})
	if !ok {
		log.Println("missing 'data' field in event")
		return
	}

	id, ok := dataField["id"].(string)
	if !ok || id == "" {
		log.Println("missing or invalid user id in login event")
		return
	}

	email, ok := dataField["email"].(string)
	if !ok {
		email = "unknown@example.com"
	}

	loginHistory := &model.LoginHistory{
		ID:        uuid.New().String(),
		UserID:    id,
		Email:     email,
		LoginAt:   time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.loginHistoryRepo.Create(loginHistory); err != nil {
		log.Println("failed to insert login history:", err)
		return
	}

	log.Printf("User logged in: id=%s, email=%s, login_history_id=%s\n", id, email, loginHistory.ID)
}
