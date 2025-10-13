package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env variables")
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
