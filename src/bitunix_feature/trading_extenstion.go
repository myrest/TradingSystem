package bitunix_feature

import (
	"context"
	"encoding/json"
	"net/http"
)

// region 可用餘額
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
