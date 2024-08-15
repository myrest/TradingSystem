package controllers

import (
	"github.com/gin-gonic/gin"
)

func handleCustomError(c *gin.Context, err error, customizeMsg ...string) {
	if err != nil || len(customizeMsg) > 0 {
		if err != nil {
			c.HTML(490, "error.html", gin.H{
				"error":      err.Error(),
				"statuscode": 490,
			})
		} else {
			c.HTML(490, "error.html", gin.H{
				"error":      customizeMsg[0],
				"statuscode": 490,
			})
		}
		c.Abort()
	}
}

func handleCustomErrorJson(c *gin.Context, err error, customizeMsg ...string) {
	if err != nil || len(customizeMsg) > 0 {
		if err != nil {
			c.JSON(490, gin.H{
				"error":      err.Error(),
				"statuscode": 490,
			})
		} else {
			c.JSON(490, gin.H{
				"error":      customizeMsg[0],
				"statuscode": 490,
			})
		}
		c.Abort()
	}
}
