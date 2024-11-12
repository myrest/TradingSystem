package models

type SideType string
type PositionSideType string

const (
	BuySideType  SideType = "BUY"
	SellSideType SideType = "SELL"

	ShortPositionSideType PositionSideType = "SHORT"
	LongPositionSideType  PositionSideType = "LONG"
)

type ______OrderResponse struct {
	Time         int64            `json:"time"`
	Symbol       string           `json:"symbol"`
	Side         SideType         `json:"side"`
	PositionSide PositionSideType `json:"positionSide"`
	Price        string           `json:"price"`
	AveragePrice string           `json:"avgPrice"`
	Quantity     string           `json:"executedQty"`
	OrderId      string           `json:"orderId"`
	Profit       string           `json:"profit"`
	Fee          string           `json:"commission"`
}

type OpenPosition struct {
	AvailableAmt float64
	PositionSide PositionSideType
}
