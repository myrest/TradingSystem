package unittest

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"testing"
)

func TestCreateCustomer(t *testing.T) {
	testEmail := "johndoe@example.com"
	type args struct {
		ctx      context.Context
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
				ctx: context.Background(),
				customer: &models.Customer{
					Name:      "John Doe",
					Email:     testEmail,
					APIKey:    "apikey",
					SecretKey: "secretkey",
					IsAdmin:   false,
				},
			},
			wantNotEmpty: true,
			wantErr:      false,
		},
		{
			name: "重覆建立帳號",
			args: args{
				ctx: context.Background(),
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
			got, err := services.CreateCustomer(tt.args.ctx, tt.args.customer)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCustomer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNotEmpty && got == "" {
				t.Errorf("CreateCustomer() = %v, want %v", got, "沒有取得CustomerID")
			}
		})
	}

	c := context.Background()
	GetCustomerTEST(t, c, testEmail)
	UpdateCustomerTest(t, c, testEmail)

	dbCustomer, _ := services.GetCustomerByEmail(c, testEmail)
	if dbCustomer != nil {
		services.DeleteCustomer(c, dbCustomer.ID)
	}
	t.Run("GetCustomerByEmail", func(t *testing.T) {
		dbCustomer, err := services.GetCustomerByEmail(c, testEmail)
		if err != nil {
			t.Errorf("GetCustomer() error = %v", err)
			return
		}
		if dbCustomer != nil {
			t.Errorf("GetCustomerByEmail() 測試 帳號刪除失敗 = %v", dbCustomer.ID)
			return
		}
	})
}

func GetCustomerTEST(t *testing.T, c context.Context, email string) {
	t.Run("GetCustomerByEmail", func(t *testing.T) {
		_, err := services.GetCustomerByEmail(c, email)
		if err != nil {
			t.Errorf("GetCustomerByEmail() error = %v", err)
			return
		}
	})

	t.Run("GetCustomerByEmail", func(t *testing.T) {
		dbCustomer, err := services.GetCustomerByEmail(c, "NoEmail")
		if err != nil {
			t.Errorf("GetCustomerByEmail() error = %v", err)
			return
		}
		if dbCustomer != nil {
			t.Errorf("GetCustomerByEmail() 怎麼可能有帳號？ = %v", dbCustomer.ID)
			return
		}
	})
}

func UpdateCustomerTest(t *testing.T, c context.Context, email string) {
	newName := "RoyTEST"
	t.Run("GetCustomerByEmail", func(t *testing.T) {
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
