package binance_connector

// Binance Get current open orders (GET /api/v3/openOrders)
// GetOpenOrdersService get open orders
type GetCurrentOpenOrderService struct {
	c                 *Client
	symbol            *string
	orderid           *int64
	origClientOrderId *string
}

// Create NewCurrentOpenOrderResponse
type NewCurrentOpenOrderResponse struct {
	AvgPrice                string `json:"avgPrice"`
	ClientOrderID           string `json:"clientOrderId"`
	CumQuote                string `json:"cumQuote"`
	ExecutedQty             string `json:"executedQty"`
	OrderID                 int64  `json:"orderId"`
	OrigQty                 string `json:"origQty"`
	OrigType                string `json:"origType"`
	Price                   string `json:"price"`
	ReduceOnly              bool   `json:"reduceOnly"`
	Side                    string `json:"side"`
	PositionSide            string `json:"positionSide"`
	Status                  string `json:"status"`
	StopPrice               string `json:"stopPrice"`     // to ignore when order type is TRAILING_STOP_MARKET
	ClosePosition           bool   `json:"closePosition"` // if Close-All
	Symbol                  string `json:"symbol"`
	Time                    int64  `json:"time"` // order time
	TimeInForce             string `json:"timeInForce"`
	Type                    string `json:"type"`          // TRAILING_STOP_MARKET
	ActivatePrice           string `json:"activatePrice"` // activation price, only return with TRAILING_STOP_MARKET order
	PriceRate               string `json:"priceRate"`     // callback rate, only return with TRAILING_STOP_MARKET order
	UpdateTime              int64  `json:"updateTime"`
	WorkingType             string `json:"workingType"`
	PriceProtect            bool   `json:"priceProtect"`            // if conditional order trigger is protected
	PriceMatch              string `json:"priceMatch"`              // price match mode
	SelfTradePreventionMode string `json:"selfTradePreventionMode"` // self trading prevention mode
	GoodTillDate            int64  `json:"goodTillDate"`            // order pre-set auto cancel time for TIF GTD order
}

// Symbol set symbol
func (s *GetCurrentOpenOrderService) Symbol(symbol string) *GetCurrentOpenOrderService {
	s.symbol = &symbol
	return s
}

// Symbol set orderid
func (s *GetCurrentOpenOrderService) Orderid(orderid int64) *GetCurrentOpenOrderService {
	s.orderid = &orderid
	return s
}

func (s *GetCurrentOpenOrderService) OrigClientOrderId(origClientOrderId string) *GetCurrentOpenOrderService {
	s.origClientOrderId = &origClientOrderId
	return s
}

// Get All Open orders
type GetAllOpenOrderService struct {
	c      *Client
	symbol *string
}

func (s *GetAllOpenOrderService) Symbol(symbol string) *GetAllOpenOrderService {
	s.symbol = &symbol
	return s
}
