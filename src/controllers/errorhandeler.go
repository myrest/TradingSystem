package controllers

import (
	"github.com/gin-gonic/gin"
)

func handleCustomError(c *gin.Context, statusCode int, errormessage string, flags ...bool) {
	isReturnHtml := false
	if len(flags) > 1 {
		isReturnHtml = flags[0]
	}

	if isReturnHtml {
		// 返回 HTML 页面
		c.HTML(statusCode, "error.html", gin.H{
			"error":      errormessage,
			"statuscode": statusCode,
		})
	} else {
		// 返回 JSON 错误信息
		if err := recover(); err != nil {
			// 在此处定义统一的错误响应
			c.JSON(statusCode, gin.H{
				"error":      err.(error).Error(),
				"statuscode": statusCode,
			})
		}
	}
	c.Abort()
}
