package bingx

import (
	"context"
	"encoding/json"
	"net/http"
)

type SetMarginTypeService struct {
	c          *Client
	symbol     string
	margintype MarginTradingType
}

type GetMarginTypeService struct {
	c      *Client
	symbol string
}

func (s *SetMarginTypeService) Symbol(symbol string) *SetMarginTypeService {
	s.symbol = symbol
	return s
}

func (s *SetMarginTypeService) Margin(mt MarginTradingType) *SetMarginTypeService {
	s.margintype = mt
	return s
}

type MarginType struct {
	Symbol string `json:"symbol"`     //交易對
	Margin string `json:"marginType"` //保證金模式 ISOLATED(逐倉), CROSSED(全倉)
}

func (s *SetMarginTypeService) Do(ctx context.Context, opts ...RequestOption) (res *LeverageType, err error) {
	r := &request{method: http.MethodPost, endpoint: "/openApi/swap/v2/trade/marginType"}

	if s.symbol != "" {
		r.addParam("symbol", s.symbol)
	}

	if s.symbol != "" {
		r.addParam("marginType", s.margintype)
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

func (s *GetMarginTypeService) Do(ctx context.Context, opts ...RequestOption) (res *LeverageType, err error) {
	r := &request{method: http.MethodGet, endpoint: "/openApi/swap/v2/trade/marginType"}

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
