package models

type CurrencySymbolBase struct {
	Symbol  string `json:"symbol"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type AdminCurrencySymbol struct {
	CurrencySymbolBase
	Cert string `json:"cert"`
}

type CustomerCurrencySymbol struct {
	CurrencySymbolBase
	Amount     float64 `json:"amount"`
	Simulation bool    `json:"simulation"`
	CustomerID string
}

type CustomerCurrencySymboResponse struct {
	CurrencySymbolBase
	Amount       float64 `json:"amount"`
	Simulation   bool    `json:"simulation"`
	SystemStatus string
}

type AdminSymboListUI struct {
	AdminCurrencySymbol
	PositionSize string `json:"positionsize"`
}
