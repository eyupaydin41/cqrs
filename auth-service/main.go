package main

import (
	"os"

	"github.com/eyupaydin41/auth-service/api"
	"github.com/eyupaydin41/auth-service/config"
	"github.com/eyupaydin41/auth-service/event"
	"github.com/eyupaydin41/auth-service/repository"
	"github.com/eyupaydin41/auth-service/service"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	db := config.NewPostgresDB()
	userRepo := repository.NewUserRepository(db)

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "user-events"
	}

	producer := event.NewKafkaProducer(kafkaBroker, kafkaTopic)
	userService := service.NewUserService(userRepo, producer)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	r.POST("/register", api.RegisterHandler(userService))
	r.POST("/login", api.LoginHandler(userService))

	r.Run(":8088")
}
