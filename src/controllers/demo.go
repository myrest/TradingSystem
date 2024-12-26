package controllers

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func DemoList(c *gin.Context) {
	session := sessions.Default(c)

	customerid := systemsettings.DemoCustomerID

	reportStartDate := time.Now().UTC()

	startDate, endDate := common.GetMonthStartEndDate(reportStartDate)

	//將日期區間寫入DB
	common.SetReportStartEndDate(session, startDate, endDate)

	montylyreport, err := services.GetCustomerMonthlyReportCurrencyList(c, customerid, startDate, endDate)
	if err != nil {
		c.Error(err) // 將錯誤添加到上下文中
		return
	}

	//找出每月1號的清單
	monthday := common.GetMonthlyDay1(3)

	firstDayByMonth := []string{}
	for _, day := range monthday {
		firstDayByMonth = append(firstDayByMonth, common.FormatDate(day))
	}

	// 排序切片
	sort.Slice(montylyreport, func(i, j int) bool {
		return montylyreport[i].Symbol > montylyreport[j].Symbol // 降冪排序
	})

	c.HTML(http.StatusOK, "monthlyreport.html", gin.H{
		"data":    montylyreport,
		"mondays": firstDayByMonth,
		"days":    common.FormatDate(startDate),
		//"cid":               customerid,
		"month":             common.GetMonthsInRange(startDate)[0],
		"StaticFileVersion": systemsettings.StartTimestemp,
	})
}

func DemoListII(c *gin.Context) {
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
