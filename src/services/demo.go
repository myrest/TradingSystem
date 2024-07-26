package services

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"google.golang.org/api/iterator"
)

func GetDemoCurrencyList(ctx context.Context, numberOfDays int) ([]models.DemoSymbolList, error) {
	const layout = "2006-01-02"
	systemSettings := common.GetEnvironmentSetting()
	DaysAgo := time.Now().UTC().AddDate(0, 0, numberOfDays*-1).Format(layout)
	client := getFirestoreClient()
	//先找出所有的History
	iter := client.Collection("placeOrderLog").
		Where("Time", ">", DaysAgo).
		Where("CustomerID", "==", systemSettings.DemoCustomerID).
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

		symbol, exists := symbollist[log.Symbol]
		var CloseCount, OpenCount, WinCount, LossCount int32
		Amount := log.Amount * log.Price
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
		if exists {
			symbol.CloseCount += CloseCount
			symbol.OpenCount += OpenCount
			symbol.WinCount += WinCount
			symbol.LossCount += LossCount
			symbol.Profit += log.Profit
			symbol.Amount += Amount
			symbollist[log.Symbol] = symbol
		} else {
			symbollist[log.Symbol] = models.DemoSymbolList{
				Symbol:     log.Symbol,
				Profit:     log.Profit,
				CloseCount: CloseCount,
				OpenCount:  OpenCount,
				WinCount:   WinCount,
				LossCount:  LossCount,
				Amount:     Amount,
			}
		}
	}
	var rtn []models.DemoSymbolList

	for _, value := range symbollist {
		winrate := float64(value.WinCount) / float64(value.WinCount+value.LossCount) * 100
		value.Winrate = formatWinRate(winrate)
		value.Profit = common.Decimal(value.Profit, 2)
		rtn = append(rtn, value)
	}
	sort.Slice(rtn, func(i, j int) bool {
		return rtn[i].Profit > rtn[j].Profit
	})

	return rtn, nil
}

func formatWinRate(winrate float64) string {
	if math.IsNaN(winrate) {
		winrate = 0
	}
	rounded := math.Round(winrate*100) / 100 // 四舍五入到小数点后两位
	return fmt.Sprintf("%.2f%%", rounded)
}
