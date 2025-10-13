package model

import "time"

type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string
	CreatedAt time.Time
}
