package consumer

import (
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/eyupaydin41/event-store/model"
	"github.com/eyupaydin41/event-store/service"
	"github.com/google/uuid"
)

type EventStoreConsumer struct {
	consumer *kafka.Consumer
	topic    string
	service  *service.EventService
}

func NewEventStoreConsumer(broker, group, topic string, service *service.EventService) *EventStoreConsumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          group,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}

	if err := c.SubscribeTopics([]string{topic}, nil); err != nil {
		log.Fatalf("failed to subscribe topic: %v", err)
	}

	log.Printf("event store consumer subscribed to topic: %s", topic)

	return &EventStoreConsumer{
		consumer: c,
		topic:    topic,
		service:  service,
	}
}

func (c *EventStoreConsumer) Start() {
	log.Println("event store consumer started")
	for {
		msg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("consumer error: %v", err)
			continue
		}

		if err := c.handleEvent(msg.Value); err != nil {
			log.Printf("failed to handle event: %v", err)
		}
	}
}

func (c *EventStoreConsumer) handleEvent(eventData []byte) error {
	var envelope map[string]interface{}
	if err := json.Unmarshal(eventData, &envelope); err != nil {
		return err
	}

	eventType, ok := envelope["type"].(string)
	if !ok {
		log.Println("missing event type in message")
		return nil
	}

	aggregateID := ""
	if data, ok := envelope["data"].(map[string]interface{}); ok {
		if id, ok := data["id"].(string); ok {
			aggregateID = id
		} else if userID, ok := data["user_id"].(string); ok {
			aggregateID = userID
		}
	}

	event := &model.Event{
		ID:          uuid.New().String(),
		EventType:   eventType,
		AggregateID: aggregateID,
		Payload:     string(eventData),
		Timestamp:   time.Now(),
	}

	if err := c.service.SaveEvent(event); err != nil {
		return err
	}

	return nil
}

func (c *EventStoreConsumer) Close() {
	c.consumer.Close()
}
