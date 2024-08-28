package services

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func generateCustomerReport(ctx context.Context, customerID, startDate, endDate string) ([]models.CustomerWeeklyReport, error) {
	//取出week值
	weeks := common.GetWeeksInDateRange(common.ParseTime(startDate), common.ParseTime(endDate))
	if len(weeks) > 1 {
		return nil, fmt.Errorf("一次只能產生一週的報表資料。%s ~ %s 跨週了。", startDate, endDate)
	}
	var rtn []models.CustomerWeeklyReport
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
		rtn = append(rtn, models.CustomerWeeklyReport{
			DemoSymbolList: value,
			CustomerID:     customerID,
			YearWeek:       weeks[0],
		})
	}
	return rtn, nil
}

const DBCustomerWeeklyReport = "CustomerWeeklyReport"

// 濃縮成一筆
func GetCustomerReportCurrencyList(ctx context.Context, customerID, startDate, endDate string) ([]models.CustomerWeeklyReport, error) {
	type mapkey struct {
		Symbol   string
		YearWeek string
	}
	var mapData = make(map[mapkey]models.CustomerWeeklyReport)
	//依日期，取出週數
	weeks := common.GetWeeksInDateRange(common.ParseTime(startDate), common.ParseTime(endDate))
	if len(weeks) == 0 || weeks == nil {
		return nil, errors.New("日期區間錯誤。")
	}

	//用來判斷最後一週的資料有沒有產生。
	lastWeek := common.GetWeeksByDate(common.ParseTime(endDate))
	lastWeekinReport := false

	client := getFirestoreClient()

	for _, week := range weeks {
		weekKey := mapkey{
			YearWeek: week,
		}
		//依週數取出資料,放入map裏
		//因為不同週數，Symbol有可能重覆，需要相加起來
		iter := client.Collection(DBCustomerWeeklyReport).
			Where("CustomerID", "==", customerID).
			Where("YearWeek", "==", week).
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

			var data models.CustomerWeeklyReport
			doc.DataTo(&data)
			weekKey.Symbol = data.Symbol
			if weeklyreportbysymbol, exists := mapData[weekKey]; exists {
				weeklyreportbysymbol.Merge(data)
				mapData[weekKey] = weeklyreportbysymbol
			} else {
				mapData[weekKey] = data
			}

			if (data.YearWeek == lastWeek) && !lastWeekinReport {
				lastWeekinReport = true
			}
		}
	}

	if !lastWeekinReport {
		//最後一週報表沒產生，要從DB撈
		lastStartDT, lastEndDT, err := common.WeekToDateRange(lastWeek)
		if err != nil {
			return nil, err
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
		weekKey := mapkey{
			YearWeek: lastWeek,
		}
		for _, lastWeeklyReport := range latestReports {
			weekKey.Symbol = lastWeeklyReport.Symbol
			if weeklyreportbysymbol, exists := mapData[weekKey]; exists {
				weeklyreportbysymbol.Merge(lastWeeklyReport)
				mapData[weekKey] = weeklyreportbysymbol
			} else {
				mapData[weekKey] = lastWeeklyReport
			}
		}

		//最後一週有資料且還沒結束，先取得最新的週數
		systemlastWeek := common.GetWeeksByDate(time.Now().UTC())
		if systemlastWeek != lastWeek && len(latestReports) > 0 {
			//表示lastWeek不是當下時間的週數，要寫入DB
			if err := insertWeeklyReportIntoDB(ctx, latestReports); err != nil {
				return nil, err
			}
		}
	}

	var rtn []models.CustomerWeeklyReport
	for _, report := range mapData {
		rtn = append(rtn, report)
	}
	return rtn, nil
}

func insertWeeklyReportIntoDB(ctx context.Context, reports []models.CustomerWeeklyReport) error {
	client := getFirestoreClient()
	for _, data := range reports {
		_, _, err := client.Collection(DBCustomerWeeklyReport).Add(ctx, data)
		if err != nil {
			return nil
		}
	}
	return nil
}

func GetCustomerReportCurrencySummaryList(ctx context.Context, customerID, startDate, endDate string) ([]models.CustomerWeeklyReportSummary, error) {
	var rtn []models.CustomerWeeklyReportSummary
	middleRtn := make(map[string]models.CustomerWeeklyReportSummary)
	weeklyData, err := GetCustomerReportCurrencyList(ctx, customerID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	for _, data := range weeklyData {
		if weeklyreport, exists := middleRtn[data.YearWeek]; exists {
			weeklyreport.Profit += data.Profit
			middleRtn[data.YearWeek] = weeklyreport
		} else {
			middleRtn[data.YearWeek] = models.CustomerWeeklyReportSummary{
				YearWeek: data.YearWeek,
				Profit:   data.Profit,
			}
		}
	}

	for _, value := range middleRtn {
		rtn = append(rtn, value)
	}
	return rtn, nil
}
