package bitunix

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
	OrderId    string `json:"orderId"`
	UserId     string `json:"userId"`
	OrderType  string `json:"orderType"`
	Amount     string `json:"amount"`
	DealAmount string `json:"dealAmount"`
	Volume     string `json:"volume"` //這個有兩個。要查：Todo:
	LeftAmount string `json:"leftAmount"`
	DealVolume string `json:"dealVolume"`
	LeftVolume string `json:"leftVolume"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	Side       string `json:"side"`
	Price      string `json:"price"`
	AvgPrice   string `json:"avgPrice"`
	Progress   string `json:"progress"`
	Ctime      string `json:"ctime"`
	Utime      string `json:"utime"`
	Base       string `json:"base"`
	Quote      string `json:"quote"`
	Symbol     string `json:"symbol"`
	Fee        string `json:"fee"`
	FeeCoin    string `json:"feeCoin"`
}

func (s *GetOpenPositionsService) Do(ctx context.Context, opts ...RequestOption) (res *[]Position, err error) {
	r := &request{method: http.MethodPost, endpoint: "/api/spot/v1/order/pending/list"}

	if s.symbol != "" {
		r.addParam("symbol", s.symbol)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		bitunixAPIRawResponse
		Data *[]Position `json:"data"`
	})

	err = json.Unmarshal(data, &resp)
	res = resp.Data

	if err != nil {
		return nil, err
	}

	return res, nil
}
