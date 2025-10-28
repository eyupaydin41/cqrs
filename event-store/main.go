package main

import (
	"log"
	"os"

	"github.com/eyupaydin41/event-store/api"
	. "github.com/eyupaydin41/event-store/config"
	"github.com/eyupaydin41/event-store/consumer"
	grpcserver "github.com/eyupaydin41/event-store/grpc"
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

	eventConsumer := consumer.NewEventStoreConsumer(kafkaBroker, kafkaGroup, kafkaTopic, eventService, snapshotService)
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

	// HTTP Server port
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8090"
	}

	// gRPC Server port
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	// gRPC server'ı background'da başlat
	// HTTP'den fark: Ayrı bir goroutine'de çalışır
	go func() {
		log.Printf("🚀 gRPC server starting on port %s", grpcPort)
		if err := grpcserver.StartGRPCServer(":"+grpcPort, eventService, snapshotService); err != nil {
			log.Fatalf("failed to start gRPC server: %v", err)
		}
	}()

	// HTTP server'ı main goroutine'de başlat
	log.Printf("🌐 HTTP server starting on port %s", httpPort)
	if err := router.Run(":" + httpPort); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
