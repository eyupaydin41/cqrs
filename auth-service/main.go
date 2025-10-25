package main

import (
	"log"
	"os"

	"github.com/eyupaydin41/auth-service/api"
	"github.com/eyupaydin41/auth-service/command"
	"github.com/eyupaydin41/auth-service/config"
	"github.com/eyupaydin41/auth-service/event"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	producer := event.NewKafkaProducer(kafkaBroker, kafkaTopic)

	cmdHandler := command.NewCommandHandler(producer)

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
