package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	envPath := "../.env"
	if err := godotenv.Load(envPath); err != nil {
		log.Println("Warning: .env file not found, using system env variables")
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
