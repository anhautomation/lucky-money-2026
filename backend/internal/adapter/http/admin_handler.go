package http

import (
	"net/http"

	"lucky-money/internal/application"
	"lucky-money/internal/domain/luckymoney"

	"github.com/gin-gonic/gin"
)

func requireAdminSession(c *gin.Context, svc *application.AdminService) bool {
	token := c.GetHeader("X-Admin-Session")
	if token == "" || !svc.IsSessionValid(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "chưa đăng nhập"})
		return false
	}
	return true
}

func AdminLogin(svc *application.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			User string `json:"user"`
			Pass string `json:"pass"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "thất bại"})
			return
		}
		token, err := svc.Login(req.User, req.Pass)
		if err != nil {
			c.JSON(401, gin.H{"error": "chưa đăng nhập"})
			return
		}
		c.JSON(200, gin.H{"session": token})
	}
}

func AdminLogout(svc *application.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Admin-Session")
		if err := svc.Logout(token); err != nil {
			c.JSON(500, gin.H{"error": "logout thất bại"})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	}
}

func AdminSetPool(svc *application.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireAdminSession(c, svc) {
			return
		}

		var req struct {
			Items []luckymoney.PoolItem `json:"items"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "thất bại"})
			return
		}

		if err := svc.SetPool(req.Items); err != nil {
			c.JSON(500, gin.H{"error": "set pool thất bại"})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	}
}

func AdminGetPool(svc *application.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireAdminSession(c, svc) {
			return
		}
		items := svc.GetPool()
		c.JSON(200, gin.H{"items": items})
	}
}

func AdminIssueIDs(svc *application.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireAdminSession(c, svc) {
			return
		}

		var req struct {
			IDs []string `json:"ids"`
		}
		if err := c.BindJSON(&req); err != nil || len(req.IDs) == 0 {
			c.JSON(400, gin.H{"error": "thất bại"})
			return
		}

		if err := svc.IssueIDs(req.IDs); err != nil {
			c.JSON(500, gin.H{"error": "issue ids thất bại"})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	}
}

func AdminGetClaims(svc *application.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireAdminSession(c, svc) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": svc.GetClaims()})
	}
}

func AdminListIDs(svc *application.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireAdminSession(c, svc) {
			return
		}
		ids := svc.ListIssuedIDs()
		c.JSON(200, gin.H{"items": ids})
	}
}

func AdminDeleteID(svc *application.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireAdminSession(c, svc) {
			return
		}

		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "không có mã này"})
			return
		}

		err := svc.DeleteIssuedID(id)
		if err == luckymoney.ErrIDAlreadyDrawn {
			c.JSON(409, gin.H{"error": "mã này đã được kích hoạt"})
			return
		}
		if err != nil {
			c.JSON(400, gin.H{"error": "xoá thất bại"})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	}
}
