package http

import (
	"strings"

	"lucky-money/internal/application"
	"lucky-money/internal/domain/luckymoney"

	"github.com/gin-gonic/gin"
)

func UserRegister(svc *application.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID          string `json:"id"`
			AccountName string `json:"name"`
			Bank        string `json:"bank"`
			BankNo      string `json:"bankno"`
			FullName    string `json:"name"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "bad request"})
			return
		}

		if strings.TrimSpace(req.ID) == "" {
			c.JSON(400, gin.H{"error": "mã không hợp lệ"})
			return
		}

		err := svc.Register(req.ID, req.AccountName, req.Bank, req.BankNo, req.FullName)

		switch err {
		case nil:
			c.JSON(200, gin.H{"ok": true})
		case luckymoney.ErrIDNotIssued:
			c.JSON(404, gin.H{"error": "mã không tồn tại"})
		case luckymoney.ErrAlreadyRegistered:
			c.JSON(409, gin.H{"error": "mã này đã được sử dụng"})
		default:
			c.JSON(400, gin.H{"error": "mã này có vấn đề ! liên hệ admin"})
		}
	}
}

func UserSubmitAndDraw(svc *application.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID          string `json:"id"`
			AccountName string `json:"account"`
			Bank        string `json:"bank"`
			BankNo      string `json:"bankno"`
			FullName    string `json:"name"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "thất bại"})
			return
		}

		if strings.TrimSpace(req.ID) == "" {
			c.JSON(400, gin.H{"error": "mã không hợp lệ"})
			return
		}

		amount, err := svc.SubmitAndDraw(req.ID, req.AccountName, req.Bank, req.BankNo, req.FullName)

		switch err {
		case nil:
			c.JSON(200, gin.H{
				"status": "rút thành công",
				"amount": amount,
			})
		case luckymoney.ErrIDNotIssued:
			c.JSON(404, gin.H{"error": "mã này không tồn tại"})
		case luckymoney.ErrInfoMismatch:
			c.JSON(400, gin.H{"error": "thông tin chưa đúng"})
		case luckymoney.ErrAlreadyDrawn:
			c.JSON(409, gin.H{"error": "mã này đã được sử dụng"})
		case luckymoney.ErrPoolEmpty:
			c.JSON(400, gin.H{"error": "không còn lì xì"})
		default:
			c.JSON(400, gin.H{"error": "có lỗi rút lì xì ! liên hệ admin"})
		}
	}
}
