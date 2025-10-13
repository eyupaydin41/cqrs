package model

import "time"

type User struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	Email     string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	CreatedAt time.Time
}
