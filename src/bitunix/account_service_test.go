package bitunix

import (
	"context"
	"testing"
)

func TestGetBalanceService_Do(t *testing.T) {
	client := NewClient("xxx", "xxx")
	client.Debug = true
	res, err := client.NewGetBalanceService().Do(context.Background())
	if err != nil {
		t.Errorf("GetBalanceService() error = %v", err)
		return
	}
	t.Logf("GetBalanceService() res = %v", res)
}
