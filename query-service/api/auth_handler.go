package api

import (
	"net/http"
	"os"
	"time"

	"github.com/eyupaydin41/query-service/event"
	"github.com/eyupaydin41/query-service/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginHandler - Login endpoint'i (QUERY)
func LoginHandler(authService *service.AuthService, producer *event.KafkaProducer) gin.HandlerFunc {
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

		// 5. Login event'ini Kafka'ya publish et
		go publishLoginEvent(producer, authProj.ID, c.ClientIP(), c.Request.UserAgent())

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

// publishLoginEvent - Login event'ini Kafka'ya publish eder
func publishLoginEvent(producer *event.KafkaProducer, userID, ipAddress, userAgent string) {
	loginEvent := event.UserLoginRecordedEvent{
		EventType:   "user.login.recorded",
		AggregateID: userID,
		Timestamp:   time.Now(),
		Version:     1,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}

	producer.Publish("user.login.recorded", loginEvent)
}
