package controllers

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

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
			processPlaceOrder(customer, tvData, TvWebHookLog, customer.APIKey, customer.SecretKey, c)
		}(customerList[i])
	}
	wg.Wait()
	//要更新績效的cache
	go func() {
		//清暫存檔
		services.RemoveLog_TVExpiredCacheFiles()
	}()
	return nil
}

func processPlaceOrder(Customer models.CustomerCurrencySymboWithCustomer, tv models.TvSiginalData, TvWebHookLog, APIKey, SecertKey string, c *gin.Context) {
	client := bingx.NewClient(APIKey, SecertKey, Customer.Simulation)
	// client.Debug = true
	// 定义日期字符串的格式
	needSendToTG := false

	placeOrderLog := models.Log_TvSiginalData{
		PlaceOrderType: tv.PlaceOrderType,
		CustomerID:     Customer.CustomerID,
		Time:           common.GetUtcTimeNow(),
		Simulation:     Customer.Simulation,
		WebHookRefID:   TvWebHookLog,
		Symbol:         tv.Symbol,
	}

	//查出目前持倉情況
	positions, err := client.NewGetOpenPositionsService().Symbol(tv.TVData.Symbol).Do(c)
	if err != nil {
		placeOrderLog.Result = "Get open position failed." + err.Error()
		asyncWriteTVsignalData(needSendToTG, Customer.Customer, placeOrderLog, c)
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
	Leverage := Customer.Leverage
	if Leverage == 0 { //向下相容，為了舊客戶，沒有Leverage設定
		Leverage = 10
	}
	placeAmount := tv.TVData.Contracts * Customer.Amount * Customer.Leverage / 1000
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
		needSendToTG = true
	}

	if placeAmount == 0 {
		placeOrderLog.Result = "Place amount is 0."
		asyncWriteTVsignalData(needSendToTG, Customer.Customer, placeOrderLog, c)
		return
	}

	//下單
	order, err := client.NewCreateOrderService().
		PositionSide(tv.PlaceOrderType.PositionSideType).
		Symbol(tv.TVData.Symbol).
		Quantity(placeAmount).
		Type(bingx.MarketOrderType).
		Side(tv.PlaceOrderType.Side).
		Do(c)

	//如果下單有問題，就記錄下來後return
	if err != nil {
		placeOrderLog.Result = "Place order failed:" + err.Error()
		asyncWriteTVsignalData(true, Customer.Customer, placeOrderLog, c)
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
		Do(c)

	//無法取得下單的資料
	if (err != nil) || (placedOrder == nil) {
		placeOrderLog.Result = placeOrderLog.Result + "\nGet placed order failed:" + err.Error()
		services.SystemEventLog{
			EventName:  services.PlaceOrder,
			CustomerID: Customer.CustomerID,
			Message:    fmt.Sprintf("Get placed order failed:%s\nsymbol:%s\norderId:%d", err.Error(), tv.TVData.Symbol, order.OrderId),
		}.Send()
		placedOrder = &bingx.GetOrderResponse{}
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

	asyncWriteTVsignalData(needSendToTG, Customer.Customer, placeOrderLog, c)
}

// 寫log
func asyncWriteTVsignalData(needSendResult bool, customer models.Customer, tvdata models.Log_TvSiginalData, c *gin.Context) {
	go func(data models.Log_TvSiginalData) {
		_, err := services.SaveCustomerPlaceOrderResultLog(c, data)
		if err != nil {
			log.Printf("Failed to save webhook data: %v", err)
		}
		if needSendResult && customer.TgChatID > 0 {
			positionside := "多"
			side := "開"
			if tvdata.PositionSideType == bingx.ShortPositionSideType {
				positionside = "空"
			}
			if (tvdata.PositionSideType == bingx.ShortPositionSideType && tvdata.Side == bingx.BuySideType) ||
				(tvdata.PositionSideType == bingx.LongPositionSideType && tvdata.Side == bingx.SellSideType) {
				side = "平"
			}
			tgMessage := fmt.Sprintf("幣種：%s \n方向：%s%s\n結果/單號：%s", tvdata.Symbol, side, positionside, tvdata.Result)
			if tvdata.Profit != 0 {
				tgMessage = fmt.Sprintf("%s\n盈虧：%s", tgMessage, formatFloat64(6, tvdata.Profit))
			}
			err := services.TGSendMessage(customer.TgChatID, tgMessage)
			if err != nil {
				services.CustomerEventLog{
					CustomerID: customer.ID,
					EventName:  services.PlaceOrder,
					Message:    err.Error(),
				}.Send(c)
			}
		}
	}(tvdata)
}

func formatFloat64(round int, f float64) string {
	value := common.Decimal(f, round)
	return strconv.FormatFloat(value, 'f', -1, 64)
}
