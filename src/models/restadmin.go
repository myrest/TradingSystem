package models

type CurrencySymbolBase struct {
	Symbol string `json:"symbol"`
	Status bool   `json:"status"`
}

type AdminCurrencySymbol struct {
	CurrencySymbolBase
	Cert    string `json:"cert"`
	Message string `json:"message"`
}

type CustomerCurrencySymbol struct {
	CurrencySymbolBase
	Amount     float64 `json:"amount"`
	Leverage   float64 `json:"leverage"`
	Simulation bool    `json:"simulation"`
	CustomerID string
}

type CustomerCurrencySymbolUI struct {
	CustomerCurrencySymbol
	CustomerRelationUI
}

type CustomerCurrencySymboResponse struct {
	CurrencySymbolBase
	Amount       float64 `json:"amount"`
	Leverage     float64 `json:"leverage"`
	Simulation   bool    `json:"simulation"`
	Message      string  `json:"message"`
	SystemStatus string
}
