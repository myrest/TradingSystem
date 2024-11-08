package bitunix

import (
	"context"
	"encoding/json"
	"net/http"
)

type GetBalanceService struct {
	c *Client
}

type Balance struct {
	Coin          string `json:"coin"`
	Balance       string `json:"balance"`
	BalanceLocked string `json:"balanceLocked"`
}

func (s *GetBalanceService) Do(ctx context.Context, opts ...RequestOption) (res *[]Balance, err error) {
	r := &request{method: http.MethodGet, endpoint: "/api/spot/v1/user/account"}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		bitunixAPIRawResponse
		Data *[]Balance `json:"data"`
	})

	err = json.Unmarshal(data, resp)
	res = resp.Data

	if err != nil {
		return nil, err
	}

	return res, nil
}
