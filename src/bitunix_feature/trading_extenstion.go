package bitunix_feature

import (
	"context"
	"encoding/json"
	"net/http"
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
type GetPositionService struct {
	c            *Client
	positionMode SidePositionMode
}

type UMPositionResponse struct {
	StandardResponse
	Data struct {
		PositionMode string `json:"positionMode"`
	} `json:"data"`
}

func (s *GetPositionService) DualSidePosition(positionMode SidePositionMode) *GetPositionService {
	s.positionMode = positionMode
	return s
}

func (s *GetPositionService) Do(ctx context.Context, opts ...RequestOption) (res *UMPositionResponse, err error) {
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
type GetLeverageService struct {
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

func (s *GetLeverageService) Leverage(leverage int64) *GetLeverageService {
	s.leverage = &leverage
	return s
}

func (s *GetLeverageService) MarginCoin(marginCoin string) *GetLeverageService {
	s.marginCoin = &marginCoin
	return s
}

func (s *GetLeverageService) Symbol(symbol string) *GetLeverageService {
	s.symbol = &symbol
	return s
}

func (s *GetLeverageService) Do(ctx context.Context, opts ...RequestOption) (res *LeverageResponse, err error) {
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

// region 可用餘額、持倉模式等等
type GetAccountBalanceService struct {
	c          *Client
	marginCoin *string
}

type AccountBalanceResponse struct {
	Available              string `json:"available"`
	Bonus                  string `json:"bonus"`
	CrossUnrealizedPNL     string `json:"crossUnrealizedPNL"`
	Frozen                 string `json:"frozen"`
	IsolationUnrealizedPNL string `json:"isolationUnrealizedPNL"`
	Margin                 string `json:"margin"`
	MarginCoin             string `json:"marginCoin"`
	PositionMode           string `json:"positionMode"`
	Transfer               string `json:"transfer"`
}

func (s *GetAccountBalanceService) Symbol(symbol string) *GetAccountBalanceService {
	s.marginCoin = &symbol
	return s
}

func (s *GetAccountBalanceService) Do(ctx context.Context, opts ...RequestOption) (res []*AccountBalanceResponse, err error) {
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
	res = make([]*AccountBalanceResponse, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//endregion
