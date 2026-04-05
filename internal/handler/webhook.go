package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TheTuxis/gondor-event-bus/internal/model"
	"github.com/TheTuxis/gondor-event-bus/internal/service"
)

type WebhookHandler struct {
	webhookService *service.WebhookService
}

func NewWebhookHandler(webhookService *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService}
}

func (h *WebhookHandler) List(c *gin.Context) {
	var params model.ListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid query parameters",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	result, err := h.webhookService.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to list webhooks",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *WebhookHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	webhook, err := h.webhookService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "webhook not found",
		})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var input model.WebhookCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid request body",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	webhook, err := h.webhookService.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to create webhook",
		})
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

func (h *WebhookHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	var input model.WebhookUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid request body",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	webhook, err := h.webhookService.Update(id, input)
	if err != nil {
		if err == service.ErrWebhookNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "webhook not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to update webhook",
		})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	if err := h.webhookService.Delete(id); err != nil {
		if err == service.ErrWebhookNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "webhook not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to delete webhook",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook deleted successfully"})
}

func (h *WebhookHandler) Test(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	webhook, err := h.webhookService.Test(id)
	if err != nil {
		if err == service.ErrWebhookNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "webhook not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to test webhook",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "test event sent",
		"webhook": webhook,
	})
}
