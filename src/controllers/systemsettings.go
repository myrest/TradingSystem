package controllers

import (
	"TradingSystem/src/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SystemSettings(c *gin.Context) {
	sys, err := common.GetDBSystemSettings(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sys.SectestWord = systemsettings.SectestWord

	c.HTML(http.StatusOK, "systemsettings.html", gin.H{
		"data":              sys,
		"StaticFileVersion": systemsettings.StartTimestemp,
	})
}

func SaveSystemSettings(c *gin.Context) {
	var newInputSetting common.SystemSettings
	err := c.Bind(&newInputSetting)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//需要限制只處理當前環境，也可以從該設定檔知道目前執行的環境
	newInputSetting.Env = systemsettings.Env

	err = common.SaveDBSystemSettings(c, newInputSetting)
	if err != nil {
		c.JSON(common.HttpStatusSystemErrorCode, gin.H{"error": err.Error()})
	} else {
		common.ApplySystemSettings(newInputSetting)
		ApplyTgBotSetting(newInputSetting.TgToken)
		c.JSON(http.StatusOK, gin.H{"error": ""})
	}
}
