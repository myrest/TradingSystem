package services

import (
	"TradingSystem/src/models"
	"context"
)

func GetAccountBalance(c context.Context, APIkey, SecretKey string, ExchangeName models.ExchangeSystem) (float64, error) {
	client := GetTradingClient(APIkey, SecretKey, false, ExchangeName)
	res, err := client.GetBalance(c)
	if err != nil {
		return 0, err
	}
	return res, nil
}
