package services

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func generateCustomerReport(ctx context.Context, customerID, startDate, endDate string) ([]models.CustomerProfitReport, error) {
	//取出week值
	weeks := common.GetWeeksInDateRange(common.ParseTime(startDate), common.ParseTime(endDate))
	if len(weeks) > 1 {
		return nil, fmt.Errorf("一次只能產生一週的報表資料。%s ~ %s 跨週了。", startDate, endDate)
	}
	var rtn []models.CustomerProfitReport
	client := getFirestoreClient()
	//先找出所有的History
	iter := client.Collection("placeOrderLog").
		Where("CustomerID", "==", customerID).
		Where("Simulation", "==", false).
		Where("Time", ">=", startDate).
		Where("Time", "<", endDate).
		Documents(ctx)
	defer iter.Stop()
	symbollist := make(map[string]models.DemoSymbolList)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var log models.Log_TvSiginalData
		if err := doc.DataTo(&log); err != nil {
			return nil, err
		}

		if log.Price == 0 { //開單失敗
			continue
		}

		//計算Amount, CloseCount, OpenCount, WinCount, LossCoun，每輪都會歸零
		var CloseCount, OpenCount, WinCount, LossCount int32
		Amount := log.Amount * log.Price //float64
		if (log.Side == bingx.SellSideType && log.PositionSideType == bingx.LongPositionSideType) ||
			(log.Side == bingx.BuySideType && log.PositionSideType == bingx.ShortPositionSideType) {
			CloseCount = 1
			if log.Profit < 0 {
				LossCount = 1
			} else {
				WinCount = 1
			}
		} else {
			OpenCount = 1
		}

		if symbol, exists := symbollist[log.Symbol]; exists {
			symbol.CloseCount += CloseCount
			symbol.OpenCount += OpenCount
			symbol.WinCount += WinCount
			symbol.LossCount += LossCount
			symbol.Profit += log.Profit + log.Fee
			symbol.Amount += Amount
			symbollist[log.Symbol] = symbol
		} else {
			symbollist[log.Symbol] = models.DemoSymbolList{
				Symbol:     log.Symbol,
				Profit:     log.Profit + log.Fee,
				CloseCount: CloseCount,
				OpenCount:  OpenCount,
				WinCount:   WinCount,
				LossCount:  LossCount,
				Amount:     Amount,
			}
		}
	}

	for _, value := range symbollist {
		winrate := float64(value.WinCount) / float64(value.WinCount+value.LossCount) * 100
		value.Winrate = formatWinRate(winrate)
		value.Profit = common.Decimal(value.Profit, 2)
		rtn = append(rtn, models.CustomerProfitReport{
			DemoSymbolList: value,
			CustomerID:     customerID,
			YearWeek:       weeks[0],
		})
	}
	return rtn, nil
}

const DBCustomerWeeklyReport = "CustomerWeeklyReport"

type reportMapkey struct {
	Symbol   string
	YearWeek string
}

func getCustomerFirstPlaceOrderDateTime(ctx context.Context, customerID string) time.Time {
	client := getFirestoreClient()

	// 查詢所有的 placeOrderLog
	iter := client.Collection("placeOrderLog").
		Where("CustomerID", "==", customerID).
		Where("Simulation", "==", false).
		OrderBy("Time", firestore.Asc). // 按時間排序
		Limit(1).                       // 只取第一筆
		Documents(ctx)

	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return common.TimeMax()
	}
	if err != nil {
		return common.TimeMax()
	}

	var log models.Log_TvSiginalData
	if err := doc.DataTo(&log); err != nil {
		return common.TimeMax()
	}
	return common.ParseTime(log.Time)
}

func GetCustomerReportCurrencyList(ctx context.Context, customerID, startDate, endDate string) ([]models.CustomerProfitReport, error) {
	var mapData = make(map[reportMapkey]models.CustomerProfitReport)
	//依日期，取出週數
	weeks := common.GetWeeksInDateRange(common.ParseTime(startDate), common.ParseTime(endDate))
	if len(weeks) == 0 || weeks == nil {
		return nil, errors.New("日期區間錯誤。")
	}

	//用來判斷最後一週的資料有沒有產生。
	lastWeek := common.GetWeeksByDate(common.ParseTime(endDate))
	lastWeekinReport := false

	client := getFirestoreClient()

	missingWeeks := make(map[string]struct{}, len(weeks))
	for _, week := range weeks {
		missingWeeks[week] = struct{}{} // 將 weeks 中的項目存入集合
	}

	//因為不同週數，Symbol有可能重覆，需要相加起來
	iter := client.Collection(DBCustomerWeeklyReport).
		Where("CustomerID", "==", customerID).
		Where("YearWeek", "in", weeks).
		OrderBy("YearWeek", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var data models.CustomerProfitReport
		doc.DataTo(&data)
		weekKey := reportMapkey{
			YearWeek: data.YearWeek,
			Symbol:   data.Symbol,
		}
		if weeklyreportbysymbol, exists := mapData[weekKey]; exists {
			weeklyreportbysymbol.Merge(data)
			mapData[weekKey] = weeklyreportbysymbol
		} else {
			mapData[weekKey] = data
		}

		if (data.YearWeek == lastWeek) && !lastWeekinReport {
			lastWeekinReport = true
		}
		// 移除已找到的週數
		delete(missingWeeks, data.YearWeek)
	}

	//找出客戶的第一筆資料
	firstPlaceOrderTime := getCustomerFirstPlaceOrderDateTime(ctx, customerID)
	//處理尚未產生的週資料
	for week := range missingWeeks {
		_, edt, _ := common.WeekToDateRange(week)
		if common.ParseTime(edt).Before(firstPlaceOrderTime) {
			continue
		}
		getLastWeekReport(ctx, week, customerID, mapData)
	}

	if !lastWeekinReport {
		//最後一週報表沒產生，要從DB撈，其中mapData是傳址
		err := getLastWeekReport(ctx, lastWeek, customerID, mapData)
		if err != nil {
			return nil, err
		}
	}

	var rtn []models.CustomerProfitReport
	for _, report := range mapData {
		rtn = append(rtn, report)
	}
	return rtn, nil
}

func getLastWeekReport(ctx context.Context, lastWeek string, customerID string, mapData map[reportMapkey]models.CustomerProfitReport) error {
	//判斷最後一週有資料且還沒結束，先取得最新的週數
	lastStartDT, lastEndDT, err := common.WeekToDateRange(lastWeek)
	if err != nil {
		return err
	}
	latestReports, err := generateCustomerReport(ctx, customerID, lastStartDT, lastEndDT)
	if err != nil {
		//寫入log
		CustomerEventLog{
			EventName:  WeeklyReport,
			CustomerID: customerID,
			Message:    fmt.Sprintf("產生報表錯誤：%s", err.Error()),
		}.SendWithoutIP()
	}
	for _, lastWeeklyReport := range latestReports {
		weekKey := reportMapkey{
			YearWeek: lastWeek,
			Symbol:   lastWeeklyReport.Symbol,
		}
		if weeklyreportbysymbol, exists := mapData[weekKey]; exists {
			weeklyreportbysymbol.Merge(lastWeeklyReport)
			mapData[weekKey] = weeklyreportbysymbol
		} else {
			mapData[weekKey] = lastWeeklyReport
		}
	}

	systemlastWeek := common.GetWeeksByDate(time.Now().UTC())
	//表示lastWeek不是當下時間的週數，要寫入DB
	if systemlastWeek != lastWeek && len(latestReports) > 0 {
		if err := insertWeeklyReportIntoDB(ctx, latestReports); err != nil {
			return err
		}
	}
	return nil
}

func insertWeeklyReportIntoDB(ctx context.Context, reports []models.CustomerProfitReport) error {
	client := getFirestoreClient()
	for _, data := range reports {
		_, _, err := client.Collection(DBCustomerWeeklyReport).Add(ctx, data)
		if err != nil {
			return nil
		}
	}
	return nil
}

func GetCustomerReportCurrencySummaryList(ctx context.Context, customerID, startDate, endDate string) ([]models.CustomerReportSummary, error) {
	var rtn []models.CustomerReportSummary
	middleRtn := make(map[string]models.CustomerReportSummary)
	weeklyData, err := GetCustomerReportCurrencyList(ctx, customerID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	for _, data := range weeklyData {
		if weeklyreport, exists := middleRtn[data.YearWeek]; exists {
			weeklyreport.Profit += data.Profit
			middleRtn[data.YearWeek] = weeklyreport
		} else {
			middleRtn[data.YearWeek] = models.CustomerReportSummary{
				YearWeek: data.YearWeek,
				Profit:   data.Profit,
			}
		}
	}

	for _, value := range middleRtn {
		rtn = append(rtn, value)
	}

	// 排序切片
	sort.Slice(rtn, func(i, j int) bool {
		// 提取年份和週數
		var year1, week1, year2, week2 int
		fmt.Sscanf(rtn[i].YearWeek, "%d-%d", &year1, &week1)
		fmt.Sscanf(rtn[j].YearWeek, "%d-%d", &year2, &week2)

		// 先比較年份，再比較週數
		if year1 != year2 {
			return year1 > year2 // 降冪排序
		}
		return week1 > week2 // 降冪排序
	})
	return rtn, nil
}

func GetCustomerReportCurrencySummaryListMonthly(ctx context.Context, customerID, startDate, endDate string) ([]models.CustomerReportSummary, error) {
	var rtn []models.CustomerReportSummary
	middleRtn := make(map[string]models.CustomerReportSummary)
	weeklyData, err := GetCustomerReportCurrencyList(ctx, customerID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	for _, data := range weeklyData {
		if weeklyreport, exists := middleRtn[data.YearWeek]; exists {
			weeklyreport.Profit += data.Profit
			middleRtn[data.YearWeek] = weeklyreport
		} else {
			middleRtn[data.YearWeek] = models.CustomerReportSummary{
				YearWeek: data.YearWeek,
				Profit:   data.Profit,
			}
		}
	}

	for _, value := range middleRtn {
		rtn = append(rtn, value)
	}

	// 排序切片
	sort.Slice(rtn, func(i, j int) bool {
		// 提取年份和週數
		var year1, week1, year2, week2 int
		fmt.Sscanf(rtn[i].YearWeek, "%d-%d", &year1, &week1)
		fmt.Sscanf(rtn[j].YearWeek, "%d-%d", &year2, &week2)

		// 先比較年份，再比較週數
		if year1 != year2 {
			return year1 > year2 // 降冪排序
		}
		return week1 > week2 // 降冪排序
	})
	return rtn, nil
}
