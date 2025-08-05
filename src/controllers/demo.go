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

	d := c.Query("d")
	reportStartDate := time.Now().UTC()

	if d != "" {
		reportStartDate = common.ParseTime(d)
	}

	startDate, endDate := common.GetMonthStartEndDate(reportStartDate)

	//將日期區間寫入DB
	common.SetReportStartEndDate(session, startDate, endDate)

	montylyreport, err := services.GetCustomerMonthlyReportCurrencyList(c, customerid, startDate, endDate, true)
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

func DemoHistory(c *gin.Context) {
	symbol := c.Query("symbol")
	customerid := systemsettings.DemoCustomerID
	session := sessions.Default(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	symbol = common.FormatSymbol(symbol)

	var rtn []Log_PlaceBetHistoryUI

	if customerid == "" {
		cid := session.Get("id")
		if cid != nil {
			customerid = cid.(string)
		}
	}
	if customerid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Customer Data."})
		return
	}

	//如果Session有值，就以Session的為主，若沒有就取三個月內的
	sdt, edt := common.GetReportStartEndDate(session)
	if sdt.Equal(edt) {
		sdt, edt = common.GetMonthStartEndDate(time.Now().UTC())
		sdt = sdt.AddDate(0, -3, 0) //一次三個月內的資料
	}

	common.SetReportStartEndDate(session, sdt, edt)

	list, totalPages, err := services.GetPlaceOrderHistory(c, symbol, customerid, sdt, edt, page, pageSize)
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

	c.HTML(http.StatusOK, "placeorderhistory.html", gin.H{
		"data":              rtn,
		"page":              page,
		"pageSize":          pageSize,
		"totalPages":        totalPages,
		"symbol":            symbol,
		"cid":               c.Query("cid"),
		"StaticFileVersion": systemsettings.StartTimestemp,
	})
}
