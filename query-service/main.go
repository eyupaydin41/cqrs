package main

import (
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
	userRepo := repository.NewUserRepository(db)
	loginHistoryRepo := repository.NewLoginHistoryRepository(db)
	userService := service.NewUserService(userRepo, loginHistoryRepo)

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}

	kafkaGroup := os.Getenv("KAFKA_GROUP")
	if kafkaGroup == "" {
		kafkaGroup = "query-group"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "user-events"
	}

	consumer := event.NewKafkaConsumer(kafkaBroker, kafkaGroup, kafkaTopic, userService)
	go consumer.Start()

	r := gin.Default()
	r.GET("/users", api.GetUsersHandler(userRepo))
	r.Run(":8089")
}
