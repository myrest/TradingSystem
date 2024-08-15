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
				// 在此处定义统一的错误响应
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.(error).Error()})
				c.Abort()
			}
		}()
		c.Next()
	}
}
