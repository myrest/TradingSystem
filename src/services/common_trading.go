package services

import (
	"TradingSystem/src/binance_connector_portfolio"
	"TradingSystem/src/bingx"
	"TradingSystem/src/models"
	"TradingSystem/src/strategyinterface"
	"context"
)

func GetTradingClient(apiKey, secretKey string, isTEST bool, ExchangeName models.ExchangeSystem) (client strategyinterface.TradingClient) {
	switch ExchangeName {
	case models.ExchangeBingx:
		client = bingx.NewClient(apiKey, secretKey, isTEST)
	case models.ExchangeBinance_P:
		client = binance_connector_portfolio.NewPortfolioClient(apiKey, secretKey)
	}
	return
}

func GetAccountBalance(c context.Context, APIkey, SecretKey string, ExchangeName models.ExchangeSystem) (float64, error) {
	client := GetTradingClient(APIkey, SecretKey, false, ExchangeName)
	res, err := client.GetBalance(c)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func UpdateLeverage(c context.Context, APIkey, SecretKey string, ExchangeName models.ExchangeSystem, Symbol string, Leverage int64) error {
	client := GetTradingClient(APIkey, SecretKey, false, ExchangeName)
	return client.UpdateLeverage(c, Symbol, Leverage)
}
