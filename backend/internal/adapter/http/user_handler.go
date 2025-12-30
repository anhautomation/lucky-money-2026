package http

import (
	"lucky-money/internal/application"
	"lucky-money/internal/domain/luckymoney"

	"github.com/gin-gonic/gin"
)

func UserRegister(svc *application.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID          string `json:"id"`
			AccountName string `json:"account_name"`
			Phone       string `json:"phone"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "bad request"})
			return
		}
		err := svc.Register(req.ID, req.AccountName, req.Phone)
		if err == luckymoney.ErrIDNotIssued {
			c.JSON(404, gin.H{"error": "mã không tồn tại"})
			return
		}
		if err == luckymoney.ErrAlreadyRegistered {
			c.JSON(409, gin.H{"error": "mã này đã được sử dụng"})
			return
		}
		if err != nil {
			c.JSON(400, gin.H{"error": "mã này có vấn đề ! liên hệ admin"})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	}
}

func UserSubmitAndDraw(svc *application.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID          string `json:"id"`
			AccountName string `json:"account_name"`
			Phone       string `json:"phone"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "thất bại"})
			return
		}

		amount, err := svc.SubmitAndDraw(req.ID, req.AccountName, req.Phone)

		switch err {
		case nil:
			c.JSON(200, gin.H{"status": "rút thành công", "amount": amount})
		case luckymoney.ErrIDNotIssued:
			c.JSON(404, gin.H{"error": "mã này không tồn tại"})
		case luckymoney.ErrInfoMismatch:
			c.JSON(400, gin.H{"error": "thông tin chưa đúng"})
		case luckymoney.ErrAlreadyDrawn:
			c.JSON(200, gin.H{"status": "mã này đã được sử dụng", "amount": amount})
		case luckymoney.ErrPoolEmpty:
			c.JSON(400, gin.H{"error": "không còn lì xì"})
		default:
			c.JSON(400, gin.H{"error": "có lỗi rút lì xì ! liên hệ admin"})
		}
	}
}
