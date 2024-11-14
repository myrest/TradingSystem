package strategyinterface

import (
	"TradingSystem/src/models"
	"context"
)

type TradingClient interface {
	CreateOrder(ctx context.Context, orderData models.TvSiginalData, customer models.CustomerCurrencySymboWithCustomer) (models.Log_TvSiginalData, bool, models.AlertMessageModel, error)
	//GetOpenPositions(ctx context.Context, symbol string, customer models.CustomerCurrencySymboWithCustomer) ([]models.Position, error)
}
