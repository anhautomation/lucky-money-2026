package http

import (
	"net/http"

	"lucky-money/internal/application"

	"github.com/gin-gonic/gin"
)

func DefaultCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-Admin-Session, Authorization")
		c.Header("Access-Control-Expose-Headers", "X-Admin-Session")
		c.Header("Access-Control-Max-Age", "86400")
		c.Header("Access-Control-Allow-Credentials", "true")

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
