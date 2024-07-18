package unittest

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"testing"
)

// 客戶端幣種增刪修
// 刪還沒寫
func TestUpdateCustomerCurrency(t *testing.T) {
	testEmail := "johndoe@example.com"
	adminSymbol := "TESTUSDT.P"
	c := context.Background()
	//先建立測試需要的資料
	CreateAdminSymbolWithoutAutoSubscriberTest(t, c, adminSymbol, true)
	CreateCustomerTEST(t, c, testEmail, true)

	//本函式測試標的
	UpdateCustomerCurrencyTest(t, c, testEmail, adminSymbol, false)
	DeleteCustomerCurrencyTest(t, c, testEmail, adminSymbol)

	//刪除測試資料
	DeleteAdminSymbol(t, c, adminSymbol)
	DeleteAccountTest(t, c, testEmail)
}

func UpdateCustomerCurrencyTest(t *testing.T, c context.Context, TestEmail, AdminSymbol string, OnlyCreation bool) {
	type args struct {
		ctx              context.Context
		customercurrency *models.CustomerCurrencySymbol
	}

	customer, _ := services.GetCustomerByEmail(c, TestEmail)
	if customer == nil {
		t.Errorf("在準備階段，取不到客戶資料。")
		return
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "建立客戶端幣種",
			args: args{
				ctx: c,
				customercurrency: &models.CustomerCurrencySymbol{
					CurrencySymbolBase: models.CurrencySymbolBase{
						Symbol: AdminSymbol,
						Status: false,
					},
					Amount:     1,
					Simulation: true,
					CustomerID: customer.ID,
				},
			},
			wantErr: false,
		},
		{
			name: "建立不存在的客戶端幣種",
			args: args{
				ctx: c,
				customercurrency: &models.CustomerCurrencySymbol{
					CurrencySymbolBase: models.CurrencySymbolBase{
						Symbol: AdminSymbol,
						Status: false,
					},
					Amount:     0,
					Simulation: true,
					CustomerID: "Not Exist Customer ID",
				},
			},
			wantErr: true,
		},
		{
			name: "建立客戶存在幣種不存在",
			args: args{
				ctx: c,
				customercurrency: &models.CustomerCurrencySymbol{
					CurrencySymbolBase: models.CurrencySymbolBase{
						Symbol: "Sumbol Not Exist",
						Status: false,
					},
					Amount:     0,
					Simulation: true,
					CustomerID: customer.ID,
				},
			},
			wantErr: true,
		},
		{
			name: "更新客戶幣種",
			args: args{
				ctx: c,
				customercurrency: &models.CustomerCurrencySymbol{
					CurrencySymbolBase: models.CurrencySymbolBase{
						Symbol: AdminSymbol,
						Status: false,
					},
					Amount:     0,
					Simulation: true,
					CustomerID: customer.ID,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := services.UpdateCustomerCurrency(tt.args.ctx, tt.args.customercurrency); (err != nil) != tt.wantErr {
				t.Errorf("UpdateCustomerCurrency() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if OnlyCreation {
			return
		}
	}
}

func DeleteCustomerCurrencyTest(t *testing.T, c context.Context, TestEmail, AdminSymbol string) {
	customer, _ := services.GetCustomerByEmail(c, TestEmail)
	if customer == nil {
		t.Errorf("在準備刪除階段，取不到客戶資料。")
		return
	}

	err := services.DeleteCustomerCurrency(c, customer.ID, AdminSymbol)
	if err != nil {
		t.Errorf("DeleteCustomerCurrency() error = %v", err)
		return
	}
}
