package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var OauthContent []byte

func FireAuthConfig(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", OauthContent)
}
