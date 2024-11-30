package strategyinterface

import (
	"TradingSystem/src/models"
	"context"
)

type TradingClient interface {
	CreateOrder(ctx context.Context, orderData models.TvSiginalData, customer models.CustomerCurrencySymboWithCustomer) (models.Log_TvSiginalData, bool, models.AlertMessageModel, error)
	GetBalance(ctx context.Context) (float64, error)
	UpdateLeverage(ctx context.Context, symbol string, leverage int64) error
	//GetOpenPositions(ctx context.Context, symbol string, customer models.CustomerCurrencySymboWithCustomer) ([]models.Position, error)
}
