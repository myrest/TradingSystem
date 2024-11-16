package binance_connector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Do send request
// orderId or origClientOrderId 必需要有一個
// https://developers.binance.com/docs/derivatives/usds-margined-futures/trade/rest-api/Query-Current-Open-Order
func (s *GetCurrentOpenOrderService) Do(ctx context.Context, opts ...RequestOption) (res []*NewCurrentOpenOrderResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/openOrder",
		secType:  secTypeSigned,
	}
	isPassPamater := false
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}

	if s.orderid != nil {
		r.setParam("orderId", *s.orderid)
		isPassPamater = true
	}

	if s.origClientOrderId != nil {
		r.setParam("origClientOrderId", *s.origClientOrderId)
		isPassPamater = true
	}

	if !isPassPamater {
		return nil, fmt.Errorf("must set orderId or origClientOrderId")
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = make([]*NewCurrentOpenOrderResponse, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// https://developers.binance.com/docs/derivatives/usds-margined-futures/trade/rest-api/Current-All-Open-Orders
func (s *GetAllOpenOrderService) Do(ctx context.Context, opts ...RequestOption) (res []*NewCurrentOpenOrderResponse, err error) {
	s.c.Debug = true
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/openOrders",
		secType:  secTypeSigned,
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = make([]*NewCurrentOpenOrderResponse, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
