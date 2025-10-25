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

	// Repositories
	eventRepo := repository.NewEventRepository(conn)
	snapshotRepo := repository.NewSnapshotRepository(conn)

	// Snapshot tablosunu oluştur
	if err := snapshotRepo.CreateTable(); err != nil {
		log.Printf("Warning: Failed to create snapshot table: %v", err)
	}

	// Services
	eventService := service.NewEventService(eventRepo)
	replayService := service.NewReplayService(eventRepo)
	snapshotService := service.NewSnapshotService(snapshotRepo, eventRepo)

	kafkaBroker := GetEnv("KAFKA_BROKER")
	kafkaTopic := GetEnv("KAFKA_TOPIC")
	kafkaGroup := GetEnv("KAFKA_GROUP")

	eventConsumer := consumer.NewEventStoreConsumer(kafkaBroker, kafkaGroup, kafkaTopic, eventService)
	go eventConsumer.Start()

	// Handlers
	handler := api.NewEventHandler(eventService)
	replayHandler := api.NewReplayHandler(replayService)
	snapshotHandler := api.NewSnapshotHandler(snapshotService)

	router := gin.Default()

	router.GET("/health", handler.HealthCheck)

	// Event endpoints (sadece query için, artık HTTP ile write yok)
	router.GET("/events", handler.GetEvents)
	router.GET("/events/aggregate/:id", handler.GetEventsByAggregate)
	router.GET("/events/replay", handler.ReplayEvents)
	router.GET("/events/count", handler.GetEventCount)

	// Snapshot endpoints
	router.POST("/snapshots/:aggregate_id", snapshotHandler.CreateSnapshot)
	router.GET("/snapshots/:aggregate_id", snapshotHandler.GetLatestSnapshot)
	router.GET("/snapshots/:aggregate_id/state", snapshotHandler.GetAggregateState)

	// Time Travel endpoints
	router.GET("/replay/user/:id/state", replayHandler.GetUserState)
	router.GET("/replay/user/:id/state-at", replayHandler.GetUserStateAt)
	router.GET("/replay/user/:id/history", replayHandler.GetUserHistory)
	router.GET("/replay/user/:id/compare", replayHandler.CompareStates)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Printf("event store service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
