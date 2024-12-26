package controllers

import (
	"TradingSystem/src/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

//Todo:重新考慮是否只處理Production

func SystemSettings(c *gin.Context) {
	//需要限制只處理現行環境
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
	var settings common.SystemSettings
	err := c.Bind(&settings)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//需要限制只處理現行環境
	settings.Env = systemsettings.Env

	err = common.SaveDBSystemSettings(c, settings)
	if err != nil {
		c.JSON(common.HttpStatusSystemErrorCode, gin.H{"error": err.Error()})
	} else {
		common.ApplySystemSettings(settings)
		ApplyTgBotSetting(settings.TgToken)
		c.JSON(http.StatusOK, gin.H{"error": ""})
	}
}
