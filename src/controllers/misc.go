package controllers

import (
	"TradingSystem/src/common"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var OauthContent []byte
var systemsettings common.SystemSettings

func init() {
	systemsettings = common.GetEnvironmentSetting() //確保環境變數有值
	ApplyTgBotSetting(systemsettings.TgToken)
	settings := common.GetFirebaseSetting()

	fileContent, err := os.ReadFile(settings.OAuthKeyFullPath)
	if err != nil {
		log.Printf("Error reading JSON file: %v", err)
		return
	}
	OauthContent = fileContent
}

func FireAuthConfig(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", OauthContent)
}
