package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/eyupaydin41/event-store/model"
	"github.com/eyupaydin41/event-store/service"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	service *service.EventService
}

func NewEventHandler(service *service.EventService) *EventHandler {
	return &EventHandler{service: service}
}

func (h *EventHandler) GetEvents(c *gin.Context) {
	var filter model.EventFilter

	if eventType := c.Query("event_type"); eventType != "" {
		filter.EventType = eventType
	}

	if aggregateID := c.Query("aggregate_id"); aggregateID != "" {
		filter.AggregateID = aggregateID
	}

	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = t
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	events, err := h.service.GetEvents(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

func (h *EventHandler) GetEventsByAggregate(c *gin.Context) {
	aggregateID := c.Param("id")

	fromVersion := 0
	if v := c.Query("from_version"); v != "" {
		if ver, err := strconv.Atoi(v); err == nil {
			fromVersion = ver
		}
	}

	events, err := h.service.GetEventsByAggregateID(aggregateID, fromVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"aggregate_id": aggregateID,
		"events":       events,
		"count":        len(events),
	})
}

func (h *EventHandler) ReplayEvents(c *gin.Context) {
	sinceStr := c.Query("since")
	if sinceStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "since parameter is required (RFC3339 format)"})
		return
	}

	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time format, use RFC3339"})
		return
	}

	events, err := h.service.GetEventsSince(since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"since":  since,
		"events": events,
		"count":  len(events),
	})
}

func (h *EventHandler) GetEventCount(c *gin.Context) {
	count, err := h.service.CountEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_events": count,
	})
}

func (h *EventHandler) HealthCheck(c *gin.Context) {
	count, err := h.service.CountEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "healthy",
		"total_events": count,
	})
}
