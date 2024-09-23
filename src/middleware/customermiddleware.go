package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// CustomerMiddleware 是一個檢查是否己登入的中間件
func CustomerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		isLogined := session.Get("email")
		isAdmin := session.Get("isadmin")
		c.Set("IsAdmin", isAdmin != nil && isAdmin.(bool))
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
			handleError(c, c.Errors[0].Err)
			c.Abort()
		}
	}
}

func handleError(c *gin.Context, err error, httpstatuscode ...int) {
	statusCode := http.StatusInternalServerError
	if len(httpstatuscode) == 0 {
		statusCode = httpstatuscode[0]
	}
	// 檢查請求的Accept標頭
	acceptHeader := c.GetHeader("Accept")
	if acceptHeader == "text/html" {
		// 返回HTML格式
		c.HTML(statusCode, "error.html", gin.H{"error": err.Error()})
	} else {
		// 返回JSON格式
		c.JSON(statusCode, gin.H{"error": err.Error()})
	}
}
