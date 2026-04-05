package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TheTuxis/gondor-event-bus/internal/model"
	"github.com/TheTuxis/gondor-event-bus/internal/service"
)

type DeadLetterHandler struct {
	dlqService *service.DeadLetterService
}

func NewDeadLetterHandler(dlqService *service.DeadLetterService) *DeadLetterHandler {
	return &DeadLetterHandler{dlqService: dlqService}
}

func (h *DeadLetterHandler) List(c *gin.Context) {
	var params model.ListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid query parameters",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	result, err := h.dlqService.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list dead letter messages",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *DeadLetterHandler) Retry(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	msg, err := h.dlqService.Retry(id)
	if err != nil {
		if err == service.ErrDLQMessageNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "dead letter message not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, msg)
}

func (h *DeadLetterHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	if err := h.dlqService.Delete(id); err != nil {
		if err == service.ErrDLQMessageNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "dead letter message not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to delete dead letter message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dead letter message deleted successfully"})
}
