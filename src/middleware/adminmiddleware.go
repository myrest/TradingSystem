package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AdminMiddleware 是一個檢查是否為管理者的中間件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		isAdmin := session.Get("isadmin")
		if isAdmin != nil && isAdmin.(bool) {
			c.Set("IsAdmin", true)
			c.Next() // 是管理者，繼續處理請求
		} else {
			handleError(c, errors.New("access forbidden"), http.StatusForbidden)
			c.Abort() // 阻止請求繼續進行
		}
	}
}
