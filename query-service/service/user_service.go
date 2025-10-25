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

	dataField, ok := envelope["data"].(map[string]interface{})
	if !ok {
		log.Println("missing data field in event")
		return
	}

	aggregateID, _ := dataField["aggregate_id"].(string)
	email, _ := dataField["email"].(string)

	timestamp := time.Now()
	if tStr, ok := dataField["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, tStr); err == nil {
			timestamp = t
		}
	}

	if email == "" {
		email = "unknown@example.com"
	}

	user := &model.User{
		ID:        aggregateID,
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

	aggregateID, ok := dataField["aggregate_id"].(string)
	if !ok || aggregateID == "" {
		log.Println("missing or invalid aggregate_id in login event")
		return
	}

	ipAddress, _ := dataField["ip_address"].(string)
	userAgent, _ := dataField["user_agent"].(string)

	// User bilgisini repo'dan al
	user, err := s.repo.FindByID(aggregateID)
	if err != nil {
		log.Printf("user not found for login event: %s, error: %v", aggregateID, err)
		return
	}

	loginHistory := &model.LoginHistory{
		ID:        uuid.New().String(),
		UserID:    aggregateID,
		Email:     user.Email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		LoginAt:   time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.loginHistoryRepo.Create(loginHistory); err != nil {
		log.Println("failed to insert login history:", err)
		return
	}

	log.Printf("User logged in: id=%s, email=%s, ip=%s, login_history_id=%s\n", aggregateID, user.Email, ipAddress, loginHistory.ID)
}
