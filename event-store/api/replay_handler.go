package api

import (
	"net/http"
	"time"

	"github.com/eyupaydin41/event-store/service"
	"github.com/gin-gonic/gin"
)

type ReplayHandler struct {
	replayService *service.ReplayService
}

func NewReplayHandler(replayService *service.ReplayService) *ReplayHandler {
	return &ReplayHandler{replayService: replayService}
}

// GetUserState - Kullanıcının şu anki durumunu event'lerden reconstruct eder
// GET /replay/user/:id/state
func (h *ReplayHandler) GetUserState(c *gin.Context) {
	userID := c.Param("id")

	aggregate, err := h.replayService.ReplayUserState(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"state":   aggregate,
		"message": "Current state reconstructed from events",
	})
}

// GetUserStateAt - Belirli bir zamandaki kullanıcı durumu (TIME TRAVEL!)
// GET /replay/user/:id/state-at?timestamp=2024-01-15T10:00:00Z
func (h *ReplayHandler) GetUserStateAt(c *gin.Context) {
	userID := c.Param("id")
	timestampStr := c.Query("timestamp")

	if timestampStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "timestamp parameter required (RFC3339 format)",
		})
		return
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid timestamp format, use RFC3339 (e.g., 2024-01-15T10:00:00Z)",
		})
		return
	}

	aggregate, err := h.replayService.ReplayUserStateAt(userID, timestamp)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"point_in_time": timestamp,
		"state":         aggregate,
		"message":       "State reconstructed at specified time",
	})
}

// GetUserHistory - Kullanıcının tüm değişiklik geçmişi
// GET /replay/user/:id/history
func (h *ReplayHandler) GetUserHistory(c *gin.Context) {
	userID := c.Param("id")

	history, err := h.replayService.GetUserHistory(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"history":       history,
		"total_changes": len(history),
		"message":       "Full history of state changes",
	})
}

// CompareStates - İki farklı zamandaki state'leri karşılaştır
// GET /replay/user/:id/compare?time1=2024-01-01T00:00:00Z&time2=2024-01-15T00:00:00Z
func (h *ReplayHandler) CompareStates(c *gin.Context) {
	userID := c.Param("id")
	time1Str := c.Query("time1")
	time2Str := c.Query("time2")

	if time1Str == "" || time2Str == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "both time1 and time2 parameters required (RFC3339 format)",
		})
		return
	}

	time1, err := time.Parse(time.RFC3339, time1Str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time1 format"})
		return
	}

	time2, err := time.Parse(time.RFC3339, time2Str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time2 format"})
		return
	}

	before, after, err := h.replayService.CompareStates(userID, time1, time2)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Değişiklikleri hesapla
	changes := map[string]interface{}{}
	if before.Email != after.Email {
		changes["email"] = map[string]string{
			"before": before.Email,
			"after":  after.Email,
		}
	}
	if before.Status != after.Status {
		changes["status"] = map[string]string{
			"before": before.Status,
			"after":  after.Status,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"time1":   time1,
		"time2":   time2,
		"before":  before,
		"after":   after,
		"changes": changes,
	})
}
