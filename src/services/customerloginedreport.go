package services

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/api/iterator"
)

func generateCustomerReport(ctx context.Context, customerID, startDate, endDate string) ([]models.CustomerWeeklyReport, error) {
	var rtn []models.CustomerWeeklyReport
	client := getFirestoreClient()
	//先找出所有的History
	iter := client.Collection("placeOrderLog").
		Where("CustomerID", "==", customerID).
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
		})
	}
	return rtn, nil
}

const DBCustomerWeeklyReport = "CustomerWeeklyReport"

func GetCustomerReportCurrencyList(ctx context.Context, customerID, startDate, endDate string, BoolFlags ...bool) ([]models.CustomerWeeklyReport, error) {
	var rtn []models.CustomerWeeklyReport
	var mapData = make(map[string]models.CustomerWeeklyReport)
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
		//依週數取出資料,放入map裏
		//因為不同週數，Symbol有可能重覆，需要相加起來
		iter := client.Collection(DBCustomerWeeklyReport).
			Where("CustomerID", "==", customerID).
			Where("YearWeek", "=", week).
			Documents(ctx)
		defer iter.Stop()

		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var data models.CustomerWeeklyReport
		doc.DataTo(&data)
		if weeklyreportbysymbol, exists := mapData[data.Symbol]; exists {
			inlineReort := mapData[data.Symbol]
			inlineReort.Merge(weeklyreportbysymbol)
			mapData[data.Symbol] = inlineReort
		} else {
			mapData[data.Symbol] = data
		}

		if (data.YearWeek == lastWeek) && !lastWeekinReport {
			lastWeekinReport = true
		}
	}

	if !lastWeekinReport {
		//最後一週報表沒產生，要從DB撈
		latestReports, err := generateCustomerReport(ctx, customerID, startDate, endDate)
		if err != nil {
			//寫入log
			CustomerEventLog{
				EventName:  WeeklyReport,
				CustomerID: customerID,
				Message:    fmt.Sprintf("產生報表錯誤：%s", err.Error()),
			}.SendWithoutIP()
		}
		for _, lastWeeklyReport := range latestReports {
			if weeklyreportbysymbol, exists := mapData[lastWeeklyReport.Symbol]; exists {
				inlineReort := mapData[lastWeeklyReport.Symbol]
				inlineReort.Merge(weeklyreportbysymbol)
				mapData[lastWeeklyReport.Symbol] = inlineReort
			} else {
				mapData[lastWeeklyReport.Symbol] = lastWeeklyReport
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
	return rtn, nil
}

func insertWeeklyReportIntoDB(ctx context.Context, reports []models.CustomerWeeklyReport) error {
	// DBCustomerWeeklyReport
	client := getFirestoreClient()
	for _, data := range reports {
		_, _, err := client.Collection("DBCustomerWeeklyReport").Add(ctx, data)
		if err != nil {
			return nil
		}
	}
	return nil
}
