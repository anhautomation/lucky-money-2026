package http

import (
	"lucky-money/internal/application"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, admin *application.AdminService, user *application.UserService) {
	// admin
	r.POST("/admin/login", AdminLogin(admin))
	adminGroup := r.Group("/admin")
	adminGroup.Use(RequireAdmin(admin))
	{
		adminGroup.POST("/pool", AdminSetPool(admin))
		adminGroup.GET("/pool", AdminGetPool(admin))
		adminGroup.POST("/ids", AdminIssueIDs(admin))
		adminGroup.GET("/ids", AdminListIDs(admin))
		adminGroup.DELETE("/ids/:id", AdminDeleteID(admin))
		adminGroup.GET("/claims", AdminGetClaims(admin))
		adminGroup.POST("/logout", AdminLogout(admin))
	}

	r.POST("/user/register", UserRegister(user))
	r.POST("/user/submit", UserSubmitAndDraw(user))
}
