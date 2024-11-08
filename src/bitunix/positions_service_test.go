package bitunix

import (
	"context"
	"testing"
)

func TestGetOpenPositionsService_Do(t *testing.T) {
	//client := NewClient("513bc96f854b7b71d458df5ec5da09d7", "8e3cbf9cac443c679e3579b858875cb2")
	client := NewClient("xxx", "xxx")
	client.Debug = true
	res, err := client.NewGetOpenPositionsService().Symbol("btcusdt").Do(context.Background())
	if err != nil {
		t.Errorf("GetTradingPairService() error = %v", err)
		return
	}
	t.Logf("GetTradingPairService() res = %v", res)
}
