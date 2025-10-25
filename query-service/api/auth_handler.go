package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/eyupaydin41/query-service/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginHandler - Login endpoint'i (QUERY)
// Authentication query service'de yapılır
func LoginHandler(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// 1. Auth projection'dan user'ı bul
		authProj, err := authService.FindByEmail(req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// 2. User aktif mi kontrol et
		if authProj.Status != "active" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user is not active"})
			return
		}

		// 3. Password'ü doğrula
		err = bcrypt.CompareHashAndPassword([]byte(authProj.PasswordHash), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// 4. JWT token oluştur
		token, err := generateJWT(authProj.ID, authProj.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		// 5. Login event'ini command service'e gönder (async)
		go recordLoginEvent(authProj.ID, c.ClientIP(), c.Request.UserAgent())

		c.JSON(http.StatusOK, gin.H{
			"token":   token,
			"user_id": authProj.ID,
			"email":   authProj.Email,
		})
	}
}

// generateJWT - JWT token oluşturur
func generateJWT(userID, email string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// recordLoginEvent - Command service'e login event'i gönderir
func recordLoginEvent(userID, ipAddress, userAgent string) {
	commandServiceURL := os.Getenv("COMMAND_SERVICE_URL")
	if commandServiceURL == "" {
		commandServiceURL = "http://localhost:8088"
	}

	reqBody := map[string]string{
		"user_id": userID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Failed to marshal login event: %v\n", err)
		return
	}

	url := fmt.Sprintf("%s/login/record", commandServiceURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to create login record request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", ipAddress)
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to record login event: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Login record returned status: %d\n", resp.StatusCode)
	}
}
