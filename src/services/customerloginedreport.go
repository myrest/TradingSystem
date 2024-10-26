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

func generateCustomerReport(ctx context.Context, customerID string, startDate, endDate time.Time) ([]models.CustomerProfitReport, error) {
	//取出week值
	weeks := common.GetWeeksInDateRange(startDate, endDate)
	if len(weeks) > 1 {
		return nil, fmt.Errorf("一次只能產生一週的報表資料。%s ~ %s 跨週了。", startDate, endDate)
	}
	var rtn []models.CustomerProfitReport
	client := getFirestoreClient()
	//先找出所有的History
	iter := client.Collection("placeOrderLog").
		Where("CustomerID", "==", customerID).
		Where("Simulation", "==", false).
		Where("Time", ">=", common.FormatDate(startDate)).
		Where("Time", "<", common.FormatTime(endDate)).
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
			YearUnit:       weeks[0],
		})
	}
	return rtn, nil
}

const DBCustomerWeeklyReport = "CustomerWeeklyReport"
const DBCustomerMonthlyReport = "CustomerMonthlyReport"

type reportMapkey struct {
	Symbol   string
	YearUnit string
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

func GetCustomerWeeklyReportCurrencyList(ctx context.Context, customerID string, startDate, endDate time.Time) ([]models.CustomerProfitReport, error) {
	//找出客戶的第一筆資料，如果起始日期早於它，則以第一筆資料為起始日期
	firstPlaceOrderTime := getCustomerFirstPlaceOrderDateTime(ctx, customerID)
	if startDate.Before(firstPlaceOrderTime) {
		startDate = firstPlaceOrderTime
	}

	var mapData = make(map[reportMapkey]models.CustomerProfitReport)
	//依日期，取出週數
	weeks := common.GetWeeksInDateRange(startDate, endDate)
	if len(weeks) == 0 || weeks == nil {
		//如果日期區間的endDate在第一筆下單之前，會導致startDate > endDate，此時，直接回傳空值就好。
		//return nil, errors.New("日期區間錯誤。")
		return nil, nil
	}

	client := getFirestoreClient()

	missingWeeks := make(map[string]struct{}, len(weeks))
	for _, week := range weeks {
		missingWeeks[week] = struct{}{} // 將 weeks 中的項目存入集合
	}

	//因為不同週數，Symbol有可能重覆，需要相加起來
	iter := client.Collection(DBCustomerWeeklyReport).
		Where("CustomerID", "==", customerID).
		Where("YearUnit", "in", weeks).
		OrderBy("YearUnit", firestore.Asc).
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
			YearUnit: data.YearUnit,
			Symbol:   data.Symbol,
		}
		if weeklyreportbysymbol, exists := mapData[weekKey]; exists {
			weeklyreportbysymbol.Merge(data)
			mapData[weekKey] = weeklyreportbysymbol
		} else {
			mapData[weekKey] = data
		}

		// 移除已找到的週數
		delete(missingWeeks, data.YearUnit)
	}

	//處理尚未產生的週資料
	for week := range missingWeeks {
		getWeekReport(ctx, week, customerID, mapData)
	}

	var rtn []models.CustomerProfitReport
	for _, report := range mapData {
		rtn = append(rtn, report)
	}
	return rtn, nil
}

func GetCustomerMonthlyReportCurrencyList(ctx context.Context, customerID string, startDate, endDate time.Time) ([]models.CustomerProfitReport, error) {
	var mapData = make(map[reportMapkey]models.CustomerProfitReport)
	client := getFirestoreClient()

	//找出客戶的第一筆資料，如果起始日期早於它，則以第一筆資料為起始日期
	firstPlaceOrderTime := getCustomerFirstPlaceOrderDateTime(ctx, customerID)
	if startDate.Before(firstPlaceOrderTime) {
		startDate = firstPlaceOrderTime
	}

	//取得月份列表
	months := common.GetMonthsInRange(startDate, endDate)

	if len(months) == 0 || months == nil {
		//如果日期區間的endDate在第一筆下單之前，會導致startDate > endDate，此時，直接回傳空值就好。
		//return nil, errors.New("日期區間錯誤。")
		return nil, nil
	}

	missingMonths := make(map[string]struct{}, len(months))
	for _, m := range months {
		missingMonths[m] = struct{}{} // 將 Month 中的項目存入集合
	}

	//先從DB的月報找資料，再把同月份的累計起來
	iter := client.Collection(DBCustomerMonthlyReport).
		Where("CustomerID", "==", customerID).
		Where("YearUnit", "in", months).
		OrderBy("YearUnit", firestore.Asc).
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
		monthKey := reportMapkey{
			YearUnit: data.YearUnit,
			Symbol:   data.Symbol,
		}
		if monthlyreportbysymbol, exists := mapData[monthKey]; exists {
			monthlyreportbysymbol.Merge(data)
			mapData[monthKey] = monthlyreportbysymbol
		} else {
			mapData[monthKey] = data
		}

		// 移除已找到的月份
		delete(missingMonths, data.YearUnit)
	}

	//處理尚未產生的週資料
	for month := range missingMonths {
		getMonthReport(ctx, month, customerID, mapData)
	}

	var rtn []models.CustomerProfitReport
	for _, report := range mapData {
		rtn = append(rtn, report)
	}
	return rtn, nil
}

// 依月份、Currency建立月報資料
func getMonthReport(ctx context.Context, month string, customerID string, mapData map[reportMapkey]models.CustomerProfitReport) error {
	//傳入的map資料中，不應該有相同月份的資料。
	for key := range mapData {
		if key.YearUnit == month {
			return errors.New("己有相同月份資料")
		}
	}

	sdt, edt := common.GetMonthStartEndDate(common.ParseTime(month))
	//一次撈一週資料
	client := getFirestoreClient()
	//先找出所有的History，
	iter := client.Collection("placeOrderLog").
		Where("CustomerID", "==", customerID).
		Where("Simulation", "==", false).
		Where("Time", ">=", common.FormatDate(sdt)).
		Where("Time", "<", common.FormatTime(edt)).
		Documents(ctx)
	defer iter.Stop()
	symbollist := make(map[string]models.CustomerProfitReport) //Symbol -> Report 資料

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		var log models.Log_TvSiginalData
		if err := doc.DataTo(&log); err != nil {
			return err
		}

		if log.Price == 0 { //開單失敗，所以沒價格資料，跳過
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
			symbollist[log.Symbol] = models.CustomerProfitReport{
				DemoSymbolList: models.DemoSymbolList{
					Symbol:     log.Symbol,
					Profit:     log.Profit + log.Fee,
					CloseCount: CloseCount,
					OpenCount:  OpenCount,
					WinCount:   WinCount,
					LossCount:  LossCount,
					Amount:     Amount,
				},
				CustomerID: customerID,
				YearUnit:   month,
			}
		}
	}

	currentMonth := common.GetMonthsInRange(common.GetMonthStartEndDate(time.Now().UTC()))

	//計算勝率，並寫入原mapData資料裏
	for _, value := range symbollist {
		winrate := float64(value.WinCount) / float64(value.WinCount+value.LossCount) * 100
		value.Winrate = formatWinRate(winrate)
		value.Profit = common.Decimal(value.Profit, 2)
		datamapkey := reportMapkey{
			Symbol:   value.Symbol,
			YearUnit: value.YearUnit,
		}
		mapData[datamapkey] = value

		if currentMonth[0] != value.YearUnit { //當月資料因為還沒結束，所以不寫入
			//Todo: 應該要檢查是否己存在DB裏
			_, _, err := client.Collection(DBCustomerMonthlyReport).Add(ctx, value)
			if err != nil {
				return nil
			}
		}
	}
	return nil
}

func getWeekReport(ctx context.Context, currentWeek string, customerID string, mapData map[reportMapkey]models.CustomerProfitReport) error {
	//傳入的map資料中，不應該有相同週數的資料。
	for key := range mapData {
		if key.YearUnit == currentWeek {
			return errors.New("己有相同週數的資料")
		}
	}

	//判斷最後一週有資料且還沒結束，先取得最新的週數
	lastStartDT, lastEndDT, _ := common.WeekToDateRange(currentWeek)
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
			YearUnit: currentWeek,
			Symbol:   lastWeeklyReport.Symbol,
		}
		if weeklyreportbysymbol, exists := mapData[weekKey]; exists {
			weeklyreportbysymbol.Merge(lastWeeklyReport)
			mapData[weekKey] = weeklyreportbysymbol
		} else {
			mapData[weekKey] = lastWeeklyReport
		}
	}

	systemWeekNow := common.GetWeeksByDate(time.Now().UTC())
	//表示currentWeek不是當下時間的週數，要寫入DB
	if systemWeekNow != currentWeek && len(latestReports) > 0 {
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

func GetCustomerReportCurrencySummaryList(ctx context.Context, customerID string, startDate, endDate time.Time) ([]models.CustomerReportSummary, error) {
	var rtn []models.CustomerReportSummary
	middleRtn := make(map[string]models.CustomerReportSummary)
	weeklyData, err := GetCustomerWeeklyReportCurrencyList(ctx, customerID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	for _, data := range weeklyData {
		if weeklyreport, exists := middleRtn[data.YearUnit]; exists {
			weeklyreport.Profit += data.Profit
			middleRtn[data.YearUnit] = weeklyreport
		} else {
			middleRtn[data.YearUnit] = models.CustomerReportSummary{
				YearUnit: data.YearUnit,
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
		fmt.Sscanf(rtn[i].YearUnit, "%d-%d", &year1, &week1)
		fmt.Sscanf(rtn[j].YearUnit, "%d-%d", &year2, &week2)

		// 先比較年份，再比較週數
		if year1 != year2 {
			return year1 > year2 // 降冪排序
		}
		return week1 > week2 // 降冪排序
	})
	return rtn, nil
}

func GetCustomerReportCurrencySummaryListMonthly(ctx context.Context, customerID string, startDate, endDate time.Time) ([]models.CustomerReportSummary, error) {
	var rtn []models.CustomerReportSummary
	middleRtn := make(map[string]models.CustomerReportSummary)
	weeklyData, err := GetCustomerMonthlyReportCurrencyList(ctx, customerID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	for _, data := range weeklyData {
		if weeklyreport, exists := middleRtn[data.YearUnit]; exists {
			weeklyreport.Profit += data.Profit
			middleRtn[data.YearUnit] = weeklyreport
		} else {
			middleRtn[data.YearUnit] = models.CustomerReportSummary{
				YearUnit: data.YearUnit,
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
		fmt.Sscanf(rtn[i].YearUnit, "%d-%d", &year1, &week1)
		fmt.Sscanf(rtn[j].YearUnit, "%d-%d", &year2, &week2)

		// 先比較年份，再比較週數
		if year1 != year2 {
			return year1 > year2 // 降冪排序
		}
		return week1 > week2 // 降冪排序
	})
	return rtn, nil
}
