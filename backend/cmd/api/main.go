package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	httpadapter "lucky-money/internal/adapter/http"
	"lucky-money/internal/application"
	"lucky-money/internal/store"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	r := gin.Default()
	r.Use(
		func(c *gin.Context) {
			log.Println("[REQ]", c.Request.Method, c.Request.URL.Path)
			c.Next()
		},
		httpadapter.DefaultCORS(),
	)

	adminUser := getEnv("ADMIN_USERNAME", "")
	adminPass := getEnv("ADMIN_PASSWORD", "")

	mem := store.NewMemoryStore()

	adminSvc := application.NewAdminService(
		mem, mem, mem, mem, mem,
		adminUser,
		adminPass,
	)
	userSvc := application.NewUserService(mem, mem, mem, mem)

	httpadapter.RegisterRoutes(r, adminSvc, userSvc)

	port := getEnv("PORT", "8080")
	log.Println("listening on :" + port)
	_ = r.Run(":" + port)
}
