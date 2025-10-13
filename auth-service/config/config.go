package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB() *gorm.DB {
	LoadEnv()

	host := GetEnv("DB_HOST")
	user := GetEnv("DB_USER")
	password := GetEnv("DB_PASSWORD")
	dbname := GetEnv("DB_NAME")
	port := GetEnv("DB_PORT")
	sslmode := GetEnv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	fmt.Println("Connected to PostgreSQL!")
	return db
}
