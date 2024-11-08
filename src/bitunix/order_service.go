package bitunix

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Only Market and Limit orders supported
type CreateOrderService struct {
	c         *Client
	symbol    string    //Trading pair
	orderType OrderType //Order Type(1:Limit 2:Market)
	side      SideType  //Side (1 Sell 2 Buy)
	price     string
	volume    string
}

func (s *CreateOrderService) Symbol(symbol string) *CreateOrderService {
	s.symbol = symbol
	return s
}

func (s *CreateOrderService) Type(orderType OrderType) *CreateOrderService {
	s.orderType = orderType
	return s
}

func (s *CreateOrderService) Side(side SideType) *CreateOrderService {
	s.side = side
	return s
}

func (s *CreateOrderService) Price(price string) *CreateOrderService {
	s.price = price
	return s
}

func (s *CreateOrderService) Quantity(quantity string) *CreateOrderService {
	s.volume = quantity
	return s
}

type CreateOrderResponse struct {
	OrderId     string `json:"orderId"`
	PlaceStatus int    `json:"placeStatus"` //1:Success 其它為:Fail
	PlaceCode   int    `json:"placeCode"`
	PlaceMsg    string `json:"placeMsg"`
}

func (s *CreateOrderService) Do(ctx context.Context, opts ...RequestOption) (res *CreateOrderResponse, err error) {
	r := &request{method: http.MethodPost, endpoint: "/api/spot/v1/order/place_order"}

	r.addParam("symbol", s.symbol)
	r.addParam("type", s.orderType)
	r.addParam("side", s.side)
	r.addParam("price", s.price)
	r.addParam("volume", s.volume)

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		bitunixAPIRawResponse
		Data *CreateOrderResponse `json:"data"`
	})

	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	res = resp.Data

	if res.PlaceStatus != 1 {
		err = fmt.Errorf("<Bitunix API Inner Error> code=%d, msg=%s", res.PlaceCode, res.PlaceMsg)
	}

	return res, err
}

type CancelOrderService struct {
	c             *Client
	symbol        string
	orderId       int
	clientOrderID string
}

func (s *CancelOrderService) Symbol(symbol string) *CancelOrderService {
	s.symbol = symbol
	return s
}

func (s *CancelOrderService) OrderId(orderId int) *CancelOrderService {
	s.orderId = orderId

	return s
}

func (s *CancelOrderService) ClientOrderId(clientOrderID string) *CancelOrderService {
	s.clientOrderID = clientOrderID

	return s
}

// Define response of cancel order request
type CancelOrderResponse struct {
	Time          int              `json:"time"`
	Symbol        string           `json:"symbol"`
	Side          SideType         `json:"side"`
	OrderType     OrderType        `json:"type"`
	PositionSide  PositionSideType `json:"positionSide"`
	CumQuote      string           `json:"cumQuote"`
	Status        OrderStatus      `json:"status"`
	StopPrice     string           `json:"stopPrice"`
	Price         string           `json:"price"`
	OrigQty       string           `json:"origQty"`
	AvgPrice      string           `json:"avgPrice"`
	ExecutedQty   string           `json:"executedQty"`
	OrderId       int              `json:"orderId"`
	Profit        string           `json:"profit"`
	Commission    string           `json:"commission"`
	UpdateTime    int              `json:"updateTime"`
	ClientOrderID string           `json:"clientOrderID"`
}

func (s *CancelOrderService) Do(ctx context.Context, opts ...RequestOption) (res *CancelOrderResponse, err error) {
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

type GetOrderService struct {
	c             *Client
	symbol        string
	orderId       int64
	clientOrderID string
}

// Define response of get order request
type GetOrderResponse struct {
	Time          int64            `json:"time"`
	Symbol        string           `json:"symbol"`
	Side          SideType         `json:"side"`
	OrderType     OrderType        `json:"type"`
	PositionSide  PositionSideType `json:"positionSide"`
	ReduceOnly    bool             `json:"reduceOnly"`
	CumQuote      string           `json:"cumQuote"`
	Status        OrderStatus      `json:"status"`
	StopPrice     string           `json:"stopPrice"`
	Price         string           `json:"price"`
	OrigQuantity  string           `json:"origQty"`
	AveragePrice  string           `json:"avgPrice"`
	Quantity      string           `json:"executedQty"`
	OrderId       int              `json:"orderId"`
	Profit        string           `json:"profit"`
	Fee           string           `json:"commission"`
	UpdateTime    int64            `json:"ppdateTime"`
	WorkingType   OrderWorkingType `json:"workingType"`
	ClientOrderID string           `json:"clientOrderID"`
}

func (s *GetOrderService) Symbol(symbol string) *GetOrderService {
	s.symbol = symbol
	return s
}

func (s *GetOrderService) OrderId(orderId int64) *GetOrderService {
	s.orderId = orderId
	return s
}

func (s *GetOrderService) ClientOrderId(clientOrderID string) *GetOrderService {
	s.clientOrderID = clientOrderID
	return s
}

func (s *GetOrderService) Do(ctx context.Context, opts ...RequestOption) (res *GetOrderResponse, err error) {
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

type GetOpenOrdersService struct {
	c      *Client
	symbol string
}

// Define response of get order request
type GetOpenOrdersResponse struct {
	Orders []*GetOrderResponse `json:"orders"`
}

// type OpenOrderResponse struct {
// 	Symbol        string           `json:"symbol"`
// 	OrderId       int              `json:"orderId"`
// 	Side          SideType         `json:"side"`
// 	PositionSide  PositionSideType `json:"positionSide"`
// 	OrderType     OrderType        `json:"type"`
// 	OrigQuantity  string           `json:"origQty"`
// 	Price         string           `json:"price"`
// 	Quantity      string           `json:"executedQty"`
// 	AveragePrice  string           `json:"avgPrice"`
// 	CumQuote      string           `json:"cumQuote"`
// 	StopPrice     string           `json:"stopPrice"`
// 	Profit        string           `json:"profit"`
// 	Fee           string           `json:"commission"`
// 	Status        OrderStatus      `json:"status"`
// 	Time          int64            `json:"time"`
// 	UpdateTime    int64            `json:"ppdateTime"`
// 	WorkingType   OrderWorkingType `json:"workingType"`
// 	ClientOrderID string           `json:"clientOrderID"`
// }

func (s *GetOpenOrdersService) Symbol(symbol string) *GetOpenOrdersService {
	s.symbol = symbol
	return s
}

func (s *GetOpenOrdersService) Do(ctx context.Context, opts ...RequestOption) (res *GetOpenOrdersResponse, err error) {
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
