package controllers

import (
	"TradingSystem/src/common"
	"log"
	"os"
)

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
