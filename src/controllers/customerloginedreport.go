package controllers

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CustomerReportList(c *gin.Context) {
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
		"data": systemSymboList,
		"days": days,
	})
}

func CustomerWeeklyReportList(c *gin.Context) {
	session := sessions.Default(c)
	cid := c.DefaultQuery("cid", "")

	customerid := session.Get("id").(string)

	//只有管理員可以看到其它人的記錄。
	if cid != "" && session.Get("isadmin") != nil && session.Get("isadmin").(bool) {
		customerid = cid
	}

	if customerid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Customer Data."})
		return
	}

	d := c.Query("d")
	reportStartDate := time.Now().UTC()

	if d != "" {
		reportStartDate = common.ParseTime(d)
	}

	startDate, endDate := common.GetWeeksStartEndDateByDate(reportStartDate)

	weeklyreport, err := services.GetCustomerReportCurrencyList(c, customerid, startDate, endDate)
	if err != nil {
		return
	}

	//找出星期一清單
	mondays, err := common.GetPreviousMondays(time.Now().UTC(), 12)
	if err != nil {
		return
	}

	if session.Get("isadmin") == nil || !session.Get("isadmin").(bool) { //不是管理員cid要清掉不給看
		customerid = ""
	}
	c.HTML(http.StatusOK, "weeklyreport.html", gin.H{
		"data":    weeklyreport,
		"mondays": mondays,
		"days":    common.FormatDate(common.ParseTime(startDate)),
		"cid":     customerid,
		"week":    common.GetWeeksByDate(common.ParseTime(startDate)),
	})
}

func CustomerWeeklyReportSummaryList(c *gin.Context) {
	session := sessions.Default(c)
	cid := c.DefaultQuery("cid", "")

	customerid := session.Get("id").(string)

	//只有管理員可以看到其它人的記錄。
	if cid != "" && session.Get("isadmin") != nil && session.Get("isadmin").(bool) {
		customerid = cid
	}

	if customerid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Customer Data."})
		return
	}

	d := c.Query("d")
	enddate := time.Now().UTC()

	if d != "" {
		enddate = common.ParseTime(d)
	}

	reportStartDate := common.FormatDate(enddate.AddDate(0, -2, 0))
	reportEndDate := common.FormatDate(enddate)

	weeklyreport, err := services.GetCustomerReportCurrencySummaryList(c, customerid, reportStartDate, reportEndDate)
	if err != nil {
		return
	}

	if session.Get("isadmin") == nil || !session.Get("isadmin").(bool) { //不是管理員cid要清掉不給看
		customerid = ""
	}
	var rtn []models.CustomerWeeklyReportSummaryUI
	for _, w := range weeklyreport {
		stde, enddt, _ := common.WeekToDateRange(w.YearWeek)
		w.Profit = common.Decimal(w.Profit, 2)
		rtn = append(rtn, models.CustomerWeeklyReportSummaryUI{
			CustomerWeeklyReportSummary: w,
			StartDate:                   stde,
			EndDate:                     enddt,
		})
	}

	c.HTML(http.StatusOK, "weeklyreportsummary.html", gin.H{
		"data": rtn,
		"cid":  customerid,
	})
}
