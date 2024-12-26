package controllers

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CustomerWeeklyReportList(c *gin.Context) {
	session := sessions.Default(c)
	cid := c.DefaultQuery("cid", "")

	customerid := session.Get("id").(string)

	//只有管理員可以看到其它人的記錄。
	if cid != "" && session.Get("isadmin") != nil && session.Get("isadmin").(bool) {
		customerid = cid
	}

	if customerid == "" {
		c.Error(errors.New("no Customer Data")) // 將錯誤添加到上下文中
		return
	}

	d := c.Query("d")
	reportStartDate := time.Now().UTC()

	if d != "" {
		reportStartDate = common.ParseTime(d)
	}

	startDate, endDate := common.GetWeeksStartEndDateByDate(reportStartDate)

	//將日期區間寫入DB
	common.SetReportStartEndDate(session, startDate, endDate)

	weeklyreport, err := services.GetCustomerWeeklyReportCurrencyList(c, customerid, startDate, endDate)
	if err != nil {
		c.Error(err) // 將錯誤添加到上下文中
		return
	}

	//找出星期一清單
	mondays, err := common.GetPreviousMondays(time.Now().UTC(), 12)
	if err != nil {
		c.Error(err) // 將錯誤添加到上下文中
		return
	}

	mondaysList := []string{}
	for _, day := range mondays {
		mondaysList = append(mondaysList, common.FormatDate(day))
	}

	if session.Get("isadmin") == nil || !session.Get("isadmin").(bool) { //不是管理員cid要清掉不給看
		customerid = ""
	}

	// 排序切片
	sort.Slice(weeklyreport, func(i, j int) bool {
		return weeklyreport[i].Symbol > weeklyreport[j].Symbol // 降冪排序
	})

	c.HTML(http.StatusOK, "weeklyreport.html", gin.H{
		"data":              weeklyreport,
		"mondays":           mondaysList,
		"days":              common.FormatDate(startDate),
		"cid":               customerid,
		"week":              fmt.Sprintf("%s ~ %s (%s)", common.FormatDate(startDate), common.FormatDate(endDate), common.GetWeeksByDate(startDate)),
		"IsAdmin":           c.GetBool("IsAdmin"),
		"StaticFileVersion": systemsettings.StartTimestemp,
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

	reportStartDate := enddate.AddDate(0, -2, 0).Truncate(24 * time.Hour) //去掉時分秒，取兩個月內的資料
	reportEndDate := time.Date(enddate.Year(), enddate.Month(), enddate.Day(), 23, 59, 59, 0, enddate.Location())

	weeklyreport, err := services.GetCustomerReportCurrencySummaryList(c, customerid, reportStartDate, reportEndDate)
	if err != nil {
		return
	}

	if session.Get("isadmin") == nil || !session.Get("isadmin").(bool) { //不是管理員cid要清掉不給看
		customerid = ""
	}
	var rtn []models.CustomerReportSummaryUI
	for _, w := range weeklyreport {
		stde, enddt, _ := common.WeekToDateRange(w.YearUnit)
		w.Profit = common.Decimal(w.Profit, 2)
		rtn = append(rtn, models.CustomerReportSummaryUI{
			CustomerReportSummary: w,
			StartDate:             common.FormatDate(stde),
			EndDate:               common.FormatDate(enddt),
		})
	}

	c.HTML(http.StatusOK, "weeklyreportsummary.html", gin.H{
		"data":              rtn,
		"cid":               customerid,
		"StaticFileVersion": systemsettings.StartTimestemp,
	})
}

func CustomerMonthlyReportSummaryList(c *gin.Context) {
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
	date := time.Now().UTC()

	if d != "" {
		date = common.ParseTime(d)
	}

	sdt, edt := common.GetMonthStartEndDate(date)
	sdt = sdt.AddDate(0, -6, 0) //一次取半年的資料

	monthlyreport, err := services.GetCustomerReportCurrencySummaryListMonthly(c, customerid, sdt, edt)
	if err != nil {
		return
	}

	if session.Get("isadmin") == nil || !session.Get("isadmin").(bool) { //不是管理員cid要清掉不給看
		customerid = ""
	}
	var rtn []models.CustomerReportSummaryUI
	for _, w := range monthlyreport {
		dt := common.ParseTime(w.YearUnit)
		stde, enddt := common.GetMonthStartEndDate(dt)
		w.Profit = common.Decimal(w.Profit, 2)
		rtn = append(rtn, models.CustomerReportSummaryUI{
			CustomerReportSummary: w,
			StartDate:             common.FormatDate(stde),
			EndDate:               common.FormatDate(enddt),
		})
	}

	c.HTML(http.StatusOK, "monthlyreportsummary.html", gin.H{
		"data":              rtn,
		"cid":               customerid,
		"StaticFileVersion": systemsettings.StartTimestemp,
	})
}

func CustomerMonthlyReportList(c *gin.Context) {
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
	//////

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

	if session.Get("isadmin") == nil || !session.Get("isadmin").(bool) { //不是管理員cid要清掉不給看
		customerid = ""
	}

	// 排序切片
	sort.Slice(montylyreport, func(i, j int) bool {
		return montylyreport[i].Symbol > montylyreport[j].Symbol // 降冪排序
	})

	c.HTML(http.StatusOK, "monthlyreport.html", gin.H{
		"data":              montylyreport,
		"mondays":           firstDayByMonth,
		"days":              common.FormatDate(startDate),
		"cid":               customerid,
		"month":             common.GetMonthsInRange(startDate)[0],
		"StaticFileVersion": systemsettings.StartTimestemp,
	})
}
