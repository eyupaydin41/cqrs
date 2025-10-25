package model

import "time"

// AuthProjection - Authentication için özel projection
// Bu projection sadece login işlemi için gerekli bilgileri tutar
type AuthProjection struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"password_hash" db:"password_hash"`
	Status       string    `json:"status" db:"status"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
