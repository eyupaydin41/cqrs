package api

import (
	"net/http"

	"github.com/eyupaydin41/event-store/service"
	"github.com/gin-gonic/gin"
)

type SnapshotHandler struct {
	snapshotService *service.SnapshotService
}

func NewSnapshotHandler(snapshotService *service.SnapshotService) *SnapshotHandler {
	return &SnapshotHandler{
		snapshotService: snapshotService,
	}
}

// CreateSnapshot - Aggregate için snapshot oluşturur
func (h *SnapshotHandler) CreateSnapshot(c *gin.Context) {
	aggregateID := c.Param("aggregate_id")
	if aggregateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "aggregate_id is required"})
		return
	}

	err := h.snapshotService.CreateSnapshot(aggregateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create snapshot",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "snapshot created successfully",
		"aggregate_id": aggregateID,
	})
}

// GetLatestSnapshot - En son snapshot'ı getirir
func (h *SnapshotHandler) GetLatestSnapshot(c *gin.Context) {
	aggregateID := c.Param("aggregate_id")
	if aggregateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "aggregate_id is required"})
		return
	}

	aggregate, err := h.snapshotService.LoadAggregateWithSnapshot(aggregateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "snapshot not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"aggregate_id": aggregateID,
		"version":      aggregate.Version,
		"state":        aggregate,
	})
}

// GetAggregateState - Aggregate'in mevcut state'ini getirir (snapshot + event'ler)
func (h *SnapshotHandler) GetAggregateState(c *gin.Context) {
	aggregateID := c.Param("aggregate_id")
	if aggregateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "aggregate_id is required"})
		return
	}

	aggregate, err := h.snapshotService.LoadAggregateWithSnapshot(aggregateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "aggregate not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"aggregate_id": aggregateID,
		"version":      aggregate.Version,
		"email":        aggregate.Email,
		"status":       aggregate.Status,
		"created_at":   aggregate.CreatedAt,
		"updated_at":   aggregate.UpdatedAt,
		"event_count":  aggregate.EventCount,
	})
}
