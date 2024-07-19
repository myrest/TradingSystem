package unittest

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"testing"
)

func TestCreateCustomer(t *testing.T) {
	testEmail := "johndoe@example.com"
	c := context.Background()
	CreateCustomerTEST(t, c, testEmail, false)
	GetCustomerTEST(t, c, testEmail)
	UpdateCustomerTest(t, c, testEmail)
	DeleteAccountTest(t, c, testEmail)
}

// 預設都為False
func CreateCustomerTEST(t *testing.T, c context.Context, testEmail string, OnlyTestCreation bool) {
	type args struct {
		customer *models.Customer
	}
	tests := []struct {
		name         string
		args         args
		wantNotEmpty bool
		wantErr      bool
	}{
		{
			name: "建立帳號",
			args: args{
				customer: &models.Customer{
					Name:  "John Doe",
					Email: testEmail,
				},
			},
			wantNotEmpty: true,
			wantErr:      false,
		},
		{
			name: "重覆建立帳號",
			args: args{
				customer: &models.Customer{
					Name:      "John Doe",
					Email:     testEmail,
					APIKey:    "apikey",
					SecretKey: "secretkey",
					IsAdmin:   false,
				},
			},
			wantNotEmpty: false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := services.CreateCustomer(c, tt.args.customer)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCustomer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNotEmpty && got == "" {
				t.Errorf("CreateCustomer() = %v, want %v", got, "沒有取得CustomerID")
			}
		})
		if OnlyTestCreation {
			return
		}
	}
}

func DeleteAccountTest(t *testing.T, c context.Context, testEmail string) {
	t.Run("刪除帳號", func(t *testing.T) {
		dbCustomer, _ := services.GetCustomerByEmail(c, testEmail)
		if dbCustomer != nil {
			services.DeleteCustomer(c, dbCustomer.ID)
		}

		dbCustomer, err := services.GetCustomerByEmail(c, testEmail)
		if err != nil {
			t.Errorf("GetCustomerByEmail() error = %v", err)
		}
		if dbCustomer != nil {
			t.Errorf("GetCustomerByEmail() 測試 帳號刪除失敗 = %v", dbCustomer.ID)
		}
	})
}

func GetCustomerTEST(t *testing.T, c context.Context, email string) {
	t.Run("依Email取得帳號資料", func(t *testing.T) {
		_, err := services.GetCustomerByEmail(c, email)
		if err != nil {
			t.Errorf("GetCustomerByEmail() error = %v", err)
			return
		}
	})

	t.Run("依不存在的Email取得帳號資料", func(t *testing.T) {
		dbCustomer, err := services.GetCustomerByEmail(c, "NoEmail")
		if err != nil {
			t.Errorf("GetCustomerByEmail() error = %v", err)
			return
		}
		if dbCustomer != nil {
			t.Errorf("GetCustomerByEmail() 不應該取得帳號，CustomerID = %v", dbCustomer.ID)
			return
		}
	})
}

func UpdateCustomerTest(t *testing.T, c context.Context, email string) {
	newName := "RoyTEST"
	t.Run("更新帳號", func(t *testing.T) {
		dbCustomer, err := services.GetCustomerByEmail(c, email)
		if err != nil {
			t.Errorf("GetCustomerByEmail() error = %v", err)
			return
		}
		dbCustomer.Name = newName
		err = services.UpdateCustomer(c, dbCustomer)
		if err != nil {
			t.Errorf("UpdateCustomer() error = %v", err)
			return
		}
		dbCustomer, err = services.GetCustomerByEmail(c, email)
		if err != nil {
			t.Errorf("GetCustomerByEmail() error = %v", err)
			return
		}
		if dbCustomer == nil {
			t.Errorf("GetCustomerByEmail() Can't get test account")
			return
		}
		if dbCustomer.Name != newName {
			t.Errorf("GetCustomerByEmail() Update data do not write to DB.")
			return
		}
	})
}
