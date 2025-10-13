package api

import (
	"net/http"

	"github.com/eyupaydin41/query-service/repository"
	"github.com/gin-gonic/gin"
)

func GetUsersHandler(repo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repo.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
			return
		}
		c.JSON(http.StatusOK, users)
	}
}
