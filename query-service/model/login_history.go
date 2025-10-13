package model

import "time"

type LoginHistory struct {
	ID        string `gorm:"primaryKey"`
	UserID    string `gorm:"index"`
	Email     string
	LoginAt   time.Time `gorm:"index"`
	CreatedAt time.Time
}
