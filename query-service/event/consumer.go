package event

import (
	"encoding/json"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/eyupaydin41/query-service/service"
)

type KafkaConsumer struct {
	consumer *kafka.Consumer
	topic    string
	service  *service.UserService
}

func NewKafkaConsumer(broker, group, topic string, service *service.UserService) *KafkaConsumer {
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

	return &KafkaConsumer{
		consumer: c,
		topic:    topic,
		service:  service,
	}
}

func (kc *KafkaConsumer) Start() {
	for {
		msg, err := kc.consumer.ReadMessage(-1)
		if err == nil {
			kc.handleEvent(msg.Value)
		} else {
			log.Printf("consumer error: %v", err)
		}
	}
}

func (kc *KafkaConsumer) handleEvent(eventData []byte) {
	var envelope map[string]interface{}
	if err := json.Unmarshal(eventData, &envelope); err != nil {
		log.Println("failed to parse event envelope:", err)
		return
	}

	eventType, ok := envelope["type"].(string)
	if !ok {
		log.Println("missing event type in message")
		return
	}

	switch eventType {
	case "UserRegistered":
		kc.service.HandleUserRegisteredEvent(eventData)
	case "UserLoggedIn":
		kc.service.HandleUserLoggedInEvent(eventData)
	default:
		log.Printf("unknown event type: %s\n", eventType)
	}
}
