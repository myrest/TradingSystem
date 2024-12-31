package services

import (
	"TradingSystem/src/binance_connector"
	"TradingSystem/src/binance_connector_portfolio"
	"TradingSystem/src/bingx"
	"TradingSystem/src/bitunix_feature"
	"TradingSystem/src/models"
	"TradingSystem/src/strategyinterface"
	"context"
)

func GetTradingClient(apiKey, secretKey string, isTEST bool, ExchangeName models.ExchangeSystem) (client strategyinterface.TradingClient) {
	switch ExchangeName {
	case models.ExchangeBinance_P:
		client = binance_connector_portfolio.NewPortfolioClient(apiKey, secretKey)
	case models.ExchangeBinance_N:
		client = binance_connector.NewClient(apiKey, secretKey)
	case models.ExchangeBitunix_Feature:
		client = bitunix_feature.NewClient(apiKey, secretKey)
	default:
		client = bingx.NewClient(apiKey, secretKey, isTEST)
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
