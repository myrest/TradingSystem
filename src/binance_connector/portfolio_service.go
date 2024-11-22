package binance_connector

import (
	"context"
	"encoding/json"
	"net/http"
)

// region 基本定義

type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

type PositionSide string

const (
	PositionBoth  PositionSide = "BOTH"
	PositionLong  PositionSide = "LONG"
	PositionShort PositionSide = "SHORT"
)

type OrderType string

const (
	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"
)

type TimeInForce string

const (
	GTC TimeInForce = "GTC" // Good Till Cancelled
	GTD TimeInForce = "GTD" // Good Till Date
)

type SelfTradePreventionMode string

const (
	None        SelfTradePreventionMode = "NONE"
	ExpireTaker SelfTradePreventionMode = "EXPIRE_TAKER"
	ExpireMaker SelfTradePreventionMode = "EXPIRE_MAKER"
	ExpireBoth  SelfTradePreventionMode = "EXPIRE_BOTH"
)

type UMStandardResponse struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}

// endregion 基本定義

// region 取得帳戶資產
// https://developers.binance.com/docs/derivatives/portfolio-margin/account/Get-UM-Account-Detail
type GetUMAccountAssetService struct {
	c *Client
}

// Asset 表示單個資產的結構
type Asset struct {
	Asset                  string `json:"asset"`                  // 資產名稱
	CrossWalletBalance     string `json:"crossWalletBalance"`     // 錢包餘額
	CrossUnPnl             string `json:"crossUnPnl"`             // 未實現盈虧
	MaintMargin            string `json:"maintMargin"`            // 所需維護保證金
	InitialMargin          string `json:"initialMargin"`          // 當前市價所需的總初始保證金
	PositionInitialMargin  string `json:"positionInitialMargin"`  // 當前市價所需的頭寸初始保證金
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"` // 當前市價所需的未平倉訂單初始保證金
	UpdateTime             int64  `json:"updateTime"`             // 最後更新時間
}

// Position 表示單個交易品種的頭寸
type Position struct {
	Symbol                 string `json:"symbol"`                 // 交易品種名稱
	InitialMargin          string `json:"initialMargin"`          // 當前市價所需的初始保證金
	MaintMargin            string `json:"maintMargin"`            // 所需維護保證金
	UnrealizedProfit       string `json:"unrealizedProfit"`       // 未實現盈虧
	PositionInitialMargin  string `json:"positionInitialMargin"`  // 當前市價所需的頭寸初始保證金
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"` // 當前市價所需的未平倉訂單初始保證金
	Leverage               string `json:"leverage"`               // 當前初始杠桿
	EntryPrice             string `json:"entryPrice"`             // 平均入場價格
	MaxNotional            string `json:"maxNotional"`            // 當前杠桿下的最大可用名義
	BidNotional            string `json:"bidNotional"`            // 買入名義（忽略）
	AskNotional            string `json:"askNotional"`            // 賣出名義（忽略）
	PositionSide           string `json:"positionSide"`           // 頭寸方向
	PositionAmt            string `json:"positionAmt"`            // 頭寸數量
	UpdateTime             int64  `json:"updateTime"`             // 最後更新時間
}

// AccountAssetResponse 表示帳戶資產響應的結構
type AccountAssetResponse struct {
	Assets    []Asset    `json:"assets"`    // 資產列表
	Positions []Position `json:"positions"` // 頭寸列表
}

func (s *GetUMAccountAssetService) Do(ctx context.Context, opts ...RequestOption) (res *AccountAssetResponse, err error) {
	s.c.Debug = true
	r := &request{
		method:   http.MethodGet,
		endpoint: "/papi/v1/um/account",
		secType:  secTypeSigned,
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &AccountAssetResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//endregion

// region 持倉模式 "true": 双向持仓模式；"false": 单向持仓模式
// https://developers.binance.com/docs/zh-CN/derivatives/portfolio-margin/account/Get-UM-Current-Position-Mode
type GetUMPositionService struct {
	c                *Client
	dualSidePosition string
}

//DualSidePosition bool `json:"dualSidePosition,string"` // 字串轉 bool

type UMPositionResponse struct {
	//"true": 双向持仓模式；"false": 单向持仓模式
	DualSidePosition bool `json:"dualSidePosition"`
}

func (s *GetUMPositionService) Do(ctx context.Context, opts ...RequestOption) (res *UMPositionResponse, err error) {
	s.c.Debug = true
	r := &request{
		method:   http.MethodGet,
		endpoint: "/papi/v1/um/positionSide/dual",
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
// https://developers.binance.com/docs/derivatives/portfolio-margin/account/Change-UM-Position-Mode

func (s *GetUMPositionService) DualSidePosition(dualSidePosition string) *GetUMPositionService {
	s.dualSidePosition = dualSidePosition
	return s
}

func (s *GetUMPositionService) DoUpdate(ctx context.Context, opts ...RequestOption) (res *UMStandardResponse, err error) {
	s.c.Debug = true
	r := &request{
		method:   http.MethodPost,
		endpoint: "/papi/v1/um/positionSide/dual",
		secType:  secTypeSigned,
	}

	if s.dualSidePosition != "" {
		r.setParam("dualSidePosition", s.dualSidePosition)
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &UMStandardResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//endregion

// region 可用餘額
// https://developers.binance.com/docs/derivatives/portfolio-margin/account
type GetUMAccountBalanceService struct {
	c     *Client
	asset string
}

type UMAccountBalanceResponse struct {
	Asset               string `json:"asset"`               // 資產名稱
	TotalWalletBalance  string `json:"totalWalletBalance"`  // 錢包餘額
	CrossMarginAsset    string `json:"crossMarginAsset"`    // 跨保證金資產
	CrossMarginBorrowed string `json:"crossMarginBorrowed"` // 跨保證金借入的本金
	CrossMarginFree     string `json:"crossMarginFree"`     // 跨保證金的自由資產
	CrossMarginInterest string `json:"crossMarginInterest"` // 跨保證金的利息
	CrossMarginLocked   string `json:"crossMarginLocked"`   // 跨保證金鎖定的資產
	UMWalletBalance     string `json:"umWalletBalance"`     // UM 錢包餘額
	UMUnrealizedPNL     string `json:"umUnrealizedPNL"`     // UM 的未實現利潤
	CMWalletBalance     string `json:"cmWalletBalance"`     // CM 錢包餘額
	CMUnrealizedPNL     string `json:"cmUnrealizedPNL"`     // CM 的未實現利潤
	UpdateTime          int64  `json:"updateTime"`          // 更新時間
	NegativeBalance     string `json:"negativeBalance"`     // 負餘額
}

func (s *GetUMAccountBalanceService) Asset(asset string) *GetUMAccountBalanceService {
	s.asset = asset
	return s
}

func (s *GetUMAccountBalanceService) Do(ctx context.Context, opts ...RequestOption) (res []*UMAccountBalanceResponse, err error) {
	s.c.Debug = true
	r := &request{
		method:   http.MethodGet,
		endpoint: "/papi/v1/balance",
		//endpoint: "/api/v3/account",
		secType: secTypeSigned,
	}

	if s.asset != "" {
		r.setParam("asset", s.asset)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = make([]*UMAccountBalanceResponse, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//endregion

// region 下單
// https://developers.binance.com/docs/derivatives/portfolio-margin/trade
// 目前會遇到 Order's position side does not match user's setting.的錯誤
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

type NewUMOrderResponse struct {
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
	Symbol                  string `json:"symbol"`
	TimeInForce             string `json:"timeInForce"`
	Type                    string `json:"type"`
	SelfTradePreventionMode string `json:"selfTradePreventionMode"`
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

func (s *GetUMNewOrderService) Do(ctx context.Context, opts ...RequestOption) (res []*NewUMOrderResponse, err error) {
	s.c.Debug = true
	r := &request{
		method:   http.MethodPost,
		endpoint: "/papi/v1/um/order",
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
	res = make([]*NewUMOrderResponse, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// endregion
