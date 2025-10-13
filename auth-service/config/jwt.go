package config

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID string) (string, error) {
	jwtKey := []byte(GetEnv("JWT_SECRET"))

	claims := &jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
