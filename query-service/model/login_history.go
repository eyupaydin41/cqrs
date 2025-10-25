package model

import "time"

type LoginHistory struct {
	ID        string `gorm:"primaryKey"`
	UserID    string `gorm:"index"`
	Email     string
	IPAddress string
	UserAgent string
	LoginAt   time.Time `gorm:"index"`
	CreatedAt time.Time
}
