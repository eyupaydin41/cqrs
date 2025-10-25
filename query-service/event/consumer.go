package event

import (
	"encoding/json"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/eyupaydin41/query-service/service"
)

type KafkaConsumer struct {
	consumer    *kafka.Consumer
	topic       string
	userService *service.UserService
	authService *service.AuthService
}

func NewKafkaConsumer(broker, group, topic string, service *service.UserService, authService *service.AuthService) *KafkaConsumer {
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
		consumer:    c,
		topic:       topic,
		userService: service,
		authService: authService,
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
	log.Println("✅ Received event:", string(eventData))
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
	case "user.created":
		if kc.authService != nil {
			if err := kc.authService.HandleUserCreatedEvent(eventData); err != nil {
				log.Printf("failed to handle user.created event: %v", err)
			}
		}
		kc.userService.HandleUserRegisteredEvent(eventData)

	case "user.password.changed":
		if kc.authService != nil {
			if err := kc.authService.HandlePasswordChangedEvent(eventData); err != nil {
				log.Printf("failed to handle user.password.changed event: %v", err)
			}
		}

	case "user.email.changed":
		if kc.authService != nil {
			if err := kc.authService.HandleEmailChangedEvent(eventData); err != nil {
				log.Printf("failed to handle user.email.changed event: %v", err)
			}
		}

	case "user.deactivated":
		if kc.authService != nil {
			if err := kc.authService.HandleUserDeactivatedEvent(eventData); err != nil {
				log.Printf("failed to handle user.deactivated event: %v", err)
			}
		}

	case "user.login.recorded":
		kc.userService.HandleUserLoggedInEvent(eventData)

	default:
		log.Printf("unknown event type: %s\n", eventType)
	}
}
