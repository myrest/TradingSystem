package bingx

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func (client *Client) CreateOrder(c context.Context, tv models.TvSiginalData, Customer models.CustomerCurrencySymboWithCustomer) (models.Log_TvSiginalData, bool, models.AlertMessageModel, error) {
	// client.Debug = true
	// 定义日期字符串的格式
	AlertMessageModel := models.CustomerAlertDefault
	placeOrderLog := models.Log_TvSiginalData{
		PlaceOrderType: tv.PlaceOrderType,
		CustomerID:     Customer.CustomerID,
		Time:           common.GetUtcTimeNow(),
		Simulation:     Customer.Simulation,
		Symbol:         tv.Symbol,
	}

	var isTowWayPositionOnHand = false //是否雙向持倉

	//查出目前持倉情況
	positions, err := client.NewGetOpenPositionsService().Symbol(tv.TVData.Symbol).Do(c)
	if err != nil {
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, err
	}

	var oepntrade models.OpenPosition //目前持倉
	var totalAmount float64           //總倉位
	var totalPrice float64            //總成本
	var totalFee float64              //總雜支，包含資金費率、手續費

	//這裏假設由系統來下單，應該只會持倉固定方向
	for i, position := range *positions {
		if strings.ToUpper(position.PositionSide) == string(tv.PositionSideType) {
			amount := common.Decimal(position.AvailableAmt)
			price := common.Decimal(position.AvgPrice)
			fee := common.Decimal(position.RealisedProfit)

			totalAmount += amount
			totalPrice += price * amount
			totalFee += fee
			if i == 0 {
				if strings.ToLower(position.PositionSide) == "long" {
					oepntrade.PositionSide = models.LongPositionSideType
				} else {
					oepntrade.PositionSide = models.ShortPositionSideType
				}
			}
		} else {
			//有雙向持倉的情形發生，要發警告
			isTowWayPositionOnHand = true
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
		//模擬盤固定使用1000U計算
		placeAmount = tv.TVData.Contracts * 1000 / 100
	}

	if (tv.PlaceOrderType.Side == models.BuySideType && tv.PlaceOrderType.PositionSideType == models.ShortPositionSideType) ||
		(tv.PlaceOrderType.Side == models.SellSideType && tv.PlaceOrderType.PositionSideType == models.LongPositionSideType) {
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
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, fmt.Errorf("下單數量為 0。")
	}

	//下單
	//PositionSideType及SideType要轉型，大小寫必需相同
	order, err := client.NewCreateOrderService().
		PositionSide(PositionSideType(tv.PlaceOrderType.PositionSideType)).
		Symbol(tv.TVData.Symbol).
		Quantity(placeAmount).
		Type(MarketOrderType).
		Side(SideType(tv.PlaceOrderType.Side)).
		Do(c)

	//如果下單有問題，就return
	if err != nil {
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, err
	}
	log.Printf("Customer:%s %v order created: %+v", Customer.CustomerID, MarketOrderType, order)

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
		placedOrder = &GetOrderResponse{}
	}

	profit := common.Decimal(placedOrder.Profit)
	placedPrice := common.Decimal(placedOrder.AveragePrice)
	fee := common.Decimal(placedOrder.Fee)

	placeOrderLog.Profit = profit
	placeOrderLog.Price = placedPrice

	if tv.TVData.PositionSize == 0 {
		//平倉，計算收益
		totalFee = totalFee + fee
		placeOrderLog.Fee = totalFee
	}
	if placeOrderLog.Profit < 0 {
		//虧損
		AlertMessageModel = models.CustomerAlertLoss
	} else if (tv.PositionSideType == models.ShortPositionSideType && tv.Side == models.BuySideType) ||
		(tv.PositionSideType == models.LongPositionSideType && tv.Side == models.SellSideType) {
		//平倉
		AlertMessageModel = models.CustomerAlertClose
	} else {
		//有下單
		AlertMessageModel = models.CustomerAlertAll
	}

	return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, nil
}
