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
		log.Printf("Failed to unmarshal envelope: %v", err)
		return err
	}

	// "type" field'ını al (producer format: {"type": "...", "data": {...}})
	eventType, ok := envelope["type"].(string)
	if !ok {
		log.Println("missing type field in message")
		return nil
	}

	// "data" field'ından event bilgilerini al
	dataMap, ok := envelope["data"].(map[string]interface{})
	if !ok {
		log.Println("missing or invalid data field in message")
		return nil
	}

	// AggregateID'yi data içinden al
	aggregateID, _ := dataMap["aggregate_id"].(string)
	if aggregateID == "" {
		log.Println("missing aggregate_id in data")
		return nil
	}

	// Timestamp'i data içinden al
	var timestamp time.Time
	if ts, ok := dataMap["timestamp"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			timestamp = parsed
		}
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Version'ı data içinden al
	version := uint32(0)
	if v, ok := dataMap["version"].(float64); ok {
		version = uint32(v)
	}

	// Tüm event'i (data kısmını) payload olarak kaydet
	payloadBytes, err := json.Marshal(dataMap)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return err
	}

	event := &model.Event{
		ID:          uuid.New().String(),
		EventType:   eventType,
		AggregateID: aggregateID,
		Payload:     string(payloadBytes),
		Timestamp:   timestamp,
		Version:     version,
	}

	log.Printf("Event Store: Saving event %s for aggregate %s (version %d)", eventType, aggregateID, version)

	if err := c.service.SaveEvent(event); err != nil {
		log.Printf("Failed to save event: %v", err)
		return err
	}

	log.Printf("Event Store: Successfully saved event %s", eventType)
	return nil
}

func (c *EventStoreConsumer) Close() {
	c.consumer.Close()
}
