package controllers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func DemoList(c *gin.Context) {
	if isAdministrator(c) {
		c.JSON(http.StatusOK, gin.H{"data": "OK"})
	} else {
		c.JSON(http.StatusOK, gin.H{"data": "Not OK"})
	}
}

func isAdministrator(c *gin.Context) bool {
	session := sessions.Default(c)
	isAdmin := session.Get("isadmin")
	if isAdmin != nil {
		return isAdmin.(bool)
	}
	return false
}
