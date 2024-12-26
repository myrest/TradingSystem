package controllers

import (
	"TradingSystem/src/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

var OauthContent []byte
var systemsettings common.SystemSettings

func FireAuthConfig(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", OauthContent)
}
