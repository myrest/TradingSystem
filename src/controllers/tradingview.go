package controllers

import (
	"TradingSystem/src/bingx"
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

	err := preProcessPlaceOrder(c, WebhookData, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
}

func TradingViewWebhookTEST(c *gin.Context) {
	var WebhookData models.TvWebhookData
	if err := c.ShouldBindJSON(&WebhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := preProcessPlaceOrder(c, WebhookData, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func preProcessPlaceOrder(c *gin.Context, WebhookData models.TvWebhookData, isTEST bool) error {
	// 寫入 WebhookData 到 Firestore
	go func(data models.TvWebhookData) {
		err := services.SaveWebhookData(context.Background(), data)
		if err != nil {
			log.Printf("Failed to save webhook data: %v", err)
		}
	}(WebhookData)

	//檢查Cert
	_, err := services.GetSymbo(c, WebhookData.Symbol, WebhookData.Cert)
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
			processPlaceOrder(customer.CustomerID, customer.APIKey, customer.SecretKey, customer.Amount, tvData, isTEST)
		}(customerList[i])
	}
	wg.Wait()

	return nil
}

func processPlaceOrder(CustomerID, APIKey, SecertKey string, amount float64, tv models.TvSiginalData, isTEST bool) {
	client := bingx.NewClient(APIKey, SecertKey, isTEST)

	placeOrderLog := models.Log_TvSiginalData{
		TVData:     tv.TVData,
		Profit:     0,
		CustomerID: CustomerID,
		Time:       time.Now().Unix(),
	}

	//查出目前持倉情況
	positions, err := client.NewGetOpenPositionsService().Symbol(tv.TVData.Symbol).Do(context.Background())
	if err != nil {
		//Todo:要寫Log
		placeOrderLog.Result = "Get open position failed."
		asyncWriteTVsignalData(placeOrderLog)
		return
	}

	var oepntrade openPosition //目前持倉
	if len(*positions) > 0 {
		amount, _ := strconv.ParseFloat((*positions)[0].AvailableAmt, 64)
		oepntrade.AvailableAmt = amount
		if strings.ToLower((*positions)[0].PositionSide) == "long" {
			oepntrade.PositionSide = bingx.LongPositionSideType
		} else {
			oepntrade.PositionSide = bingx.ShortPositionSideType
		}
	}
	//計算下單數量
	//isClosePosition := false
	var profit float64
	placeAmount := tv.TVData.Contracts * amount / 100
	if (tv.PlaceOrderType.Side == bingx.BuySideType && tv.PlaceOrderType.PositionSideType == bingx.ShortPositionSideType) ||
		(tv.PlaceOrderType.Side == bingx.SellSideType && tv.PlaceOrderType.PositionSideType == bingx.LongPositionSideType) {
		if oepntrade.AvailableAmt < placeAmount {
			//要防止平太多，變反向持倉
			placeAmount = oepntrade.AvailableAmt
		}
		//isClosePosition = true
	}

	if tv.TVData.PositionSize == 0 {
		//全部平倉
		placeAmount = oepntrade.AvailableAmt
	}

	if placeAmount == 0 {
		placeOrderLog.Result = "No open poition for close."
		asyncWriteTVsignalData(placeOrderLog)
		return
	}

	//下單
	order, err := client.NewCreateOrderService().
		PositionSide(tv.PlaceOrderType.PositionSideType). //bingx.LongPositionSideType
		Symbol(tv.TVData.Symbol).
		Quantity(placeAmount).
		Type(bingx.MarketOrderType).
		Side(tv.PlaceOrderType.Side).
		Do(context.Background())

	//如果下單有問題，就記錄下來後return
	if err != nil {
		placeOrderLog.Result = "Place order failed:" + err.Error()
		asyncWriteTVsignalData(placeOrderLog)
		return
	}
	log.Printf("Customer:%s %v order created: %+v", CustomerID, bingx.MarketOrderType, order)

	//寫入訂單編號
	placeOrderLog.Result = strconv.Itoa((*order).OrderId)
	placeOrderLog.Amount = placeAmount

	//依訂單編號，取出下單結果，用來記錄amount及price
	placedOrder, err := client.NewGetOrderService().
		Symbol(tv.TVData.Symbol).
		ClientOrderId(strconv.Itoa(order.OrderId)).
		Do(context.Background())

	//無法取得下單的資料
	if err != nil {
		placeOrderLog.Result = placeOrderLog.Result + "\nGet placed order failed:" + err.Error()
	}

	profit, _ = strconv.ParseFloat(placedOrder.Profit, 64)
	placedPrice, _ := strconv.ParseFloat(placedOrder.Price, 64)

	placeOrderLog.Profit = profit
	placeOrderLog.Price = placedPrice
	asyncWriteTVsignalData(placeOrderLog)
}

// 寫log
func asyncWriteTVsignalData(tvdata models.Log_TvSiginalData) {
	go func(data models.Log_TvSiginalData) {
		err := services.SaveCustomerPlaceOrderResultLog(context.Background(), data)
		if err != nil {
			log.Printf("Failed to save webhook data: %v", err)
		}
	}(tvdata)
}
