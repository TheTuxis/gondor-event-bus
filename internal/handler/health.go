package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Health(c *gin.Context) {
	status := "healthy"
	checks := make(map[string]string)

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil {
		status = "unhealthy"
		checks["database"] = "error: " + err.Error()
	} else if err := sqlDB.Ping(); err != nil {
		status = "unhealthy"
		checks["database"] = "error: " + err.Error()
	} else {
		checks["database"] = "ok"
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
			checks["redis"] = "error: " + err.Error()
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "not configured"
	}

	code := http.StatusOK
	if status == "unhealthy" {
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, gin.H{
		"status":  status,
		"service": "gondor-event-bus",
		"checks":  checks,
	})
}
