package integration_tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	authEvent "github.com/eyupaydin41/auth-service/event"
	authModel "github.com/eyupaydin41/auth-service/model"
	"github.com/eyupaydin41/event-store/consumer"
	"github.com/eyupaydin41/event-store/repository"
	"github.com/eyupaydin41/event-store/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventFlow - auth-service ve event-store kodlarını test eder
func TestEventFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// ClickHouse container ve bağlantı
	t.Log("Starting ClickHouse container...")
	clickhouseContainer, conn, err := SetupClickHouse(ctx)
	require.NoError(t, err)
	defer clickhouseContainer.Terminate(ctx)
	defer conn.Close()
	t.Log("ClickHouse ready")

	// Kafka container
	t.Log("Starting Kafka container...")
	kafkaContainer, broker, err := SetupKafka(ctx)
	require.NoError(t, err)
	defer kafkaContainer.Terminate(ctx)
	t.Log("Kafka ready at:", broker)

	// Auth-service'in Producer'ı
	t.Log("Creating auth-service KafkaProducer...")
	producer := authEvent.NewKafkaProducer(broker, "user-events")
	defer producer.Close()
	t.Log("Auth-service Producer created")

	// Event-store'un service ve Consumer'ı
	t.Log("Creating event-store components...")
	eventRepo := repository.NewEventRepository(conn)
	eventService := service.NewEventService(eventRepo)
	eventConsumer := consumer.NewEventStoreConsumer(broker, "test-group", "user-events", eventService)
	t.Log("Event-store Consumer created")

	// Consumer'ı goroutine'de başlat
	consumerDone := make(chan bool)
	go func() {
		defer close(consumerDone)

		go eventConsumer.Start()

		<-ctx.Done()
		eventConsumer.Close()
	}()
	defer func() {
		cancel()
		<-consumerDone
		t.Log("Consumer stopped")
	}()

	t.Log("Waiting for consumer to be ready...")
	time.Sleep(5 * time.Second)

	// Test verisi - Gerçek User modeli
	t.Log("Publishing user.created event via auth-service producer...")
	testUser := authModel.User{
		ID:    uuid.New().String(),
		Email: "integration-test@example.com",
	}

	// Auth-service'in Publish metodunu kullan
	producer.Publish("user.created", map[string]interface{}{
		"id":    testUser.ID,
		"email": testUser.Email,
	})

	t.Log("Event published")

	// Consumer'ın işlemesini bekle
	t.Log("Waiting for event-store to process event...")
	time.Sleep(10 * time.Second)

	// ClickHouse'dan EventService ile sorgula
	t.Log("Querying ClickHouse via event-store service...")

	// Event count kontrolü
	count, err := eventService.CountEvents()
	require.NoError(t, err)
	assert.Greater(t, count, uint64(0), "Should have at least 1 event")

	// Aggregate ID ile eventi sorgula
	events, err := eventService.GetEventsByAggregateID(testUser.ID, 0)
	require.NoError(t, err)
	require.Len(t, events, 1, "Should have exactly 1 event for this user")

	// Event detaylarını kontrol et
	event := events[0]
	assert.Equal(t, "user.created", event.EventType)
	assert.Equal(t, testUser.ID, event.AggregateID)

	// Payload içindeki email'i kontrol et
	var payload map[string]interface{}
	err = json.Unmarshal([]byte(event.Payload), &payload)
	require.NoError(t, err)

	if data, ok := payload["data"].(map[string]interface{}); ok {
		assert.Equal(t, testUser.Email, data["email"])
	}

	t.Log("🎉 Integration Test PASSED!")
	t.Log("   ✓ Auth-service Producer worked")
	t.Log("   ✓ Event-store Consumer worked")
	t.Log("   ✓ Event-store Service worked")
	t.Log("   ✓ Full event-sourcing flow verified")
}
