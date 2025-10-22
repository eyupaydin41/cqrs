package main

import (
	"log"
	"os"

	"github.com/eyupaydin41/event-store/api"
	. "github.com/eyupaydin41/event-store/config"
	"github.com/eyupaydin41/event-store/consumer"
	"github.com/eyupaydin41/event-store/repository"
	"github.com/eyupaydin41/event-store/service"
	"github.com/gin-gonic/gin"
)

func main() {
	LoadEnv()

	conn := InitClickHouse()
	defer conn.Close()

	repo := repository.NewEventRepository(conn)
	eventService := service.NewEventService(repo)

	kafkaBroker := GetEnv("KAFKA_BROKER")
	kafkaTopic := GetEnv("KAFKA_TOPIC")
	kafkaGroup := GetEnv("KAFKA_GROUP")

	eventConsumer := consumer.NewEventStoreConsumer(kafkaBroker, kafkaGroup, kafkaTopic, eventService)
	go eventConsumer.Start()

	handler := api.NewEventHandler(eventService)

	router := gin.Default()

	router.GET("/health", handler.HealthCheck)

	router.GET("/events", handler.GetEvents)
	router.GET("/events/aggregate/:id", handler.GetEventsByAggregate)
	router.GET("/events/replay", handler.ReplayEvents)
	router.GET("/events/count", handler.GetEventCount)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Printf("event store service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
