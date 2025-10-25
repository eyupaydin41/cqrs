package main

import (
	"log"
	"os"

	"github.com/eyupaydin41/auth-service/api"
	"github.com/eyupaydin41/auth-service/command"
	"github.com/eyupaydin41/auth-service/config"
	"github.com/eyupaydin41/auth-service/event"
	grpcclient "github.com/eyupaydin41/auth-service/grpc"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	// Kafka Producer (event publishing için)
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	producer := event.NewKafkaProducer(kafkaBroker, kafkaTopic)

	// gRPC Client (event-store'dan aggregate load etmek için)
	// HTTP'de: http.DefaultClient kullanırdık
	// gRPC'de: Custom client oluşturuyoruz
	eventStoreGRPC := os.Getenv("EVENT_STORE_GRPC")
	if eventStoreGRPC == "" {
		eventStoreGRPC = "event-store:9090" // Docker compose içinde
	}

	log.Printf("🔌 Connecting to Event-Store gRPC at %s", eventStoreGRPC)
	eventStoreClient, err := grpcclient.NewEventStoreClient(eventStoreGRPC)
	if err != nil {
		log.Fatalf("Failed to connect to event-store gRPC: %v", err)
	}
	defer eventStoreClient.Close()

	// Command Handler
	cmdHandler := command.NewCommandHandler(producer, eventStoreClient)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	r.POST("/register", api.RegisterHandler(cmdHandler))
	r.PUT("/users/:id/password", api.ChangePasswordHandler(cmdHandler))
	r.PUT("/users/:id/email", api.ChangeEmailHandler(cmdHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	log.Printf("Auth service (COMMAND) starting on port %s", port)
	r.Run(":" + port)
}
