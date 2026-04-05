package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseUintParam(c *gin.Context, name string) (uint, error) {
	val, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid " + name + " parameter",
		})
		return 0, err
	}
	return uint(val), nil
}
