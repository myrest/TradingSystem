package controllers

import (
	"TradingSystem/src/common"
	"log"
	"os"
)

func init() {
	ApplyTgBotSetting(systemsettings.TgToken)
	systemsettings = common.GetEnvironmentSetting()
	settings := common.GetFirebaseSetting()

	fileContent, err := os.ReadFile(settings.OAuthKeyFullPath)
	if err != nil {
		log.Printf("Error reading JSON file: %v", err)
		return
	}
	OauthContent = fileContent
}
