package binance_connector

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"fmt"
	"log"
	"strconv"
)

func (client *Client) CreateOrder(c context.Context, tv models.TvSiginalData, Customer models.CustomerCurrencySymboWithCustomer) (models.Log_TvSiginalData, bool, models.AlertMessageModel, error) {
	client.Debug = true
	AlertMessageModel := models.CustomerAlertDefault
	placeOrderLog := models.Log_TvSiginalData{
		PlaceOrderType: tv.PlaceOrderType,
		CustomerID:     Customer.CustomerID,
		Time:           common.GetUtcTimeNow(),
		Simulation:     Customer.Simulation,
		Symbol:         tv.Symbol,
	}

	var isTowWayPositionOnHand = false //是否雙向持倉
	var isCloseOrder = false           //是否為平倉

	//判斷是不是平倉
	if (tv.PlaceOrderType.Side == models.BuySideType && tv.PlaceOrderType.PositionSideType == models.ShortPositionSideType) ||
		(tv.PlaceOrderType.Side == models.SellSideType && tv.PlaceOrderType.PositionSideType == models.LongPositionSideType) {
		isCloseOrder = true
	}

	//查出目前持倉情況
	positions, err := client.GetUMPositionRiskService().Symbol(tv.TVData.Symbol).Do(c)
	if err != nil {
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, err
	}
	var LongPosition UMPositionRiskResponse
	var ShortPosition UMPositionRiskResponse

	//判斷是否為雙向持倉，己為雙向持倉不用再判斷
	if !isTowWayPositionOnHand && len(positions) == 2 {
		isTowWayPositionOnHand = true
	}

	for _, position := range positions {
		if position.PositionSide == "LONG" {
			LongPosition = *position
		} else {
			ShortPosition = *position
		}
	}

	//設定槓桿
	Leverage := Customer.Leverage
	if Leverage == 0 { //向下相容，為了舊客戶，沒有Leverage設定
		Leverage = 10
	}
	//計算下單數量
	placeAmount := tv.TVData.Contracts * Customer.Amount * Customer.Leverage / 1000
	if Customer.Simulation {
		//模擬盤固定使用1000U計算
		placeAmount = tv.TVData.Contracts * 100 / 100
	}

	if tv.TVData.PositionSize == 0 || isCloseOrder {
		//如果平空或平多，超目當下持倉數量，則改為目當下持倉數量，防止變成反向持倉
		if tv.PlaceOrderType.PositionSideType == models.LongPositionSideType && placeAmount > common.Decimal(LongPosition.PositionAmt) {
			isCloseOrder = true
		}
		if tv.PlaceOrderType.PositionSideType == models.LongPositionSideType && placeAmount > (-1*common.Decimal(ShortPosition.PositionAmt)) {
			isCloseOrder = true
		}

		if isCloseOrder {
			if tv.PlaceOrderType.PositionSideType == models.LongPositionSideType {
				placeAmount = common.Decimal(LongPosition.PositionAmt)
			} else {
				placeAmount = -1 * common.Decimal(ShortPosition.PositionAmt)
			}
		}
	}

	if placeAmount == 0 {
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, fmt.Errorf("下單數量為 0。")
	}

	//下市價單
	//PositionSideType及SideType要轉型，大小寫必需相同
	order, err := client.GetUMNewOrderService().
		PositionSide(PositionSide(tv.PlaceOrderType.PositionSideType)).
		Symbol(tv.TVData.Symbol).
		Quantity(placeAmount).
		Type(MarketOrder).
		Side(OrderSide(tv.PlaceOrderType.Side)).
		Do(c)

	//如果下單有問題，就return
	if err != nil {
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, err
	}
	log.Printf("Customer:%s %v order created: %+v", Customer.CustomerID, MarketOrder, order)

	//取出下單結果，用來記錄amount及price，只能取出歷史記錄來判斷
	history, err := client.GetUMUserTradeService().Symbol(tv.TVData.Symbol).Limit(1).Do(c)
	placedOrder := &UMUserTradeResponse{}

	//無法取得下單的資料
	if (err != nil) || (history == nil) || len(history) == 0 {
		placeOrderLog.Result = placeOrderLog.Result + "\nGet placed order failed:" + err.Error()

	}
	placedOrder = history[0]
	//因為只取最近成交的一筆，所以訂單編號應該要一致才對
	if placedOrder.ID != order.OrderId {
		placeOrderLog.Result = placeOrderLog.Result + "\n查無成交記錄。需要手動修正盈虧及手續費"
	}
	//依下單結果補足資料
	//寫入訂單編號
	placeOrderLog.Result = strconv.FormatInt((*order).OrderId, 10)
	placeOrderLog.Amount = placeAmount
	placeOrderLog.Price = common.Decimal(order.AvgPrice)
	placeOrderLog.Fee = common.Decimal(placedOrder.Commission) * 2 //因為有開、平倉，以平倉值兩倍為大約值
	if isCloseOrder {
		//平倉才有profit值
		placeOrderLog.Profit = common.Decimal(placedOrder.RealizedPnl)
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

func (Client *Client) GetBalance(ctx context.Context) (float64, error) {
	result, err := Client.GetUMAccountBalanceService().Asset("USDT").DoSingle(ctx)
	if err != nil {
		return 0, err
	}

	if (result == nil) || (result.UMWalletBalance == "") {
		return 0, fmt.Errorf("帳戶無USDT餘額。")
	}

	return strconv.ParseFloat(result.UMWalletBalance, 64)
}
