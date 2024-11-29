package binance_connector

import (
	"TradingSystem/src/models"
	"context"
)

type MyBinanceClient struct {
	*Client
}

// func (client MyBinanceClient) CreateOrder(c context.Context, tv models.TvSiginalData, Customer models.CustomerCurrencySymboWithCustomer) (models.Log_TvSiginalData, bool, models.AlertMessageModel, error) {
func (client MyBinanceClient) CreateOrder(c context.Context, tv models.TvSiginalData, Customer models.CustomerCurrencySymboWithCustomer) {

}

func (Client *MyBinanceClient) GetBalance(ctx context.Context) (float64, error) {

	return 0, nil
}

// 要改槓桿及改成雙向持倉
func (Client *MyBinanceClient) UpdateLeverage(ctx context.Context, symbol string, leverage int64) error {

	return nil
}
