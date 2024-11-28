package services

import (
	"TradingSystem/src/binance_connector"
	"TradingSystem/src/bingx"
	"TradingSystem/src/models"
	"TradingSystem/src/strategyinterface"
)

func GetTradingClient(apiKey, secretKey string, isTEST bool, ExchangeName models.ExchangeSystem) (client strategyinterface.TradingClient) {
	switch ExchangeName {
	case models.ExchangeBingx:
		client = bingx.NewClient(apiKey, secretKey, isTEST)
	case models.ExchangeBinance_P:
		client = binance_connector.NewPortfolioClient(apiKey, secretKey)
	}
	return
}
