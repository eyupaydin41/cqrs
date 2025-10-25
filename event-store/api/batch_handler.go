package api

import (
	"net/http"

	"github.com/eyupaydin41/event-store/model"
	"github.com/eyupaydin41/event-store/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BatchEventHandler struct {
	eventService *service.EventService
}

func NewBatchEventHandler(eventService *service.EventService) *BatchEventHandler {
	return &BatchEventHandler{
		eventService: eventService,
	}
}

// SaveBatchEvents - Toplu event kaydetme endpoint'i
// Command service buraya event'leri batch olarak gönderir
func (h *BatchEventHandler) SaveBatchEvents(c *gin.Context) {
	var req struct {
		Events []struct {
			EventType   string                 `json:"event_type" binding:"required"`
			AggregateID string                 `json:"aggregate_id" binding:"required"`
			Payload     string                 `json:"payload" binding:"required"`
			Timestamp   string                 `json:"timestamp"`
			Version     uint32                 `json:"version"`
			Metadata    map[string]interface{} `json:"metadata"`
		} `json:"events" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Event'leri dönüştür
	events := make([]*model.Event, len(req.Events))
	for i, eventReq := range req.Events {
		events[i] = &model.Event{
			ID:          uuid.New().String(),
			EventType:   eventReq.EventType,
			AggregateID: eventReq.AggregateID,
			Payload:     eventReq.Payload,
			Version:     eventReq.Version,
		}

		// Timestamp parse et
		if eventReq.Timestamp != "" {
			// Timestamp string'i parse et ve event.Timestamp'e ata
			// Şimdilik boş bırak, service katmanında set edilecek
		}
	}

	// Event'leri kaydet
	for _, event := range events {
		if err := h.eventService.SaveEvent(event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to save event",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "events saved successfully",
		"count":   len(events),
	})
}
