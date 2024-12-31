package bitunix_feature

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

const maxRetries = 3
const waitTime = 500 * time.Millisecond

func (Client *Client) GetBalance(ctx context.Context) (float64, error) {
	asset, err := Client.GetAccountBalanceService().Symbol("USDT").Do(ctx)
	if err != nil {
		return 0, err
	}
	return common.Decimal(asset.Data.Available), nil
}

// 要改槓桿及改成雙向持倉
func (Client *Client) UpdateLeverage(ctx context.Context, symbol string, leverage int64) error {
	symbol = common.FormatSymbol(symbol, false)
	//改多空槓桿
	_, err := Client.GetUpdateLeverageService().
		Symbol(symbol).
		Leverage(leverage).
		MarginCoin("USDT").
		Do(ctx)
	if err != nil {
		return err
	}

	//取得目前持倉模式
	asset, err := Client.GetAccountBalanceService().Symbol("USDT").Do(ctx)
	if err != nil {
		return err
	}

	if asset.Data.PositionMode != string(HEDGE) {
		//如果不是雙向持倉，需要改成雙向持倉
		_, err := Client.GetUpdatePositionService().DualSidePosition(HEDGE).Do(ctx)
		if err != nil {
			return err
		}
	}

	//取得目前全逐倉模式
	IsoOrCloseMode, err := Client.GetLeverageMarginTypeService().
		Symbol(symbol).
		MarginCoin("USDT").
		Do(ctx)
	if err != nil {
		return err
	}

	if IsoOrCloseMode.Data.MarginMode != string(MarginCrossed) {
		//修改全逐倉模式
		_, err = Client.GetUpdateMarginTypeService().
			Symbol(symbol).
			MarginType(MarginCrossed).
			MarginCoin("USDT").
			Do(ctx)
		if err != nil {
			return nil
		}
	}

	return nil
}

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
	openPositions, err := client.GetPendingPositionsService().Symbol(Symbol).Do(c)
	if err != nil {
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, err
	}

	//判斷是否為雙向持倉，己為雙向持倉不用再判斷
	if len(openPositions.Data) > 1 {
		isTowWayPositionOnHand = true
	}

	longPedingPosition := PedingPosition{}
	shortPedingPosition := PedingPosition{}
	currentPosition := PedingPosition{}

	for _, v := range openPositions.Data {
		if v.Side == string(Buy) {
			longPedingPosition = v
		} else if v.Side == string(Sell) {
			shortPedingPosition = v
		}
	}

	if tv.PlaceOrderType.PositionSideType == models.LongPositionSideType {
		currentPosition = longPedingPosition
	} else {
		currentPosition = shortPedingPosition
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

	if placeAmount == 0 {
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, fmt.Errorf("下單數量為 0。")
	}

	//如果平倉，則下單數量不得大於目前持倉數量
	if isCloseOrder && placeAmount > common.Decimal(currentPosition.Qty) {
		placeAmount = common.Decimal(currentPosition.Qty)
	}

	//下市價單
	shouldOpenOrClose := PositionOpen
	if isCloseOrder {
		shouldOpenOrClose = PositionClose
	}
	shouldBuyOrSell := Buy
	if tv.PlaceOrderType.PositionSideType == models.ShortPositionSideType {
		shouldBuyOrSell = Sell
	}

	bitunix_order_service := client.GetPlaceNewOrderService().
		Symbol(Symbol).
		TradeSide(shouldOpenOrClose).
		Quantity(placeAmount).
		Side(shouldBuyOrSell).
		OrderType(MarketOrder).
		Effect(GTC)

	if isCloseOrder {
		bitunix_order_service.PositionId(currentPosition.PositionId)
	}

	order, err := bitunix_order_service.Do(c)

	//如果下單有問題，就return
	if err != nil {
		return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, err
	}
	log.Printf("Customer:%s %v order created: %+v", Customer.CustomerID, MarketOrder, order)

	orderid := int64(common.Decimal(order.Data.OrderId))
	placedOrder := &OrderDetailResponse{}
	//取出下單結果要先暫停0.5秒，等待下單結果完成
	for i := 0; i < maxRetries; i++ {
		fmt.Println("waiting for 0.5 seconds... --> ", i)
		time.Sleep(waitTime)

		//取出下單結果
		placedOrder, err = client.GetOrderDetailService().
			OrderId(orderid).
			Do(c)

		if err != nil {
			return placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, err
		}

		if placedOrder.Data.Fee != "" {
			break
		}
		if i == maxRetries-1 {
			placeOrderLog.Result = placeOrderLog.Result + "\nGet placed order failed:"
		}
	}

	//依下單結果補足資料
	//寫入訂單編號
	placeOrderLog.Result = order.Data.OrderId
	placeOrderLog.Amount = common.Decimal(placedOrder.Data.Qty)
	placeOrderLog.Price = common.Decimal(placedOrder.Data.Price)
	placeOrderLog.Fee = common.Decimal(placedOrder.Data.Fee) * -1 //因為手續費值為正數，要變成負數
	placeOrderLog.Profit = common.Decimal(placedOrder.Data.RealizedPNL)

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
