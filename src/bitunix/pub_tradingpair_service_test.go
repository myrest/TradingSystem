package bitunix

import (
	"context"
	"testing"
)

func TestGetTradingPairService_Do(t *testing.T) {
	client := NewClient("xxx", "xxx")
	client.Debug = true
	res, err := client.NewGetTradingPairService().Do(context.Background())
	if err != nil {
		t.Errorf("GetTradingPairService() error = %v", err)
		return
	}
	t.Logf("GetTradingPairService() res = %v", res)
}
