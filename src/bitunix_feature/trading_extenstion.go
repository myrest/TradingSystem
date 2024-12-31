package bitunix_feature

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// region 基本定義
type OrderSide string
type PositionSide string
type OrderType string
type SelfTradePreventionMode string
type MarginTradingType string
type TimeInForce string
type SidePositionMode string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"

	PositionOpen  PositionSide = "OPEN"
	PositionClose PositionSide = "CLOSE"

	None        SelfTradePreventionMode = "NONE"
	ExpireTaker SelfTradePreventionMode = "EXPIRE_TAKER"
	ExpireMaker SelfTradePreventionMode = "EXPIRE_MAKER"
	ExpireBoth  SelfTradePreventionMode = "EXPIRE_BOTH"

	MarginIsolated MarginTradingType = "ISOLATION" //逐倉
	MarginCrossed  MarginTradingType = "CROSS"     //全倉

	GTC       TimeInForce = "GTC" // Good Till Cancelled
	IOC       TimeInForce = "IOC"
	FOK       TimeInForce = "FOK"
	POST_ONLY TimeInForce = "POST_ONLY"

	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"

	OneWay SidePositionMode = "ONE_WAY"
	HEDGE  SidePositionMode = "HEDGE"
)

type StandardResponse struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}

// endregion 基本定義

// region 修改持倉模式-- "HEDGE": 双向持仓模式；"OneWay": 单向持仓模式
// https://openapidoc.bitunix.com/doc/account/change_position_mode.html
type GetUpdatePositionService struct {
	c            *Client
	positionMode SidePositionMode
}

type UMPositionResponse struct {
	StandardResponse
	Data struct {
		PositionMode string `json:"positionMode"`
	} `json:"data"`
}

func (s *GetUpdatePositionService) DualSidePosition(positionMode SidePositionMode) *GetUpdatePositionService {
	s.positionMode = positionMode
	return s
}

func (s *GetUpdatePositionService) Do(ctx context.Context, opts ...RequestOption) (res *UMPositionResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/api/v1/futures/account/change_position_mode",
	}

	if s.positionMode != "" {
		r.addParam("positionMode", s.positionMode)
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

// region 調整槓桿
type GetUpdateLeverageService struct {
	c          *Client
	marginCoin *string
	symbol     *string
	leverage   *int64 //target initial leverage: int from 1 to 125
}

type LeverageResponse struct {
	StandardResponse
	Data struct {
		Leverage   int64  `json:"leverage"`
		MarginCoin string `json:"marginCoin"`
		Symbol     string `json:"symbol"`
	} `json:"data"`
}

func (s *GetUpdateLeverageService) Leverage(leverage int64) *GetUpdateLeverageService {
	s.leverage = &leverage
	return s
}

func (s *GetUpdateLeverageService) MarginCoin(marginCoin string) *GetUpdateLeverageService {
	s.marginCoin = &marginCoin
	return s
}

func (s *GetUpdateLeverageService) Symbol(symbol string) *GetUpdateLeverageService {
	s.symbol = &symbol
	return s
}

func (s *GetUpdateLeverageService) Do(ctx context.Context, opts ...RequestOption) (res *LeverageResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/api/v1/futures/account/change_leverage",
	}
	if s.symbol != nil {
		r.addParam("symbol", *s.symbol)
	}
	if s.leverage != nil {
		r.addParam("leverage", *s.leverage)
	}

	if s.marginCoin != nil {
		r.addParam("marginCoin", *s.marginCoin)
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
type GetUpdateMarginTypeService struct {
	c          *Client
	symbol     *string
	marginMode MarginTradingType
	marginCoin *string
}

type MarginTypeResponse struct {
	StandardResponse
	Data struct {
		Symbol     string `json:"symbol"`
		MarginMode string `json:"marginMode"`
		MarginCoin string `json:"marginCoin"`
	} `json:"data"`
}

func (s *GetUpdateMarginTypeService) Symbol(symbol string) *GetUpdateMarginTypeService {
	s.symbol = &symbol
	return s
}

func (s *GetUpdateMarginTypeService) MarginType(marginType MarginTradingType) *GetUpdateMarginTypeService {
	s.marginMode = marginType
	return s
}

func (s *GetUpdateMarginTypeService) MarginCoin(marginCoin string) *GetUpdateMarginTypeService {
	s.marginCoin = &marginCoin
	return s
}

func (s *GetUpdateMarginTypeService) Do(ctx context.Context, opts ...RequestOption) (res *MarginTypeResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/api/v1/futures/account/change_margin_mode",
	}
	if s.symbol != nil {
		r.addParam("symbol", *s.symbol)
	}

	if s.marginMode != "" {
		r.addParam("marginMode", string(s.marginMode))
	}

	if s.marginCoin != nil {
		r.addParam("marginCoin", *s.marginCoin)
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

// endregion

// region 取得槓桿及全逐倉模式
type GetLeverageMarginTypeService struct {
	c          *Client
	symbol     *string
	marginCoin *string
}

type LeverageMarginTypeResponse struct {
	StandardResponse
	Data struct {
		Symbol     string `json:"symbol"`
		MarginMode string `json:"marginMode"`
		MarginCoin string `json:"marginCoin"`
		Leverage   int64  `json:"leverage"`
	} `json:"data"`
}

func (s *GetLeverageMarginTypeService) Symbol(symbol string) *GetLeverageMarginTypeService {
	s.symbol = &symbol
	return s
}

func (s *GetLeverageMarginTypeService) MarginCoin(marginCoin string) *GetLeverageMarginTypeService {
	s.marginCoin = &marginCoin
	return s
}

func (s *GetLeverageMarginTypeService) Do(ctx context.Context, opts ...RequestOption) (res *LeverageMarginTypeResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/api/v1/futures/account/get_leverage_margin_mode",
	}
	if s.symbol != nil {
		r.addParam("symbol", *s.symbol)
	}

	if s.marginCoin != nil {
		r.addParam("marginCoin", *s.marginCoin)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &LeverageMarginTypeResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// endregion

// region 可用餘額、持倉模式等等
// https://openapidoc.bitunix.com/doc/account/get_single_account.html
type GetAccountBalanceService struct {
	c          *Client
	marginCoin *string
}

type AccountBalanceResponse struct {
	StandardResponse
	Data struct {
		Available              string `json:"available"`
		Bonus                  string `json:"bonus"`
		CrossUnrealizedPNL     string `json:"crossUnrealizedPNL"`
		Frozen                 string `json:"frozen"`
		IsolationUnrealizedPNL string `json:"isolationUnrealizedPNL"`
		Margin                 string `json:"margin"`
		MarginCoin             string `json:"marginCoin"`
		PositionMode           string `json:"positionMode"`
		Transfer               string `json:"transfer"`
	} `json:"data"`
}

func (s *GetAccountBalanceService) Symbol(symbol string) *GetAccountBalanceService {
	s.marginCoin = &symbol
	return s
}

func (s *GetAccountBalanceService) Do(ctx context.Context, opts ...RequestOption) (res *AccountBalanceResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/api/v1/futures/account",
	}

	if s.marginCoin != nil {
		r.addParam("marginCoin", *s.marginCoin)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &AccountBalanceResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//endregion

// region 下單
type GetPlaceNewOrderService struct {
	c            *Client
	symbol       *string      // Trading pair
	qty          *string      // Amount (base coin)
	price        *string      // Price of the order, optional
	side         *OrderSide   // BUY or SELL
	tradeSide    PositionSide // Open or Close
	positionId   *string      // Only required when "tradeSide" is "CLOSE" Todo:還不知道用法
	orderType    OrderType    // LIMIT or MARKET
	effect       *TimeInForce // Required if the orderType is limit, IOC or FOK or GTC or POST_ONLY
	clientId     *string      // Customize order ID, optional
	reduceOnly   *bool        // Whether or not to just reduce the position, optional
	tpPrice      *string      // Take profit trigger price, optional
	tpStopType   *string      // Take profit trigger type, optional
	tpOrderType  *string      // Take profit trigger place order type, optional
	tpOrderPrice *string      // Take profit trigger place order price, optional
	slPrice      *string      // Stop loss trigger price, optional
	slStopType   *string      // Stop loss trigger type, optional
	slOrderType  *string      // Stop loss trigger place order type, optional
	slOrderPrice *string      // Stop loss trigger place order price, optional
}

// UMNewOrderResponse 是新的訂單響應結構
type UMNewOrderResponse struct {
	StandardResponse
	Data struct {
		OrderId  string // Trading pair
		ClientId string // Amount (base coin)
	} `json:"data"`
}

func (s *GetPlaceNewOrderService) Symbol(symbol string) *GetPlaceNewOrderService {
	s.symbol = &symbol
	return s
}

func (s *GetPlaceNewOrderService) Side(side OrderSide) *GetPlaceNewOrderService {
	s.side = &side
	return s
}

func (s *GetPlaceNewOrderService) Quantity(quantity float64) *GetPlaceNewOrderService {
	qty := strconv.FormatFloat(quantity, 'f', -1, 64)
	s.qty = &qty
	return s
}

func (s *GetPlaceNewOrderService) Price(price float64) *GetPlaceNewOrderService {
	p := strconv.FormatFloat(price, 'f', -1, 64)
	s.price = &p
	return s
}

func (s *GetPlaceNewOrderService) ReduceOnly(reduceOnly bool) *GetPlaceNewOrderService {
	s.reduceOnly = &reduceOnly
	return s
}

func (s *GetPlaceNewOrderService) TradeSide(tradeSide PositionSide) *GetPlaceNewOrderService {
	s.tradeSide = tradeSide
	return s
}

func (s *GetPlaceNewOrderService) PositionId(positionId string) *GetPlaceNewOrderService {
	s.positionId = &positionId
	return s
}

func (s *GetPlaceNewOrderService) OrderType(orderType OrderType) *GetPlaceNewOrderService {
	s.orderType = orderType
	return s
}

func (s *GetPlaceNewOrderService) Effect(effect TimeInForce) *GetPlaceNewOrderService {
	s.effect = &effect
	return s
}

func (s *GetPlaceNewOrderService) ClientId(clientId string) *GetPlaceNewOrderService {
	s.clientId = &clientId
	return s
}

func (s *GetPlaceNewOrderService) TpPrice(tpPrice float64) *GetPlaceNewOrderService {
	p := strconv.FormatFloat(tpPrice, 'f', -1, 64)
	s.tpPrice = &p
	return s
}

func (s *GetPlaceNewOrderService) TpStopType(tpStopType string) *GetPlaceNewOrderService {
	s.tpStopType = &tpStopType
	return s
}

func (s *GetPlaceNewOrderService) TpOrderType(tpOrderType string) *GetPlaceNewOrderService {
	s.tpOrderType = &tpOrderType
	return s
}

func (s *GetPlaceNewOrderService) TpOrderPrice(tpOrderPrice float64) *GetPlaceNewOrderService {
	p := strconv.FormatFloat(tpOrderPrice, 'f', -1, 64)
	s.tpOrderPrice = &p
	return s
}

func (s *GetPlaceNewOrderService) SlPrice(slPrice float64) *GetPlaceNewOrderService {
	p := strconv.FormatFloat(slPrice, 'f', -1, 64)
	s.slPrice = &p
	return s
}

func (s *GetPlaceNewOrderService) SlStopType(slStopType string) *GetPlaceNewOrderService {
	s.slStopType = &slStopType
	return s
}

func (s *GetPlaceNewOrderService) SlOrderType(slOrderType string) *GetPlaceNewOrderService {
	s.slOrderType = &slOrderType
	return s
}

func (s *GetPlaceNewOrderService) Do(ctx context.Context, opts ...RequestOption) (res *UMNewOrderResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/api/v1/futures/trade/place_order",
	}

	r.addParam("qty", *s.qty)
	r.addParam("symbol", *s.symbol)
	r.addParam("side", *s.side)
	r.addParam("tradeSide", s.tradeSide)
	r.addParam("orderType", s.orderType)

	if s.price != nil {
		r.addParam("price", *s.price)
	}

	if s.reduceOnly != nil {
		r.addParam("reduceOnly", *s.reduceOnly)
	}

	if s.positionId != nil {
		r.addParam("positionId", *s.positionId)
	}

	if s.effect != nil {
		r.addParam("effect", *s.effect)
	}

	if s.clientId != nil {
		r.addParam("clientId", *s.clientId)
	}

	if s.tpPrice != nil {
		r.addParam("tpPrice", *s.tpPrice)
	}

	if s.tpStopType != nil {
		r.addParam("tpStopType", *s.tpStopType)
	}

	if s.tpOrderType != nil {
		r.addParam("tpOrderType", *s.tpOrderType)
	}

	if s.tpOrderPrice != nil {
		r.addParam("tpOrderPrice", *s.tpOrderPrice)
	}

	if s.slPrice != nil {
		r.addParam("slPrice", *s.slPrice)
	}

	if s.slStopType != nil {
		r.addParam("slStopType", *s.slStopType)
	}

	if s.slOrderType != nil {
		r.addParam("slOrderType", *s.slOrderType)
	}

	if s.slOrderPrice != nil {
		r.addParam("slOrderPrice", *s.slOrderPrice)
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
	if res.Code != 0 {
		return nil, errors.New(res.Msg)
	}
	return res, nil
}

// endregion

// region 取得Position ID
// https://openapidoc.bitunix.com/doc/position/get_pending_positions.html
type GetPendingPositionsService struct {
	c      *Client
	symbol *string
}

type PedingPosition struct {
	PositionId    string `json:"positionId"`
	Symbol        string `json:"symbol"`
	Qty           string `json:"qty"`
	EntryValue    string `json:"entryValue"`
	Side          string `json:"side"`
	MarginMode    string `json:"marginMode"`
	PositionMode  string `json:"positionMode"`
	Leverage      int64  `json:"leverage"`
	Fee           string `json:"fee"`
	Funding       string `json:"funding"`
	RealizedPNL   string `json:"realizedPNL"`
	Margin        string `json:"margin"`
	UnrealizedPNL string `json:"unrealizedPNL"`
	LiqPrice      string `json:"liqPrice"`
	MarginRate    string `json:"marginRate"`
	AvgOpenPrice  string `json:"avgOpenPrice"`
	Ctime         string `json:"ctime"`
	Mtime         string `json:"mtime"`
}

type PendingPositionsResponse struct {
	StandardResponse
	Data []PedingPosition `json:"data"`
}

func (s *GetPendingPositionsService) Symbol(symbol string) *GetPendingPositionsService {
	s.symbol = &symbol
	return s
}

func (s *GetPendingPositionsService) Do(ctx context.Context, opts ...RequestOption) (res *PendingPositionsResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/api/v1/futures/position/get_pending_positions",
	}
	if s.symbol != nil {
		r.addParam("symbol", *s.symbol)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &PendingPositionsResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// endregion

// region 取得訂單資料
// https://openapidoc.bitunix.com/doc/trade/get_order_detail.html
type GetOrderDetailService struct {
	c        *Client
	orderId  int64
	clientId int64
}

type OrderDetailResponse struct {
	StandardResponse
	Data struct {
		OrderId      string `json:"orderId"`
		Symbol       string `json:"symbol"`
		Qty          string `json:"qty"`
		TradeQty     string `json:"tradeQty"`
		PositionMode string `json:"positionMode"`
		MarginMode   string `json:"marginMode"`
		Leverage     int64  `json:"leverage"`
		Price        string `json:"price"`
		Side         string `json:"side"`
		OrderType    string `json:"orderType"`
		Effect       string `json:"effect"`
		ClientId     string `json:"clientId"`
		ReduceOnly   bool   `json:"reduceOnly"`
		Status       string `json:"status"`
		Fee          string `json:"fee"`
		RealizedPNL  string `json:"realizedPNL"`
		TpPrice      string `json:"tpPrice"`
		TpStopType   string `json:"tpStopType"`
		TpOrderType  string `json:"tpOrderType"`
		TpOrderPrice string `json:"tpOrderPrice"`
		SlPrice      string `json:"slPrice"`
		SlStopType   string `json:"slStopType"`
		SlOrderType  string `json:"slOrderType"`
		SlOrderPrice string `json:"slOrderPrice"`
		Ctime        string `json:"ctime"`
		Mtime        string `json:"mtime"`
	} `json:"data"`
}

func (s *GetOrderDetailService) OrderId(orderId int64) *GetOrderDetailService {
	s.orderId = orderId
	return s
}

func (s *GetOrderDetailService) ClientId(clientId int64) *GetOrderDetailService {
	s.clientId = clientId
	return s
}

func (s *GetOrderDetailService) Do(ctx context.Context, opts ...RequestOption) (res *OrderDetailResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/api/v1/futures/trade/get_order_detail",
	}

	if s.orderId != 0 {
		r.addParam("orderId", s.orderId)
	}
	if s.clientId != 0 {
		r.addParam("clientId", s.clientId)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = &OrderDetailResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// endregion
