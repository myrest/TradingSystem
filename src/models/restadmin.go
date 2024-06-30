package models

type AdminCurrencySymbo struct {
	Symbo  string `json:"symbo"`
	Status bool   `json:"status"`
}

type CurrencySymbo struct {
	AdminCurrencySymbo
	Cert string `json:"cert"`
}

//GenerateRandomString

type CustomerCurrencySymbo struct {
	CurrencySymbo
	Amount     float64 `json:"amount"`
	CustomerID string
}
