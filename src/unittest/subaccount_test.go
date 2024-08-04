package unittest

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"testing"
)

func TestCreateSubAccount(t *testing.T) {
	testCustomerID := "TestCustomerID"
	c := context.Background()

	sampleData := models.SubAccount{
		SubAccountDB: models.SubAccountDB{
			CustomerID:  testCustomerID,
			AccountName: "ThisisFirstSubaccount",
		},
	}

	CreateSubaccountTEST(t, c, sampleData, false)
	dbSubaccounts, _ := services.GetSubaccountListByID(c, sampleData.CustomerID)
	dbSubaccount := dbSubaccounts[0] //應該只會有一筆，且為第一筆
	dbSecondsubaccount := UpdateSubaccountTest(t, c, dbSubaccount)
	DeleteSubaccountTest(t, c, dbSubaccount, dbSecondsubaccount)
}

// 建立測試資料
func CreateSubaccountTEST(t *testing.T, c context.Context, sample models.SubAccount, OnlyTestCreation bool) {
	type args struct {
		subaccount *models.SubAccount
	}
	sameaccount := models.SubAccount{
		SubAccountDB: models.SubAccountDB{
			CustomerID:  sample.CustomerID,
			AccountName: sample.AccountName,
		},
	}
	tests := []struct {
		name         string
		args         args
		wantNotEmpty bool
		wantErr      bool
	}{
		{
			name: "建立子帳號",
			args: args{
				subaccount: &sample,
			},
			wantNotEmpty: true,
			wantErr:      false,
		},
		{
			name: "建立重覆帳號",
			args: args{
				subaccount: &sameaccount,
			},
			wantNotEmpty: false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := services.UpdateSubaccount(c, *tt.args.subaccount)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateSubaccount->Create subaccount error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNotEmpty && got.DocumentRefID == "" {
				t.Errorf("UpdateSubaccount->Create subaccount = %v, want %v", got, "沒有取得DocumentRefID")
			}
		})
		if OnlyTestCreation {
			return
		}
	}
}

func GetSubAccountTEST(t *testing.T, c context.Context, sample models.SubAccount) {
	t.Run("依CustomerID取得子帳號資料", func(t *testing.T) {
		rtn, err := services.GetSubaccountListByID(c, sample.CustomerID)
		if err != nil {
			t.Errorf("GetSubaccountListByID() error = %v", err)
			return
		}
		if len(rtn) == 0 {
			t.Errorf("GetSubaccountListByID() error = %v", "子帳號資料為空")
			return
		}
	})

	t.Run("依CustomerID取得子帳號資料", func(t *testing.T) {
		rtn, err := services.GetSubaccountListByID(c, "NotRealCustomerID")
		if err != nil {
			t.Errorf("GetSubaccountListByID() error = %v", err)
			return
		}
		if len(rtn) != 0 {
			t.Errorf("GetSubaccountListByID() error = %v", "取到了不該取的子帳號資料")
			return
		}
	})
}

func UpdateSubaccountTest(t *testing.T, c context.Context, sample models.SubAccount) models.SubAccount {
	//先建立第二筆帳號資料
	secondSample := models.SubAccount{
		SubAccountDB: models.SubAccountDB{
			CustomerID:  sample.CustomerID,
			AccountName: "Secondsubaccountname",
		},
	}
	//建立第二筆資料。這邊不應有錯誤
	dbSecondSubaccount, _ := services.UpdateSubaccount(c, secondSample)
	oldName := sample.AccountName
	oldCustomerID := sample.CustomerID
	oldREFID := sample.DocumentRefID

	t.Run("更新子帳號", func(t *testing.T) {
		sample.AccountName = "udpatedAccountName"
		dbData, err := services.UpdateSubaccount(c, sample)
		if err != nil {
			t.Errorf("UpdateSubaccount() error = %v", err)
			return
		}
		if dbData.AccountName == oldName {
			t.Errorf("UpdateSubaccount() not work.")
			return
		}

		sample.CustomerID = "NotExistCustomerID"
		dbData, err = services.UpdateSubaccount(c, sample)
		if err == nil {
			t.Errorf("UpdateSubaccount() CustomerID錯誤， 預計要失敗.")
			return
		}

		sample.CustomerID = oldCustomerID
		sample.DocumentRefID = "NotExistREFID"
		dbData, err = services.UpdateSubaccount(c, sample)
		if err == nil {
			t.Errorf("UpdateSubaccount() DocumentRefID錯誤，預計要失敗.")
			return
		}

		sample.DocumentRefID = oldREFID
		//要取出第二筆資料來比對有名字有沒有被改掉。
		secondSampleDB, err := services.GetSubaccountByID(c, dbSecondSubaccount.DocumentRefID)
		if err != nil {
			t.Errorf("GetSubaccountByID() 取不到資料. error = %v", err)
			return
		}
		if secondSampleDB.AccountName != secondSample.AccountName {
			t.Errorf("Update subaccount時，誤改了其它筆")
			return
		}

	})
	return dbSecondSubaccount
}

func DeleteSubaccountTest(t *testing.T, c context.Context, subaccounts ...models.SubAccount) {
	t.Run("刪除子帳號", func(t *testing.T) {
		for _, subaccount := range subaccounts {
			err := services.DeleteSubaccount(c, subaccount)
			if err != nil {
				t.Errorf("DeleteSubaccount() error = %v", err)
			}
		}
	})
}
