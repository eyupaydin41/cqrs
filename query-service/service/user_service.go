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
	var envelope map[string]interface{}
	if err := json.Unmarshal(eventData, &envelope); err != nil {
		log.Println("failed to parse event envelope:", err)
		return
	}

	aggregateID, hasAggregateID := envelope["aggregate_id"].(string)
	timestamp := time.Now()

	if timestampStr, ok := envelope["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			timestamp = t
		}
	}

	var email string
	var userID string

	if hasAggregateID {
		userID = aggregateID
		dataField, ok := envelope["data"].(map[string]interface{})
		if ok {
			if e, ok := dataField["email"].(string); ok {
				email = e
			}
			if nestedData, ok := dataField["data"].(map[string]interface{}); ok {
				if e, ok := nestedData["email"].(string); ok {
					email = e
				}
			}
		}
	}

	if email == "" {
		email = "unknown@example.com"
	}

	user := &model.User{
		ID:        userID,
		Email:     email,
		Status:    "active",
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
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
