package controllers

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type openPosition struct {
	AvailableAmt float64
	PositionSide bingx.PositionSideType
}

func TradingViewWebhook(c *gin.Context) {
	var WebhookData models.TvWebhookData
	if err := c.ShouldBindJSON(&WebhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := preProcessPlaceOrder(c, WebhookData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func preProcessPlaceOrder(c *gin.Context, WebhookData models.TvWebhookData) error {
	// 寫入 WebhookData 到 Firestore
	TvWebHookLog, err := services.SaveWebhookData(context.Background(), WebhookData)
	if err != nil {
		log.Printf("Failed to save webhook data: %v", err.Error())
	}

	//檢查Cert
	_, err = services.GetSymbol(c, WebhookData.Symbol, WebhookData.Cert)
	if err != nil {
		//Todo:要寫Log
		return err
	}

	var tvData models.TvSiginalData
	tvData.Convert(WebhookData)

	//取出有訂閱的人
	customerList, err := services.GetCustomerCurrencySymbosBySymbol(c, WebhookData.Symbol)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < len(customerList); i++ {
		wg.Add(1)
		go func(customer models.CustomerCurrencySymboWithCustomer) {
			defer wg.Done()
			processPlaceOrder(customer, tvData, TvWebHookLog, customer.APIKey, customer.SecretKey)
		}(customerList[i])
	}
	wg.Wait()

	return nil
}

func processPlaceOrder(Customer models.CustomerCurrencySymboWithCustomer, tv models.TvSiginalData, TvWebHookLog, APIKey, SecertKey string) {
	client := bingx.NewClient(APIKey, SecertKey, Customer.Simulation)
	// client.Debug = true
	// 定义日期字符串的格式
	const layout = "2006-01-02 15:04:05"
	ctx := context.Background()

	placeOrderLog := models.Log_TvSiginalData{
		PlaceOrderType: tv.PlaceOrderType,
		CustomerID:     Customer.CustomerID,
		Time:           time.Now().UTC().Format(layout),
		Simulation:     Customer.Simulation,
		WebHookRefID:   TvWebHookLog,
		Symbol:         tv.Symbol,
	}

	//查出目前持倉情況
	positions, err := client.NewGetOpenPositionsService().Symbol(tv.TVData.Symbol).Do(ctx)
	if err != nil {
		placeOrderLog.Result = "Get open position failed."
		asyncWriteTVsignalData(placeOrderLog, ctx)
		return
	}

	var oepntrade openPosition //目前持倉
	var totalAmount float64    //總倉位
	var totalPrice float64     //總成本
	var totalFee float64       //總雜支，包含資金費率、手續費

	//這裏假設由系統來下單，只會持倉固定方向，所以全部累計
	for i, position := range *positions {
		amount := common.Decimal(position.AvailableAmt)
		price := common.Decimal(position.AvgPrice)
		fee := common.Decimal(position.RealisedProfit)

		totalAmount += amount
		totalPrice += price * amount
		totalFee += fee
		if i == 0 {
			if strings.ToLower(position.PositionSide) == "long" {
				oepntrade.PositionSide = bingx.LongPositionSideType
			} else {
				oepntrade.PositionSide = bingx.ShortPositionSideType
			}
		}
	}
	oepntrade.AvailableAmt = totalAmount

	//計算下單數量
	placeAmount := tv.TVData.Contracts * Customer.Amount / 100
	if Customer.Simulation {
		//模擬盤固定使用10000U計算
		placeAmount = tv.TVData.Contracts * 10000 / 100
	}

	if (tv.PlaceOrderType.Side == bingx.BuySideType && tv.PlaceOrderType.PositionSideType == bingx.ShortPositionSideType) ||
		(tv.PlaceOrderType.Side == bingx.SellSideType && tv.PlaceOrderType.PositionSideType == bingx.LongPositionSideType) {
		if oepntrade.AvailableAmt < placeAmount {
			//要防止平太多，變反向持倉
			placeAmount = oepntrade.AvailableAmt
		}
	}

	if tv.TVData.PositionSize == 0 {
		//全部平倉
		placeAmount = oepntrade.AvailableAmt
	}

	if placeAmount == 0 {
		placeOrderLog.Result = "Place amount is 0."
		asyncWriteTVsignalData(placeOrderLog, ctx)
		return
	}

	//下單
	order, err := client.NewCreateOrderService().
		PositionSide(tv.PlaceOrderType.PositionSideType).
		Symbol(tv.TVData.Symbol).
		Quantity(placeAmount).
		Type(bingx.MarketOrderType).
		Side(tv.PlaceOrderType.Side).
		Do(ctx)

	//如果下單有問題，就記錄下來後return
	if err != nil {
		placeOrderLog.Result = "Place order failed:" + err.Error()
		asyncWriteTVsignalData(placeOrderLog, ctx)
		return
	}
	log.Printf("Customer:%s %v order created: %+v", Customer.CustomerID, bingx.MarketOrderType, order)

	//寫入訂單編號
	placeOrderLog.Result = strconv.FormatInt((*order).OrderId, 10)
	placeOrderLog.Amount = placeAmount

	//依訂單編號，取出下單結果，用來記錄amount及price
	placedOrder, err := client.NewGetOrderService().
		Symbol(tv.TVData.Symbol).
		OrderId(order.OrderId).
		Do(ctx)

	//無法取得下單的資料
	if err != nil {
		placeOrderLog.Result = placeOrderLog.Result + "\nGet placed order failed:" + err.Error()
	}

	//profit := common.Decimal(placedOrder.Profit)
	placedPrice := common.Decimal(placedOrder.AveragePrice)
	fee := common.Decimal(placedOrder.Fee)

	//placeOrderLog.Profit = profit //取不到值，所以要自己算。
	placeOrderLog.Price = placedPrice

	if tv.TVData.PositionSize == 0 {
		//平倉，計算收益
		totalFee = totalFee + fee
		placeValue := placedPrice * placeAmount //成交額

		if strings.ToLower(string(placedOrder.PositionSide)) == "long" {
			//平多，place - 持倉
			placeOrderLog.Profit = placeValue - totalPrice
		} else {
			//平空，持倉 - place
			placeOrderLog.Profit = totalPrice - placeValue
		}
		placeOrderLog.Profit = common.Decimal(placeOrderLog.Profit)
		placeOrderLog.Fee = totalFee
	}

	asyncWriteTVsignalData(placeOrderLog, ctx)
}

// 寫log
func asyncWriteTVsignalData(tvdata models.Log_TvSiginalData, c context.Context) {
	go func(data models.Log_TvSiginalData) {
		_, err := services.SaveCustomerPlaceOrderResultLog(c, data)
		if err != nil {
			log.Printf("Failed to save webhook data: %v", err)
		}
	}(tvdata)
}
