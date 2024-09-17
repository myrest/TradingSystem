package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AdminMiddleware 是一個檢查是否為管理者的中間件
func CustomerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		isLogined := session.Get("email")
		if isLogined != nil {
			c.Next() // 己登入，繼續處理請求
		} else {
			c.Redirect(http.StatusFound, "/")
			c.Abort() // 阻止請求繼續進行
		}
	}
}

// ErrorHandlingMiddleware 捕获并处理所有错误
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.(error).Error()})
				c.Abort()
			}
		}()

		c.Next() // 繼續執行其他中間件和處理器

		// 檢查上下文中的錯誤
		if len(c.Errors) > 0 {
			// 返回第一個錯誤
			c.JSON(http.StatusInternalServerError, gin.H{"error": c.Errors[0].Error()})
			c.Abort()
		}
	}
}
