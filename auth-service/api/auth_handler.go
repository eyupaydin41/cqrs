package api

import (
	"net/http"

	"github.com/eyupaydin41/auth-service/service"
	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		id, err := userService.Register(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func LoginHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		token, err := userService.Login(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}
