package http

import (
	"net/http"

	"lucky-money/internal/application"

	"github.com/gin-gonic/gin"
)

func DefaultCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-Admin-Session")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func RequireAdmin(admin *application.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Admin-Session")
		if !admin.IsSessionValid(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "chưa đăng nhập"})
			return
		}
		c.Next()
	}
}
