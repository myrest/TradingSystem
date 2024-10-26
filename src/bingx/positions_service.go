package bingx

import (
	"context"
	"encoding/json"
	"net/http"
)

type GetOpenPositionsService struct {
	c      *Client
	symbol string
}

func (s *GetOpenPositionsService) Symbol(symbol string) *GetOpenPositionsService {
	s.symbol = symbol
	return s
}

type Position struct {
	Symbol             string  `json:"symbol"`             //交易對
	PositionId         string  `json:"positionId"`         //倉位ID
	PositionSide       string  `json:"positionSide"`       //倉位方向 LONG/SHORT 多/空
	Isolated           bool    `json:"isolated"`           //是否是逐倉模式, true:逐倉模式 false:全倉
	PositionAmt        string  `json:"positionAmt"`        //持倉數量
	AvailableAmt       string  `json:"availableAmt"`       //可平倉數量
	UnrealizedProfit   string  `json:"unrealizedProfit"`   //未實現盈虧
	RealisedProfit     string  `json:"realisedProfit"`     //已實現盈虧
	InitialMargin      string  `json:"initialMargin"`      //初始保證金
	AvgPrice           string  `json:"avgPrice"`           //開倉均價
	LiquidationPrice   float64 `json:"liquidationPrice"`   //強平價
	Leverage           int     `json:"leverage"`           //槓桿
	PositionValue      string  `json:"positionValue"`      //持有價值
	MarkPrice          string  `json:"markPrice"`          //標記價格
	RiskRate           string  `json:"riskRate"`           //風險率，風險率達到100%時會強制減倉或者平倉
	MaxMarginReduction string  `json:"maxMarginReduction"` //最大可減少保證金
	PnlRatio           string  `json:"pnlRatio"`           //未實現盈虧收益率
}

func (s *GetOpenPositionsService) Do(ctx context.Context, opts ...RequestOption) (res *[]Position, err error) {
	r := &request{method: http.MethodGet, endpoint: "/openApi/swap/v2/user/positions"}

	if s.symbol != "" {
		r.addParam("symbol", s.symbol)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data *[]Position `json:"data"`
	})

	err = json.Unmarshal(data, &resp)
	res = resp.Data

	if err != nil {
		return nil, err
	}

	return res, nil
}
