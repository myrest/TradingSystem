package bitunix

import (
	"context"
	"encoding/json"
	"net/http"
)

type GetTradingPairService struct {
	c *Client
}

type TradingPair struct {
	Symbol         string `json:"symbol"` //小寫 btcusdt
	Id             int    `json:"userId"`
	Base           string `json:"base"`  //BTC
	Quote          string `json:"quote"` //USDT
	BasePrecision  int    `json:"basePrecision"`
	QuotePrecision int    `json:"quotePrecision"`
	MinPrice       string `json:"minPrice"`
	MinVolume      string `json:"minVolume"`
	IsOpen         int    `json:"isOpen"`
	IsHot          int    `json:"isHot"`
	IsRecommend    int    `json:"isRecommend"`
	IsShow         int    `json:"isShow"`
	TradeArea      string `json:"tradeArea"` //USDT
	Sort           int    `json:"sort"`
	OpenTime       string `json:"openTime"` //Null
}

func (s *GetTradingPairService) Do(ctx context.Context, opts ...RequestOption) (res *[]TradingPair, err error) {
	r := &request{method: http.MethodGet, endpoint: "/api/spot/v1/common/coin_pair/list"}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		Code string         `json:"code"`
		Msg  string         `json:"msg"`
		Data *[]TradingPair `json:"data"`
	})

	err = json.Unmarshal(data, resp)
	res = resp.Data

	if err != nil {
		return nil, err
	}

	return res, nil
}
