package main

import (
	"log"
	"os"

	"github.com/eyupaydin41/query-service/api"
	"github.com/eyupaydin41/query-service/config"
	"github.com/eyupaydin41/query-service/event"
	"github.com/eyupaydin41/query-service/repository"
	"github.com/eyupaydin41/query-service/service"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	db := config.NewPostgresDB()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	loginHistoryRepo := repository.NewLoginHistoryRepository(db)
	authRepo := repository.NewAuthProjectionRepository(db)

	// Auth projection tablosunu oluştur
	if err := authRepo.CreateTable(); err != nil {
		log.Fatalf("Failed to create auth projection table: %v", err)
	}

	// Services
	userService := service.NewUserService(userRepo, loginHistoryRepo)
	authService := service.NewAuthService(authRepo)

	// Kafka consumer
	kafkaBroker := os.Getenv("KAFKA_BROKER")

	kafkaGroup := os.Getenv("KAFKA_GROUP")

	kafkaTopic := os.Getenv("KAFKA_TOPIC")

	consumer := event.NewKafkaConsumer(kafkaBroker, kafkaGroup, kafkaTopic, userService, authService)
	go consumer.Start()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	// QUERY endpoints
	r.GET("/users", api.GetUsersHandler(userRepo))
	r.POST("/login", api.LoginHandler(authService))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}

	log.Printf("Query service starting on port %s", port)
	r.Run(":" + port)
}
