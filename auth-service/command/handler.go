package command

import (
	"fmt"
	"log"

	"github.com/eyupaydin41/auth-service/domain"
	"github.com/eyupaydin41/auth-service/event"
	grpcclient "github.com/eyupaydin41/auth-service/grpc"
)

// CommandHandler - Command'ları işler ve aggregate üzerinde çalışır
type CommandHandler struct {
	publisher        *event.KafkaProducer
	eventStoreClient *grpcclient.EventStoreClient // gRPC client (yeni!)
}

// NewCommandHandler - Yeni command handler oluşturur
func NewCommandHandler(publisher *event.KafkaProducer, eventStoreClient *grpcclient.EventStoreClient) *CommandHandler {
	return &CommandHandler{
		publisher:        publisher,
		eventStoreClient: eventStoreClient,
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
// Event Sourcing ile: Aggregate'i history'den reconstruct eder
func (h *CommandHandler) HandleChangePassword(cmd ChangePasswordCommand) error {
	log.Printf("Handling ChangePassword command for user: %s", cmd.UserID)

	// 1. Event history'yi gRPC ile çek
	// HTTP'de: http.Get("http://event-store:8090/events/aggregate/" + userID)
	// gRPC'de: client.GetAggregateEvents(...)
	log.Printf("🔄 Loading aggregate %s from event-store via gRPC...", cmd.UserID)
	events, err := h.eventStoreClient.GetAggregateHistory(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to load aggregate history: %w", err)
	}

	if len(events) == 0 {
		return fmt.Errorf("user not found: %s", cmd.UserID)
	}

	// 2. Aggregate oluştur
	aggregate := domain.NewUserAggregate(cmd.UserID)

	// 3. History'den state'i reconstruct et
	log.Printf("📦 Reconstructing aggregate from %d events", len(events))
	aggregate.LoadFromHistory(events)

	log.Printf("✅ Aggregate loaded: Status=%s, Email=%s", aggregate.Status, aggregate.Email)

	// 4. Command'ı uygula
	err = aggregate.ChangePassword(cmd.OldPassword, cmd.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	// 5. Yeni event'leri Kafka'ya publish et
	h.publishEvents(aggregate.GetUncommittedChanges())
	aggregate.MarkChangesAsCommitted()

	log.Printf("Password changed successfully for user: %s", cmd.UserID)
	return nil
}

// HandleChangeEmail - Email değiştirme command'ını işler
// Event Sourcing ile: Aggregate'i history'den reconstruct eder
func (h *CommandHandler) HandleChangeEmail(cmd ChangeEmailCommand) error {
	log.Printf("Handling ChangeEmail command for user: %s", cmd.UserID)

	// 1. Event history'yi gRPC ile çek
	log.Printf("🔄 Loading aggregate %s from event-store via gRPC...", cmd.UserID)
	events, err := h.eventStoreClient.GetAggregateHistory(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to load aggregate history: %w", err)
	}

	if len(events) == 0 {
		return fmt.Errorf("user not found: %s", cmd.UserID)
	}

	// 2. Aggregate oluştur ve history'den load et
	aggregate := domain.NewUserAggregate(cmd.UserID)
	log.Printf("📦 Reconstructing aggregate from %d events", len(events))
	aggregate.LoadFromHistory(events)
	log.Printf("✅ Aggregate loaded: Status=%s, Email=%s", aggregate.Status, aggregate.Email)

	// 3. Command'ı uygula
	err = aggregate.ChangeEmail(cmd.NewEmail)
	if err != nil {
		return fmt.Errorf("failed to change email: %w", err)
	}

	// 4. Event'leri Kafka'ya publish et
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
