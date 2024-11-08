package bitunix

import (
	"context"
	"testing"
)

func TestCreateOrderService_Do(t *testing.T) {
	client := NewClient("xxx", "xxx")
	client.Debug = true
	res, err := client.NewCreateOrderService().
		Symbol("btcusdt").
		Type(MarketOrderType).
		Side(SellSideType).
		Price("80000").
		Quantity("10.3896").
		Do(context.Background())
	if err != nil {
		t.Errorf("CreateOrderService() error = %v", err)
		return
	}
	t.Logf("CreateOrderService() res = %v", res)
}
