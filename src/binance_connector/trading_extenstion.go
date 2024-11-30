package binance_connector

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// region 基本定義
type OrderSide string
type PositionSide string
type OrderType string
type SelfTradePreventionMode string
type MarginTradingType string
type TimeInForce string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"

	PositionBoth  PositionSide = "BOTH"
	PositionLong  PositionSide = "LONG"
	PositionShort PositionSide = "SHORT"

	None        SelfTradePreventionMode = "NONE"
	ExpireTaker SelfTradePreventionMode = "EXPIRE_TAKER"
	ExpireMaker SelfTradePreventionMode = "EXPIRE_MAKER"
	ExpireBoth  SelfTradePreventionMode = "EXPIRE_BOTH"

	MarginIsolated MarginTradingType = "ISOLATED" //全倉
	MarginCrossed  MarginTradingType = "CROSSED"  //逐倉

	GTC TimeInForce = "GTC" // Good Till Cancelled
	GTD TimeInForce = "GTD" // Good Till Date

	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"
)

type StandardResponse struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}

// endregion 基本定義

// region 持倉模式 "true": 双向持仓模式；"false": 单向持仓模式
// https://developers.binance.com/docs/zh-CN/derivatives/usds-margined-futures/account/rest-api/Get-Current-Position-Mode
type GetPositionService struct {
	c                *Client
	dualSidePosition string
}

type UMPositionResponse struct {
	//"true": 双向持仓模式；"false": 单向持仓模式
	DualSidePosition bool `json:"dualSidePosition"`
}

func (s *GetPositionService) Do(ctx context.Context, opts ...RequestOption) (res *UMPositionResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/positionSide/dual",
		secType:  secTypeSigned,
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &UMPositionResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//endregion

// region 更改持倉模式 "true": 双向持仓模式；"false": 单向持仓模式

func (s *GetPositionService) DualSidePosition(dualSidePosition string) *GetPositionService {
	s.dualSidePosition = dualSidePosition
	return s
}

func (s *GetPositionService) DoUpdate(ctx context.Context, opts ...RequestOption) (res *StandardResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/fapi/v1/positionSide/dual",
		secType:  secTypeSigned,
	}

	if s.dualSidePosition != "" {
		r.setParam("dualSidePosition", s.dualSidePosition)
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &StandardResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//endregion

// region 調整槓桿
type GetLeverageService struct {
	c        *Client
	symbol   *string
	leverage *int64 //target initial leverage: int from 1 to 125
}

type LeverageResponse struct {
	Leverage         int64  `json:"leverage"`
	MaxNotionalValue string `json:"maxNotionalValue"`
	Symbol           string `json:"symbol"`
}

func (s *GetLeverageService) Symbol(symbol string) *GetLeverageService {
	s.symbol = &symbol
	return s
}

func (s *GetLeverageService) Leverage(leverage int64) *GetLeverageService {
	s.leverage = &leverage
	return s
}

func (s *GetLeverageService) Do(ctx context.Context, opts ...RequestOption) (res *LeverageResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/fapi/v1/leverage",
		secType:  secTypeSigned,
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	if s.leverage != nil {
		r.setParam("leverage", *s.leverage)
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &LeverageResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// endregion

// region 修改全逐倉模式
type GetMarginTypeService struct {
	c          *Client
	symbol     *string
	marginType MarginTradingType
}

type MarginTypeResponse struct {
	Symbol           string `json:"symbol"`
	MarginType       string `json:"marginType"`
	IsAutoAddMargin  bool   `json:"isAutoAddMargin"`
	Leverage         int64  `json:"leverage"`
	MaxNotionalValue string `json:"maxNotionalValue"`
}

func (s *GetMarginTypeService) Symbol(symbol string) *GetMarginTypeService {
	s.symbol = &symbol
	return s
}

func (s *GetMarginTypeService) MarginType(marginType MarginTradingType) *GetMarginTypeService {
	s.marginType = marginType
	return s
}

func (s *GetMarginTypeService) DoUpdate(ctx context.Context, opts ...RequestOption) (res *MarginTypeResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/fapi/v1/marginType",
		secType:  secTypeSigned,
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}

	if s.marginType != "" {
		r.setParam("marginType", string(s.marginType))
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &MarginTypeResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *GetMarginTypeService) Do(ctx context.Context, opts ...RequestOption) (*MarginTypeResponse, error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/symbolConfig",
		secType:  secTypeSigned,
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res := make([]*MarginTypeResponse, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res[0], nil
	} else {
		return nil, errors.New("symbol not found")
	}
}

// endregion

// region 可用餘額
type GetAccountBalanceService struct {
	c      *Client
	symbol *string //自己加的功能，用來filter指定的幣種
}

type AccountBalanceResponse struct {
	AccountAlias       string `json:"accountAlias"`       // 账户唯一识别码
	Asset              string `json:"asset"`              // 资产
	Balance            string `json:"balance"`            // 总余额
	CrossWalletBalance string `json:"crossWalletBalance"` // 全仓余额
	CrossUnPnl         string `json:"crossUnPnl"`         // 全仓持仓未实现盈亏
	AvailableBalance   string `json:"availableBalance"`   // 下单可用余额
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`  // 最大可转出余额
	MarginAvailable    bool   `json:"marginAvailable"`    // 是否可用作联合保证金
	UpdateTime         int64  `json:"updateTime"`
}

func (s *GetAccountBalanceService) Do(ctx context.Context, opts ...RequestOption) (res []*AccountBalanceResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v3/balance",
		secType:  secTypeSigned,
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = make([]*AccountBalanceResponse, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *GetAccountBalanceService) Symbol(symbol string) *GetAccountBalanceService {
	s.symbol = &symbol
	return s
}

func (s *GetAccountBalanceService) DoSingle(ctx context.Context, opts ...RequestOption) (*AccountBalanceResponse, error) {
	result, err := s.Do(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		for _, v := range result {
			if v.Asset == "USDT" {
				return v, nil
			}
		}
	}
	return nil, errors.New("asset not found")
}

//endregion

// region 下單
type GetUMNewOrderService struct {
	c                *Client
	symbol           *string
	side             *OrderSide
	positionSide     PositionSide
	ordertype        OrderType
	quantity         float64
	reduceOnly       *bool
	price            float64
	newClientOrderId string
	timeInForce      TimeInForce
}

type UMNewOrderResponse struct {
	ClientOrderId           string `json:"clientOrderId"`
	CumQty                  string `json:"cumQty"`
	CumQuote                string `json:"cumQuote"`
	ExecutedQty             string `json:"executedQty"`
	OrderId                 int64  `json:"orderId"`
	AvgPrice                string `json:"avgPrice"`
	OrigQty                 string `json:"origQty"`
	Price                   string `json:"price"`
	ReduceOnly              bool   `json:"reduceOnly"`
	Side                    string `json:"side"`
	PositionSide            string `json:"positionSide"`
	Status                  string `json:"status"`
	StopPrice               string `json:"stopPrice"`
	ClosePosition           bool   `json:"closePosition"`
	Symbol                  string `json:"symbol"`
	TimeInForce             string `json:"timeInForce"`
	Type                    string `json:"type"`
	OrigType                string `json:"origType"`
	ActivatePrice           string `json:"activatePrice"`
	PriceRate               string `json:"priceRate"`
	WorkingType             string `json:"workingType"`
	SelfTradePreventionMode string `json:"selfTradePreventionMode"`
	PriceProtect            bool   `json:"priceProtect"`
	GoodTillDate            int64  `json:"goodTillDate"`
	UpdateTime              int64  `json:"updateTime"`
	PriceMatch              string `json:"priceMatch"`
}

func (s *GetUMNewOrderService) Symbol(symbol string) *GetUMNewOrderService {
	s.symbol = &symbol
	return s
}
func (s *GetUMNewOrderService) Side(side OrderSide) *GetUMNewOrderService {
	s.side = &side
	return s
}

func (s *GetUMNewOrderService) PositionSide(positionSide PositionSide) *GetUMNewOrderService {
	s.positionSide = positionSide
	return s
}

func (s *GetUMNewOrderService) Type(ordertype OrderType) *GetUMNewOrderService {
	s.ordertype = ordertype
	return s
}

func (s *GetUMNewOrderService) Quantity(quantity float64) *GetUMNewOrderService {
	s.quantity = quantity
	return s
}

func (s *GetUMNewOrderService) ReduceOnly(reduceOnly bool) *GetUMNewOrderService {
	s.reduceOnly = &reduceOnly
	return s
}
func (s *GetUMNewOrderService) Price(price float64) *GetUMNewOrderService {
	s.price = price
	return s
}

func (s *GetUMNewOrderService) NewClientOrderId(newClientOrderId string) *GetUMNewOrderService {
	s.newClientOrderId = newClientOrderId
	return s
}

func (s *GetUMNewOrderService) TimeInForce(timeInForce TimeInForce) *GetUMNewOrderService {
	s.timeInForce = timeInForce
	return s
}

func (s *GetUMNewOrderService) Do(ctx context.Context, opts ...RequestOption) (res *UMNewOrderResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/fapi/v1/order",
		secType:  secTypeSigned,
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	if s.side != nil {
		r.setParam("side", *s.side)
	}
	if s.positionSide != "" {
		r.setParam("positionSide", s.positionSide)
	}
	if s.ordertype != "" {
		r.setParam("type", s.ordertype) //在document裏Key是type
	}
	if s.quantity != 0 {
		r.setParam("quantity", s.quantity)
	}
	if s.price != 0 {
		r.setParam("price", s.price)
	}
	if s.newClientOrderId != "" {
		r.setParam("newClientOrderId", s.newClientOrderId)
	}
	if s.reduceOnly != nil {
		r.setParam("reduceOnly", *s.reduceOnly)
	}
	if s.timeInForce != "" {
		r.setParam("timeInForce", s.timeInForce)
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &UMNewOrderResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// endregion

// region 依訂單編號取得訂單資料
// https://developers.binance.com/docs/derivatives/portfolio-margin/trade/Query-UM-Order
type GetUMOrderService struct {
	c       *Client
	symbol  *string
	orderid *int64
}

type UMOrderResponse struct {
	UMNewOrderResponse
}

func (s *GetUMOrderService) Symbol(symbol string) *GetUMOrderService {
	s.symbol = &symbol
	return s
}

func (s *GetUMOrderService) OrderId(orderid int64) *GetUMOrderService {
	s.orderid = &orderid
	return s
}

func (s *GetUMOrderService) Do(ctx context.Context, opts ...RequestOption) (res *UMOrderResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/order",
		secType:  secTypeSigned,
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	if s.orderid != nil {
		r.setParam("orderId", *s.orderid)
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &UMOrderResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// endregion

// region 取得目前持倉風險，用來判斷是否雙向持仓
type GetUMPositionRiskService struct {
	c      *Client
	symbol string
}

type UMPositionRiskResponse struct {
	EntryPrice       string `json:"entryPrice"`       // 开仓均价
	Leverage         string `json:"leverage"`         // 当前杠杆倍数
	MarkPrice        string `json:"markPrice"`        // 当前标记价格
	MaxNotionalValue string `json:"maxNotionalValue"` // 当前杠杆倍数允许的名义价值上限
	PositionAmt      string `json:"positionAmt"`      // 头寸数量，符号代表多空方向
	Notional         string `json:"notional"`         // 名义价值
	Symbol           string `json:"symbol"`           // 交易对
	UnRealizedProfit string `json:"unRealizedProfit"` // 持仓未实现盈亏
	LiquidationPrice string `json:"liquidationPrice"` // 清算价格
	PositionSide     string `json:"positionSide"`     // 持仓方向
	UpdateTime       int64  `json:"updateTime"`       // 更新时间
}

func (s *GetUMPositionRiskService) Symbol(symbol string) *GetUMPositionRiskService {
	s.symbol = symbol
	return s
}

func (s *GetUMPositionRiskService) Do(ctx context.Context, opts ...RequestOption) (res []*UMPositionRiskResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v3/positionRisk",
		secType:  secTypeSigned,
	}
	if s.symbol != "" {
		r.setParam("symbol", s.symbol)
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = make([]*UMPositionRiskResponse, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// endregion

// region 取得目前訂單成交盈虧資料
type GetUMUserTradeService struct {
	c         *Client
	symbol    string
	orderId   int64
	startTime int64
	endTime   int64
	limit     int64
}

type UMUserTradeResponse struct {
	Symbol          string `json:"symbol"`
	ID              int64  `json:"id"`
	OrderID         int64  `json:"orderId"`
	Side            string `json:"side"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	RealizedPnl     string `json:"realizedPnl"`
	QuoteQty        string `json:"quoteQty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            int64  `json:"time"`
	Buyer           bool   `json:"buyer"`
	Maker           bool   `json:"maker"`
	PositionSide    string `json:"positionSide"`
}

func (s *GetUMUserTradeService) Symbol(symbol string) *GetUMUserTradeService {
	s.symbol = symbol
	return s
}

func (s *GetUMUserTradeService) OrderId(orderId int64) *GetUMUserTradeService {
	s.orderId = orderId
	return s
}

func (s *GetUMUserTradeService) StartTime(startTime int64) *GetUMUserTradeService {
	s.startTime = startTime
	return s
}

func (s *GetUMUserTradeService) EndTime(endTime int64) *GetUMUserTradeService {
	s.endTime = endTime
	return s
}

func (s *GetUMUserTradeService) Limit(limit int64) *GetUMUserTradeService {
	s.limit = limit
	return s
}

func (s *GetUMUserTradeService) Do(ctx context.Context, opts ...RequestOption) (res []*UMUserTradeResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/userTrades",
		secType:  secTypeSigned,
	}
	if s.symbol != "" {
		r.setParam("symbol", s.symbol)
	}
	if s.orderId != 0 {
		r.setParam("orderId", s.orderId)
	}
	if s.startTime != 0 {
		r.setParam("startTime", s.startTime)
	}
	if s.endTime != 0 {
		r.setParam("endTime", s.endTime)
	}
	if s.limit != 0 {
		r.setParam("limit", s.limit)
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = make([]*UMUserTradeResponse, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// endregion
