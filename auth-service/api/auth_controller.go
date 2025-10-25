package api

import (
	"net/http"

	"github.com/eyupaydin41/auth-service/command"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterHandler - User kayıt endpoint'i (COMMAND)
func RegisterHandler(cmdHandler *command.CommandHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Command oluştur
		userID := uuid.New().String()
		cmd := command.RegisterUserCommand{
			UserID:   userID,
			Email:    req.Email,
			Password: req.Password,
		}

		// Command'ı işle
		err := cmdHandler.HandleRegisterUser(cmd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":      userID,
			"message": "User registered successfully. Please query from query-service.",
		})
	}
}

// ChangePasswordHandler - Şifre değiştirme endpoint'i (COMMAND)
func ChangePasswordHandler(cmdHandler *command.CommandHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		var req struct {
			OldPassword string `json:"old_password" binding:"required"`
			NewPassword string `json:"new_password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Command oluştur
		cmd := command.ChangePasswordCommand{
			UserID:      userID,
			OldPassword: req.OldPassword,
			NewPassword: req.NewPassword,
		}

		// Command'ı işle
		err := cmdHandler.HandleChangePassword(cmd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	}
}

// ChangeEmailHandler - Email değiştirme endpoint'i (COMMAND)
func ChangeEmailHandler(cmdHandler *command.CommandHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		var req struct {
			NewEmail string `json:"new_email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Command oluştur
		cmd := command.ChangeEmailCommand{
			UserID:   userID,
			NewEmail: req.NewEmail,
		}

		// Command'ı işle
		err := cmdHandler.HandleChangeEmail(cmd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email changed successfully"})
	}
}
