package event

import (
	"encoding/json"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

func NewKafkaProducer(broker, topic string) *KafkaProducer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}

	return &KafkaProducer{
		producer: p,
		topic:    topic,
	}
}

func (kp *KafkaProducer) Publish(eventType string, payload interface{}) {
	data := map[string]interface{}{
		"type": eventType,
		"data": payload,
	}

	value, _ := json.Marshal(data)
	err := kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kp.topic, Partition: kafka.PartitionAny},
		Value:          value,
	}, nil)

	if err != nil {
		log.Printf("failed to send message: %v", err)
	}
}
