package bitunix_feature

import (
	"TradingSystem/src/common"
	"context"
)

func (Client *Client) GetBalance(ctx context.Context) (float64, error) {
	asset, err := Client.GetAccountBalanceService().Symbol("USDT").Do(ctx)
	if err != nil {
		return 0, err
	}
	return common.Decimal(asset.Data.Available), nil
}

// 要改槓桿及改成雙向持倉
func (Client *Client) UpdateLeverage(ctx context.Context, symbol string, leverage int64) error {
	//改多空槓桿
	_, err := Client.GetUpdateLeverageService().
		Symbol(common.FormatSymbol(symbol, false)).
		Leverage(leverage).
		Do(ctx)
	if err != nil {
		return err
	}

	//取得目前持倉模式
	asset, err := Client.GetAccountBalanceService().Symbol("USDT").Do(ctx)
	if err != nil {
		return err
	}

	if asset.Data.PositionMode != string(HEDGE) {
		//如果不是雙向持倉，需要改成雙向持倉
		_, err := Client.GetUpdatePositionService().DualSidePosition(HEDGE).Do(ctx)
		if err != nil {
			return err
		}
	}

	//取得目前全逐倉模式
	IsoOrCloseMode, err := Client.GetLeverageMarginTypeService().
		Symbol(symbol).
		MarginCoin("USDT").
		Do(ctx)
	if err != nil {
		return err
	}

	if IsoOrCloseMode.Data.MarginMode != string(MarginCrossed) {
		//修改全逐倉模式
		_, err = Client.GetUpdateMarginTypeService().
			Symbol(symbol).
			MarginType(MarginCrossed).
			MarginCoin("USDT").
			Do(ctx)
		if err != nil {
			return nil
		}
	}

	return nil
}
