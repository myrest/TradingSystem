package unittest

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"reflect"
	"testing"
)

func TestCreateNewSymbol(t *testing.T) {
	c := context.Background()
	symbol := "TESTUSDT.P"
	CreateAdminSymbolTest(t, c, symbol, false)

	//刪除最後再測試
	DeleteAdminSymbol(t, c, symbol)
}

func CreateAdminSymbolTest(t *testing.T, c context.Context, Symbol string, OnlyCreation bool) {
	type args struct {
		ctx    context.Context
		Symbol models.AdminCurrencySymbol
	}
	tests := []struct {
		name    string
		args    args
		want    *models.AdminCurrencySymbol
		wantErr bool
	}{
		{
			name: "新增幣種",
			args: args{
				ctx: c,
				Symbol: models.AdminCurrencySymbol{
					CurrencySymbolBase: models.CurrencySymbolBase{
						Symbol: Symbol,
						Status: false,
					},
					Cert:    "AAAA",
					Message: "TEST Symbol",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "新增重覆幣種",
			args: args{
				ctx: c,
				Symbol: models.AdminCurrencySymbol{
					CurrencySymbolBase: models.CurrencySymbolBase{
						Symbol: Symbol,
						Status: false,
					},
					Cert:    "AAAA",
					Message: "TEST Symbol",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := services.CreateNewSymbol(tt.args.ctx, tt.args.Symbol)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNewSymbol() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateNewSymbol() = %v, want %v", got, tt.want)
			}
		})
		if OnlyCreation {
			return
		}
	}
}

func DeleteAdminSymbol(t *testing.T, c context.Context, Symbol string) {
	t.Run("測試刪除幣種", func(t *testing.T) {
		err := services.DeleteAdminSymbol(c, Symbol)
		if err != nil {
			t.Errorf("DeleteAdminSymbol() error = %v", err)
		}
	})
}
