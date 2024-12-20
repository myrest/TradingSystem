package services

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

const (
	SubAccountColliectionName = "SubAccounts"
)

func GetSubaccountListByID(ctx context.Context, CustomerID string) ([]models.SubAccount, error) {
	var rtn []models.SubAccount
	client := common.GetFirestoreClient()
	//找出所有的subaccount
	iter := client.Collection(SubAccountColliectionName).
		Where("CustomerID", "==", CustomerID).
		Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var data models.SubAccount
		doc.DataTo(&data)
		data.DocumentRefID = doc.Ref.ID
		rtn = append(rtn, data)
	}

	return rtn, nil
}

func getSubaccountByaccountname(ctx context.Context, CustomerID, SubAccountName string) (*models.SubAccount, error) {
	var dbData models.SubAccountDB
	client := common.GetFirestoreClient()
	iter := client.Collection(SubAccountColliectionName).
		Where("CustomerID", "==", CustomerID).
		Where("AccountName", "==", SubAccountName).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	doc.DataTo(&dbData)
	rtn := models.SubAccount{
		SubAccountDB:  dbData,
		DocumentRefID: doc.Ref.ID,
	}
	return &rtn, nil
}

func GetSubaccountByID(ctx context.Context, ID string) (*models.SubAccount, error) {
	var dbData models.SubAccountDB
	client := common.GetFirestoreClient()
	iter := client.Collection(SubAccountColliectionName).Doc(ID).Snapshots(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err != nil {
		return nil, err
	}
	doc.DataTo(&dbData)
	rtn := models.SubAccount{
		SubAccountDB:  dbData,
		DocumentRefID: doc.Ref.ID,
	}
	return &rtn, nil
}

func UpdateSubaccount(ctx context.Context, Data models.SubAccount) (models.SubAccount, error) {
	var rtn models.SubAccount
	if Data.CustomerID == "" {
		return rtn, errors.New("owner Customer ID is empty")
	}
	client := common.GetFirestoreClient()
	if Data.DocumentRefID == "" {
		//無Ref ID表示新增
		//檢查有無重覆
		subAccount, err := getSubaccountByaccountname(ctx, Data.CustomerID, Data.AccountName)
		if err != nil {
			return rtn, err
		}

		if subAccount != nil {
			return rtn, errors.New("accout name is duplicate")
		}
		err = client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
			//建立Customer Data
			customer := models.Customer{
				Name:  Data.AccountName,
				Email: Data.CustomerID,
			}
			customerREFID, err := CreateCustomer(ctx, &customer)

			if err != nil {
				return err
			}
			Data.SubCustomerID = customerREFID
			ref, _, err := client.Collection(SubAccountColliectionName).Add(ctx, Data.SubAccountDB)
			if err != nil {
				return err
			}

			Data.DocumentRefID = ref.ID
			return nil
		})

		if err != nil {
			return rtn, err
		}

		return Data, nil
	} else {
		//有Ref ID表示修改，只確定資料是否存在
		dbSubAccount, err := GetSubaccountByID(ctx, Data.DocumentRefID)
		if err != nil {
			return rtn, err
		}

		if dbSubAccount.CustomerID == "" {
			return rtn, errors.New("account REF ID not exist")
		}

		if dbSubAccount.CustomerID != Data.CustomerID {
			return rtn, errors.New("you do not have the permission to update")
		}

		_, err = client.Collection(SubAccountColliectionName).Doc(Data.DocumentRefID).Set(ctx, Data.SubAccountDB)
		if err != nil {
			return rtn, err
		}
		return Data, nil
	}
}

func DeleteSubaccount(ctx context.Context, Data models.SubAccount) error {
	if Data.CustomerID == "" {
		return errors.New("owner Customer ID is empty")
	}
	client := common.GetFirestoreClient()

	//有Ref ID表示修改，只確定資料是否存在
	dbSubAccount, err := GetSubaccountByID(ctx, Data.DocumentRefID)
	if err != nil {
		return err
	}

	if dbSubAccount.DocumentRefID == "" {
		return errors.New("account REF ID not exist")
	}

	if dbSubAccount.CustomerID != Data.CustomerID {
		return errors.New("you do not have the permission to the deletion")
	}

	_, err = client.Collection(SubAccountColliectionName).Doc(Data.DocumentRefID).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func SwitchToSubAccount(c *gin.Context, SubAccountREFID string) error {
	//取得當下的CustomerID
	session := sessions.Default(c)
	var ParentID string
	IparentCustomerID := session.Get("parentid")
	if IparentCustomerID != nil {
		ParentID = IparentCustomerID.(string)
	}
	contextC := context.Background()

	//檢查SubAccountREFID是否存在
	subaccount, err := GetSubaccountByID(contextC, SubAccountREFID)
	if err != nil {
		return err
	}
	//檢查是否能Switch
	if subaccount.CustomerID != ParentID {
		return errors.New("you do not have permission to switch to")
	}

	session.Set("id", subaccount.SubCustomerID)
	session.Set("name", subaccount.AccountName)
	session.Set("isadmin", false)
	if err := session.Save(); err != nil {
		return errors.New("failed to save session")
	}

	return nil
}

func SwitchToMainAccount(c *gin.Context) error {
	//取得當下的CustomerID
	session := sessions.Default(c)
	var ParentID string
	IparentCustomerID := session.Get("parentid")
	if IparentCustomerID != nil {
		ParentID = IparentCustomerID.(string)
	}
	contextC := context.Background()

	account, err := GetCustomer(contextC, ParentID)
	if err != nil {
		return err
	}

	session.Set("id", account.ID)
	session.Set("name", account.Name)
	session.Set("isadmin", account.IsAdmin)
	if err := session.Save(); err != nil {
		return errors.New("failed to save session")
	}

	return nil
}
