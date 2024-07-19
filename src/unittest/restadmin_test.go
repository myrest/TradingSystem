package unittest

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"testing"
)

const (
	testEmailWithAutoSubscriber    = "WithAutoSubscriber@example.com"
	testEmailWithOutAutoSubscriber = "WithOutAutoSubscriber@example.com"
)

func TestCreateNewSymbol(t *testing.T) {
	c := context.Background()
	symbol := "TESTUSDT.P"
	//建立不支援自動訂閱的測試帳號
	CreateCustomerTEST(t, c, testEmailWithOutAutoSubscriber, true)
	CreateCustomerWithAutoSubscribe(t, c, testEmailWithAutoSubscriber)

	//測試標的物
	CreateAdminSymbolWithoutAutoSubscriberTest(t, c, symbol, false)

	//刪除最後再測試
	DeleteAdminSymbol(t, c, symbol)
	DeleteCustomer(t, c, testEmailWithOutAutoSubscriber, "刪除測試無自動跟單帳號")
	DeleteCustomer(t, c, testEmailWithAutoSubscriber, "刪除測試自動跟單帳號")
}

func CreateCustomerWithAutoSubscribe(t *testing.T, c context.Context, Email string) {
	customer := &models.Customer{
		Name:             "WithAutoSubscribe",
		Email:            Email,
		APIKey:           "apikey",
		SecretKey:        "secretkey",
		IsAdmin:          false,
		IsAutoSubscribe:  true,
		AutoSubscribReal: false,
	}
	t.Run("建立自動跟單帳號", func(t *testing.T) {
		_, err := services.CreateCustomer(c, customer)
		if err != nil {
			t.Errorf("CreateCustomer() error = %v", err)
			return
		}
	})
}

func DeleteCustomer(t *testing.T, c context.Context, Email, Subject string) {
	t.Run(Subject, func(t *testing.T) {
		dbCustomer, _ := services.GetCustomerByEmail(c, Email)
		if dbCustomer != nil {
			services.DeleteCustomer(c, dbCustomer.ID)
		}

		dbCustomer, err := services.GetCustomerByEmail(c, Email)
		if err != nil {
			t.Errorf("GetCustomerByEmail() error = %v", err)
		}
		if dbCustomer != nil {
			t.Errorf("GetCustomerByEmail() 測試帳號刪除失敗 = %v", dbCustomer.ID)
		}
	})
}

// 本測試標的物
func CreateAdminSymbolWithoutAutoSubscriberTest(t *testing.T, c context.Context, Symbol string, OnlyCreation bool) {
	//新增AdminSymbol，要檢查自動訂閱的情形
	var err error
	defAdminSymbol := models.AdminCurrencySymbol{
		CurrencySymbolBase: models.CurrencySymbolBase{
			Symbol: Symbol,
			Status: false,
		},
		Message: "TEST Symbol",
	}

	_, err = services.CreateNewSymbol(c, defAdminSymbol)
	t.Run("建立新幣種", func(t *testing.T) {
		if err != nil {
			t.Errorf("CreateNewSymbol() error = %v", err)
			return
		}
	})

	if OnlyCreation {
		return
	}
	WithOutCustomer, _ := services.GetCustomerByEmail(c, testEmailWithOutAutoSubscriber)
	WithCustomer, _ := services.GetCustomerByEmail(c, testEmailWithAutoSubscriber)

	t.Run("檢查建立新幣種自動套用", func(t *testing.T) {
		customerSymbol, err := services.GetCustomerCurrency(c, WithCustomer.ID, Symbol)
		if err != nil {
			t.Errorf("GetCustomerCurrency() error = %v", err)
			return
		}
		if customerSymbol == nil {
			t.Errorf("自動套用至有設定AutoSubscriber的客戶失敗。")
			return
		}
		if customerSymbol.Simulation == WithCustomer.AutoSubscribReal {
			t.Errorf("自動套用至盤種設定(模擬、實盤)錯誤。")
			return
		}
		if customerSymbol.Status != true {
			t.Errorf("自動套用Status需要為True。")
			return
		}
	})

	t.Run("檢查建立新幣種手動套用", func(t *testing.T) {
		customerSymbol, err := services.GetCustomerCurrency(c, WithOutCustomer.ID, Symbol)
		if err != nil {
			t.Errorf("取不到客戶資料 error = %v", err)
			return
		}
		if customerSymbol != nil {
			t.Errorf("不能自動套用未設定AutoSubscriber的客戶。")
		}
	})
}

func DeleteAdminSymbol(t *testing.T, c context.Context, Symbol string) {
	t.Run("測試刪除幣種", func(t *testing.T) {
		err := services.DeleteAdminSymbol(c, Symbol)
		if err != nil {
			t.Errorf("DeleteAdminSymbol() error = %v", err)
		}
	})
}
