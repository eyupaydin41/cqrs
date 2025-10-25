package model

import "time"

type User struct {
	ID        string    `gorm:"primaryKey"`
	Email     string    `gorm:"not null"`
	Status    string    `gorm:"type:varchar(20);default:'active';not null"` // active, deleted, suspended
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
