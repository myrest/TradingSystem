package services

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func GetDemoCurrencyList(ctx context.Context, numberOfDays int, isFromCache bool) ([]models.DemoSymbolList, error) {
	const layout = "2006-01-02"
	DaysAgo := time.Now().UTC().AddDate(0, 0, numberOfDays*-1).Format(layout)
	cacheKey := getLog_TVCacheKey(systemsettings.DemoCustomerID, "SymbolList", strconv.Itoa(numberOfDays))

	if isFromCache {
		data, err := loadDemoSymbolListCache(cacheKey)
		if err == nil {
			return data, nil
		}
	}

	client := common.GetFirestoreClient()
	//先找出所有的History
	iter := client.Collection("placeOrderLog").
		Where("CustomerID", "==", systemsettings.DemoCustomerID).
		Where("Time", ">", DaysAgo).
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

		if log.Price == 0 {
			continue
		}

		symbol, exists := symbollist[log.Symbol]
		var CloseCount, OpenCount, WinCount, LossCount int32
		Amount := log.Amount * log.Price
		if (log.Side == models.SellSideType && log.PositionSideType == models.LongPositionSideType) ||
			(log.Side == models.BuySideType && log.PositionSideType == models.ShortPositionSideType) {
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

	// 写入缓存
	err := saveDemoSymbolListCache(cacheKey, rtn)
	if err != nil {
		return nil, err
	}

	return rtn, nil
}

func GetDemoHistory(ctx context.Context, numberOfDays int, Symbol string, isFromCache bool) ([]models.Log_TvSiginalData, error) {
	const layout = "2006-01-02"
	DaysAgo := time.Now().UTC().AddDate(0, 0, numberOfDays*-1).Format(layout)
	cacheKey := getLog_TVCacheKey(systemsettings.DemoCustomerID, Symbol, strconv.Itoa(numberOfDays))

	if isFromCache {
		data, err := loadLog_TVCache(cacheKey)
		if err == nil {
			return data, nil
		}
	}

	client := common.GetFirestoreClient()
	// 先找出所有的History
	iter := client.Collection("placeOrderLog").
		Where("CustomerID", "==", systemsettings.DemoCustomerID).
		Where("Symbol", "==", Symbol).
		Where("Time", ">", DaysAgo).
		OrderBy("Time", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	rtn := []models.Log_TvSiginalData{}

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
		if log.Price == 0 {
			continue
		}
		rtn = append(rtn, log)
	}

	// 写入缓存
	err := saveLog_TVCache(cacheKey, rtn)
	if err != nil {
		return nil, err
	}

	return rtn, nil
}

func formatWinRate(winrate float64) string {
	if math.IsNaN(winrate) {
		winrate = 0
	}
	rounded := math.Round(winrate*100) / 100 // 四舍五入到小数点后两位
	return fmt.Sprintf("%.2f%%", rounded)
}
