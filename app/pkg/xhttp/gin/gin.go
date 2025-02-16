package xgin

import (
	"context"
	"net/http"

	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/gin-gonic/gin"
)

func HandleReadiness(ctx context.Context) func(*gin.Context) {
	return func(c *gin.Context) {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusServiceUnavailable, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: -2, Message: "shutting down"}))
		default:
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		}
	}
}

func HandleNotFound() func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(404, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: 0, Message: "handler not found"}))
	}
}
