package binance_connector

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const maxRetries = 3
const waitTime = 500 * time.Millisecond

func (client *Client) CreateOrder(c context.Context, tv models.TvSiginalData, Customer models.CustomerCurrencySymboWithCustomer) (models.Log_TvSiginalData, bool, models.AlertMessageModel, error) {
	//取出Symbol，並改為沒有"-"
	Symbol := strings.Replace(tv.Symbol, "-", "", 1)
	//client.Debug = true
	AlertMessageModel := models.CustomerAlertDefault
	placeOrderLog := models.Log_TvSiginalData{
		PlaceOrderType: tv.PlaceOrderType,
		CustomerID:     Customer.CustomerID,
		Time:           common.GetUtcTimeNow(),
		Simulation:     Customer.Simulation,
		Symbol:         Symbol,
	}

	var isTowWayPositionOnHand = false //是否雙向持倉
	var isCloseOrder = false           //是否為平倉

	//判斷是不是平倉
	if (tv.PlaceOrderType.Side == models.BuySideType && tv.PlaceOrderType.PositionSideType == models.ShortPositionSideType) ||
		(tv.PlaceOrderType.Side == models.SellSideType && tv.PlaceOrderType.PositionSideType == models.LongPositionSideType) {
		isCloseOrder = true
	}

	//查出目前持倉情況
	positions, err := client.GetUMPositionRiskService().Symbol(Symbol).Do(c)
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
		Symbol(Symbol).
		Quantity(placeAmount).
		Type(MarketOrder).
		Side(OrderSide(tv.PlaceOrderType.Side)).
		Do(c)

	//如果下單有問題，就return
	if err != nil {
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, err
	}
	log.Printf("Customer:%s %v order created: %+v", Customer.CustomerID, MarketOrder, order)

	placedOrder := &UMUserTradeResponse{Commission: "0", RealizedPnl: "0"}

	//取出下單結果要先暫停0.5秒，等待下單結果完成
	for i := 0; i < maxRetries; i++ {
		fmt.Println("waiting for 0.5 seconds... --> ", i)
		time.Sleep(waitTime)

		// 嘗試獲取資料
		history, err := client.GetUMUserTradeService().Symbol(Symbol).OrderId(order.OrderId).Do(c)
		if (err == nil) && (len(history) > 0) {
			placedOrder = history[0]
			break
		}
		if i == maxRetries-1 {
			placeOrderLog.Result = placeOrderLog.Result + "\nGet placed order failed:" + err.Error()
		}
	}

	//依下單結果補足資料
	//寫入訂單編號
	placeOrderLog.Result = strconv.FormatInt((*order).OrderId, 10)
	placeOrderLog.Amount = common.Decimal(order.OrigQty)
	placeOrderLog.Price = common.Decimal(placedOrder.Price) //order.AvgPrice在第一時間不會有值，所以要用placedOrder
	placeOrderLog.Fee = common.Decimal(placedOrder.Commission)
	placeOrderLog.Profit = common.Decimal(placedOrder.RealizedPnl)

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
	asset, err := Client.GetAccountBalanceService().Symbol("USDT").DoSingle(ctx)
	if err != nil {
		return 0, err
	}
	return common.Decimal(asset.Balance), nil
}

// 要改槓桿及改成雙向持倉
func (Client *Client) UpdateLeverage(ctx context.Context, symbol string, leverage int64) error {
	//改多空槓桿
	_, err := Client.GetLeverageService().
		Symbol(common.FormatSymbol(symbol, false)).
		Leverage(leverage).
		Do(ctx)
	if err != nil {
		return err
	}

	//修改成雙向持倉
	positionsetting, err := Client.GetPositionService().Do(ctx)

	if (err != nil) || (positionsetting == nil) {
		return err
	}

	//如果己是雙向持倉，不用改
	if !positionsetting.DualSidePosition {
		_, err = Client.GetPositionService().DualSidePosition("true").
			DoUpdate(ctx)
		if err != nil {
			return nil
		}
	}

	//取得全逐倉模式
	IsoOrCloseMode, err := Client.GetMarginTypeService().
		Symbol(symbol).Do(ctx)
	if err != nil {
		return nil
	}

	if IsoOrCloseMode.MarginType != string(MarginIsolated) {
		//修改全逐倉模式
		_, err = Client.GetMarginTypeService().
			Symbol(symbol).MarginType(MarginIsolated).Do(ctx)
		if err != nil {
			return nil
		}
	}

	return nil
}
