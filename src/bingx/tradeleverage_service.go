package bingx

import (
	"context"
	"encoding/json"
	"net/http"
)

type SetTradeService struct {
	c        *Client
	symbol   string
	side     PositionSideType
	leverage int64
}

type GetTradeService struct {
	c      *Client
	symbol string
}

func (s *SetTradeService) Symbol(symbol string) *SetTradeService {
	s.symbol = symbol
	return s
}

func (s *SetTradeService) PositionSide(side PositionSideType) *SetTradeService {
	s.side = side
	return s
}

func (s *SetTradeService) Leverage(leverage int64) *SetTradeService {
	s.leverage = leverage
	return s
}

func (s *GetTradeService) Symbol(symbol string) *GetTradeService {
	s.symbol = symbol
	return s
}

type LeverageType struct {
	Leverage            int64  `json:"leverage"`            //槓桿倍數
	Symbol              string `json:"symbol"`              //交易對
	AvailableLongVol    string `json:"availableLongVol"`    //可開多數量
	AvailableShortVol   string `json:"availableShortVol"`   //可開空數量
	AvailableLongVal    string `json:"availableLongVal"`    //可開多價值
	AvailableShortVal   string `json:"availableShortVal"`   //可開空價值
	MaxPositionLongVal  string `json:"maxPositionLongVal"`  //持倉最大可開多價值
	MaxPositionShortVal string `json:"maxPositionShortVal"` //持倉最大可開空價值
	LongLeverage        int64  `json:"longLeverage"`        //多倉槓桿倍數
	ShortLeverage       int64  `json:"shortLeverage"`       //空倉槓桿倍數
	MaxLongLeverage     int64  `json:"maxLongLeverage"`     //最大多倉槓桿倍數
	MaxShortLeverage    int64  `json:"maxShortLeverage"`    //最大空倉槓桿倍數
}

func (s *SetTradeService) Do(ctx context.Context, opts ...RequestOption) (res *LeverageType, err error) {
	r := &request{method: http.MethodPost, endpoint: "/openApi/swap/v2/trade/leverage"}

	if s.symbol != "" {
		r.addParam("symbol", s.symbol)
	}

	if s.side != "" {
		r.addParam("side", s.side)
	}

	if s.leverage > 0 {
		r.addParam("leverage", s.leverage)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *LeverageType `json:"data"`
	})

	err = json.Unmarshal(data, &resp)
	res = resp.Data

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *GetTradeService) Do(ctx context.Context, opts ...RequestOption) (res *LeverageType, err error) {
	r := &request{method: http.MethodGet, endpoint: "/openApi/swap/v2/trade/leverage"}

	if s.symbol != "" {
		r.addParam("symbol", s.symbol)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *LeverageType `json:"data"`
	})

	err = json.Unmarshal(data, &resp)
	res = resp.Data

	if err != nil {
		return nil, err
	}

	return res, nil
}
