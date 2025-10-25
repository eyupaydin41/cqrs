package command

import (
	"fmt"
	"log"

	"github.com/eyupaydin41/auth-service/domain"
	"github.com/eyupaydin41/auth-service/event"
)

// CommandHandler - Command'ları işler ve aggregate üzerinde çalışır
type CommandHandler struct {
	publisher *event.KafkaProducer
}

// NewCommandHandler - Yeni command handler oluşturur
func NewCommandHandler(publisher *event.KafkaProducer) *CommandHandler {
	return &CommandHandler{
		publisher: publisher,
	}
}

// HandleRegisterUser - User kayıt command'ını işler
func (h *CommandHandler) HandleRegisterUser(cmd RegisterUserCommand) error {
	log.Printf("RegisterUser command for user: %s", cmd.UserID)

	// 1. Yeni aggregate oluştur
	aggregate := domain.NewUserAggregate(cmd.UserID)

	// 2. Domain logic aggregate'de
	err := aggregate.Register(cmd.Email, cmd.Password)
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	// 3. Event'leri Kafka'ya publish et
	h.publishEvents(aggregate.GetUncommittedChanges())

	// 4. Aggregate'i temizle
	aggregate.MarkChangesAsCommitted()

	log.Printf("User registered successfully, events published to Kafka: %s", cmd.UserID)
	return nil
}

// HandleChangePassword - Şifre değiştirme command'ını işler
// NOT: Event Sourcing olmadan basitleştirildi - şimdilik sadece event publish ediyor
func (h *CommandHandler) HandleChangePassword(cmd ChangePasswordCommand) error {
	log.Printf("Handling ChangePassword command for user: %s", cmd.UserID)

	// Aggregate oluştur (load yerine yeni oluşturuyoruz - basitleştirilmiş versiyon)
	aggregate := domain.NewUserAggregate(cmd.UserID)

	// Domain logic
	err := aggregate.ChangePassword(cmd.OldPassword, cmd.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	// Event'leri Kafka'ya publish et
	h.publishEvents(aggregate.GetUncommittedChanges())
	aggregate.MarkChangesAsCommitted()

	log.Printf("Password changed successfully for user: %s", cmd.UserID)
	return nil
}

// HandleChangeEmail - Email değiştirme command'ını işler
// NOT: Event Sourcing olmadan basitleştirildi - şimdilik sadece event publish ediyor
func (h *CommandHandler) HandleChangeEmail(cmd ChangeEmailCommand) error {
	log.Printf("Handling ChangeEmail command for user: %s", cmd.UserID)

	// Aggregate oluştur
	aggregate := domain.NewUserAggregate(cmd.UserID)

	// Domain logic
	err := aggregate.ChangeEmail(cmd.NewEmail)
	if err != nil {
		return fmt.Errorf("failed to change email: %w", err)
	}

	// Event'leri Kafka'ya publish et
	h.publishEvents(aggregate.GetUncommittedChanges())
	aggregate.MarkChangesAsCommitted()

	log.Printf("Email changed successfully for user: %s", cmd.UserID)
	return nil
}

// publishEvents - Event'leri Kafka'ya publish eder
func (h *CommandHandler) publishEvents(events []domain.DomainEvent) {
	for _, event := range events {
		h.publisher.Publish(event.GetEventType(), event)
		log.Printf("Published event: %s for aggregate: %s", event.GetEventType(), event.GetAggregateID())
	}
}
