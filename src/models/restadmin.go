package models

type AdminCurrencySymbol struct {
	Symbol string `json:"symbo"`
	Status bool   `json:"status"`
}

type CurrencySymbol struct {
	AdminCurrencySymbol
	Cert string `json:"cert"`
}

//GenerateRandomString

type CustomerCurrencySymbol struct {
	CurrencySymbol
	Amount     float64 `json:"amount"`
	CustomerID string
}

type AdminSymboListUI struct {
	CurrencySymbol
	PositionSize string `json:"positionsize"`
}
