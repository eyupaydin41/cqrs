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
// Event Sourcing ile: Aggregate'i snapshot'tan reconstruct eder (PERFORMANSLI!)
func (h *CommandHandler) HandleChangePassword(cmd ChangePasswordCommand) error {
	log.Printf("Handling ChangePassword command for user: %s", cmd.UserID)

	// 1. Snapshot kullanarak aggregate'i yükle
	// Snapshot varsa: snapshot + sonraki eventler (HIZLI!)
	// Snapshot yoksa: tüm eventler (yavaş ama çalışır)
	log.Printf("🔄 Loading aggregate %s with snapshot from event-store via gRPC...", cmd.UserID)
	aggregate, err := h.eventStoreClient.GetAggregateWithSnapshot(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to load aggregate with snapshot: %w", err)
	}

	log.Printf("✅ Aggregate loaded: Status=%s, Email=%s, Version=%d", aggregate.Status, aggregate.Email, aggregate.Version)

	// 2. Command'ı uygula
	err = aggregate.ChangePassword(cmd.OldPassword, cmd.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	// 3. Yeni event'leri Kafka'ya publish et
	h.publishEvents(aggregate.GetUncommittedChanges())
	aggregate.MarkChangesAsCommitted()

	log.Printf("Password changed successfully for user: %s", cmd.UserID)
	return nil
}

// HandleChangeEmail - Email değiştirme command'ını işler
// Event Sourcing ile: Aggregate'i snapshot'tan reconstruct eder (PERFORMANSLI!)
func (h *CommandHandler) HandleChangeEmail(cmd ChangeEmailCommand) error {
	log.Printf("Handling ChangeEmail command for user: %s", cmd.UserID)

	// 1. Snapshot kullanarak aggregate'i yükle
	log.Printf("🔄 Loading aggregate %s with snapshot from event-store via gRPC...", cmd.UserID)
	aggregate, err := h.eventStoreClient.GetAggregateWithSnapshot(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to load aggregate with snapshot: %w", err)
	}

	log.Printf("✅ Aggregate loaded: Status=%s, Email=%s, Version=%d", aggregate.Status, aggregate.Email, aggregate.Version)

	// 2. Command'ı uygula
	err = aggregate.ChangeEmail(cmd.NewEmail)
	if err != nil {
		return fmt.Errorf("failed to change email: %w", err)
	}

	// 3. Event'leri Kafka'ya publish et
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
