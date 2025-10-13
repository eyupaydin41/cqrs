package service

import (
	"time"

	"github.com/eyupaydin41/auth-service/config"
	"github.com/eyupaydin41/auth-service/event"
	"github.com/eyupaydin41/auth-service/model"
	"github.com/eyupaydin41/auth-service/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo     *repository.UserRepository
	producer *event.KafkaProducer
}

func NewUserService(repo *repository.UserRepository, producer *event.KafkaProducer) *UserService {
	return &UserService{
		repo:     repo,
		producer: producer,
	}
}

func (s *UserService) Register(email, password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	id := uuid.New().String()
	user := &model.User{
		ID:        id,
		Email:     email,
		Password:  string(hashed),
		CreatedAt: time.Now(),
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return "", err
	}

	s.producer.Publish("UserRegistered", map[string]string{
		"id":         user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	})

	return id, nil
}

func (s *UserService) Login(email, password string) (string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", err
	}

	token, err := config.GenerateJWT(user.ID)
	if err != nil {
		return "", err
	}

	s.producer.Publish("UserLoggedIn", map[string]string{
		"id":    user.ID,
		"email": user.Email,
	})

	return token, nil
}
