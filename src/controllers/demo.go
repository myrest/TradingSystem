package controllers

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DemoList(c *gin.Context) {
	d := c.Query("d")
	days, _ := strconv.Atoi(d)
	if days == 0 {
		days = 7
	} else if days > 30 {
		days = 30
	}

	systemSymboList, err := services.GetDemoCurrencyList(c, days, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "demosymbolist.html", gin.H{
		"data":              systemSymboList,
		"days":              days,
		"StaticFileVersion": systemsettings.StartTimestemp,
	})
}

func DemoHistory(c *gin.Context) {
	d := c.Query("d")
	symbol := c.Query("symbol")

	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Symbol Data."})
		return
	}

	days, _ := strconv.Atoi(d)
	if days == 0 {
		days = 7
	} else if days > 30 {
		days = 30
	}

	var rtn []Log_PlaceBetHistoryUI
	list, err := services.GetDemoHistory(c, days, symbol, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := 0; i < len(list); i++ {
		positionside := "多"
		side := "開"
		if list[i].PositionSideType == models.ShortPositionSideType {
			positionside = "空"
		}
		if (list[i].PositionSideType == models.ShortPositionSideType && list[i].Side == models.BuySideType) ||
			(list[i].PositionSideType == models.LongPositionSideType && list[i].Side == models.SellSideType) {
			side = "平"
		}
		rtn = append(rtn, Log_PlaceBetHistoryUI{
			Log_TvSiginalData: list[i],
			Position:          side + positionside,
		})
	}

	c.HTML(http.StatusOK, "demohistory.html", gin.H{
		"data":              rtn,
		"symbol":            symbol,
		"days":              days,
		"StaticFileVersion": systemsettings.StartTimestemp,
	})
}
