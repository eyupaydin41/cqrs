package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/eyupaydin41/auth-service/domain"
	pb "github.com/eyupaydin41/auth-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// EventStoreClient - gRPC client wrapper
// HTTP client'a benzer ama type-safe ve daha performanslı
type EventStoreClient struct {
	conn   *grpc.ClientConn
	client pb.EventStoreServiceClient
}

// NewEventStoreClient - Client oluştur
// HTTP'de: client := &http.Client{}
func NewEventStoreClient(address string) (*EventStoreClient, error) {
	log.Printf("Connecting to Event-Store gRPC server at %s", address)

	// gRPC connection oluştur
	// HTTP'den fark: Connection pooling otomatik, persistent
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // TLS yok (development)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// Proto'dan generate edilen client'ı oluştur
	client := pb.NewEventStoreServiceClient(conn)

	return &EventStoreClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetAggregateHistory - Aggregate'in tüm event history'sini getir
// HTTP karşılığı: GET /events/aggregate/:id
func (c *EventStoreClient) GetAggregateHistory(aggregateID string) ([]domain.DomainEvent, error) {
	log.Printf("gRPC Call: GetAggregateEvents for aggregate_id=%s", aggregateID)

	// Context oluştur (timeout için)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// gRPC call yap
	// HTTP'de: resp, _ := http.Get("http://event-store:8090/events/aggregate/" + id)
	resp, err := c.client.GetAggregateEvents(ctx, &pb.GetAggregateEventsRequest{
		AggregateId: aggregateID,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	log.Printf("gRPC Response: Received %d events", len(resp.Events))

	// Proto event'leri domain event'lere dönüştür
	domainEvents := make([]domain.DomainEvent, 0, len(resp.Events))

	for _, pbEvent := range resp.Events {
		// Event type'a göre domain event oluştur
		domainEvent, err := c.pbEventToDomainEvent(pbEvent)
		if err != nil {
			log.Printf("Warning: Failed to convert event %s: %v", pbEvent.Id, err)
			continue
		}
		domainEvents = append(domainEvents, domainEvent)
	}

	return domainEvents, nil
}

// pbEventToDomainEvent - Protobuf event'i domain event'e çevir
func (c *EventStoreClient) pbEventToDomainEvent(pbEvent *pb.Event) (domain.DomainEvent, error) {
	// Timestamp parse et
	timestamp, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", pbEvent.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	// Event type'a göre struct oluştur
	switch pbEvent.EventType {
	case "user.created":
		var data struct {
			Email        string `json:"email"`
			PasswordHash string `json:"password_hash"`
		}
		if err := json.Unmarshal([]byte(pbEvent.DataJson), &data); err != nil {
			return nil, err
		}

		return domain.UserCreatedEvent{
			BaseEvent: domain.BaseEvent{
				AggregateID: pbEvent.AggregateId,
				Timestamp:   timestamp,
				Version:     uint32(pbEvent.Version),
			},
			Email:        data.Email,
			PasswordHash: data.PasswordHash,
		}, nil

	case "user.password.changed":
		var data struct {
			NewPasswordHash string `json:"new_password_hash"`
		}
		if err := json.Unmarshal([]byte(pbEvent.DataJson), &data); err != nil {
			return nil, err
		}

		return domain.PasswordChangedEvent{
			BaseEvent: domain.BaseEvent{
				AggregateID: pbEvent.AggregateId,
				Timestamp:   timestamp,
				Version:     uint32(pbEvent.Version),
			},
			NewPasswordHash: data.NewPasswordHash,
		}, nil

	case "user.email.changed":
		var data struct {
			OldEmail string `json:"old_email"`
			NewEmail string `json:"new_email"`
		}
		if err := json.Unmarshal([]byte(pbEvent.DataJson), &data); err != nil {
			return nil, err
		}

		return domain.EmailChangedEvent{
			BaseEvent: domain.BaseEvent{
				AggregateID: pbEvent.AggregateId,
				Timestamp:   timestamp,
				Version:     uint32(pbEvent.Version),
			},
			OldEmail: data.OldEmail,
			NewEmail: data.NewEmail,
		}, nil

	case "user.deactivated":
		var data struct {
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal([]byte(pbEvent.DataJson), &data); err != nil {
			return nil, err
		}

		return domain.UserDeactivatedEvent{
			BaseEvent: domain.BaseEvent{
				AggregateID: pbEvent.AggregateId,
				Timestamp:   timestamp,
				Version:     uint32(pbEvent.Version),
			},
			Reason: data.Reason,
		}, nil

	default:
		return nil, fmt.Errorf("unknown event type: %s", pbEvent.EventType)
	}
}

// Close - Connection'ı kapat
func (c *EventStoreClient) Close() error {
	return c.conn.Close()
}
