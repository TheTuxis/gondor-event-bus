package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TheTuxis/gondor-event-bus/internal/model"
	"github.com/TheTuxis/gondor-event-bus/internal/service"
)

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

func (h *EventHandler) Publish(c *gin.Context) {
	var input model.PublishRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid request body",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	eventLog, err := h.eventService.Publish(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to publish event",
		})
		return
	}

	c.JSON(http.StatusCreated, eventLog)
}

func (h *EventHandler) ListLogs(c *gin.Context) {
	var params model.ListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid query parameters",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	result, err := h.eventService.ListLogs(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list event logs",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
