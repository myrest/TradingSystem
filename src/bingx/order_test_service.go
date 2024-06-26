package bingx

import (
	"context"
	"encoding/json"
	"net/http"
)

func (s *CreateOrderService) DoTest(ctx context.Context, opts ...RequestOption) (res *CreateOrderResponse, err error) {
	r := &request{method: http.MethodPost, endpoint: "/openApi/swap/v2/trade/order"}

	if s.symbol != "" {
		r.addParam("symbol", s.symbol)
	}

	if s.orderType != "" {
		r.addParam("type", s.orderType)
	}

	if s.side != "" {
		r.addParam("side", s.side)
	}

	if s.positionSide != "" {
		r.addParam("positionSide", s.positionSide)
	} else {
		r.addParam("positionSide", BothPositionSideType)
	}

	if s.clientOrderID != "" {
		r.addParam("clientOrderID", s.clientOrderID)
	}

	if s.reduceOnly != "" {
		r.addParam("reduceOnly", s.reduceOnly)
	}

	if s.price != 0 {
		r.addParam("price", s.price)
	}

	if s.quantity != 0 {
		r.addParam("quantity", s.quantity)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	//fmt.Print(string(data))

	resp := new(struct {
		Code int                             `json:"code"`
		Msg  string                          `json:"msg"`
		Data map[string]*CreateOrderResponse `json:"data"`
	})

	err = json.Unmarshal(data, &resp)
	res = resp.Data["order"]

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *CancelOrderService) DoTest(ctx context.Context, opts ...RequestOption) (res *CancelOrderResponse, err error) {
	r := &request{method: http.MethodDelete, endpoint: "/openApi/swap/v2/trade/order"}

	if s.symbol != "" {
		r.addParam("symbol", s.symbol)
	}

	if s.orderId != 0 {
		r.addParam("orderId", s.orderId)
	}

	if s.clientOrderID != "" {
		r.addParam("clientOrderID", s.clientOrderID)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		Code int                             `json:"code"`
		Msg  string                          `json:"msg"`
		Data map[string]*CancelOrderResponse `json:"data"`
	})

	err = json.Unmarshal(data, &resp)

	if err != nil {
		return nil, err
	}

	res = resp.Data["order"]

	return res, nil
}

func (s *GetOrderService) DoTest(ctx context.Context, opts ...RequestOption) (res *GetOrderResponse, err error) {
	r := &request{method: http.MethodGet, endpoint: "/openApi/swap/v2/trade/order"}

	if s.symbol != "" {
		r.addParam("symbol", s.symbol)
	}

	if s.orderId != 0 {
		r.addParam("orderId", s.orderId)
	}

	if s.clientOrderID != "" {
		r.addParam("clientOrderID", s.clientOrderID)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		Code int                          `json:"code"`
		Msg  string                       `json:"msg"`
		Data map[string]*GetOrderResponse `json:"data"`
	})

	err = json.Unmarshal(data, &resp)

	if err != nil {
		return nil, err
	}

	res = resp.Data["order"]

	return res, nil
}

func (s *GetOpenOrdersService) DoTest(ctx context.Context, opts ...RequestOption) (res *GetOpenOrdersResponse, err error) {
	r := &request{method: http.MethodGet, endpoint: "/openApi/swap/v2/trade/openOrders"}

	if s.symbol != "" {
		r.addParam("symbol", s.symbol)
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data *GetOpenOrdersResponse `json:"data"`
	})

	err = json.Unmarshal(data, &resp)

	if err != nil {
		return nil, err
	}

	res = resp.Data

	return res, nil
}
