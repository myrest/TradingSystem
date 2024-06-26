package models

type CurrencySymbo struct {
	Symbo  string `json:"symbo"`
	Status bool   `json:"status"`
}

type CustomerCurrencySymbo struct {
	CurrencySymbo
	Amount     float64 `json:"amount"`
	CustomerID string
}
