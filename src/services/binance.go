package services

import (
	"TradingSystem/src/binance_connector"
	"TradingSystem/src/models"
	"context"
	"fmt"
)

func BinPositionSiteIsDual(c context.Context, customer *models.Customer) (bool, error) {
	//查出目前持倉設定
	client := binance_connector.NewPortfolioClient(customer.APIKey, customer.SecretKey)
	position, err := client.GetUMPositionService().Do(c)
	if err != nil {
		return false, err
	}
	return position.DualSidePosition, nil
}

func BinUpdatePosition(c context.Context, customer *models.Customer, isbodual ...bool) error {
	//預設改為雙向持倉
	var dual string
	isdual := true
	if len(dual) > 0 {
		isdual = isbodual[0]
	}
	dual = fmt.Sprintf("%t", isdual)

	//查出目前持倉設定
	currentIsDual, err := BinPositionSiteIsDual(c, customer)
	if err != nil {
		return err
	}

	//如果 isdual = position.DualSidePosition 表示方向一致，不用再改了。
	if isdual == currentIsDual {
		return fmt.Errorf("持倉設定相同，不用再改")
	}

	client := binance_connector.NewPortfolioClient(customer.APIKey, customer.SecretKey)
	_, err = client.GetUMPositionService().DualSidePosition(dual).DoUpdate(c)
	if err != nil {
		return err
	}

	return nil
}
