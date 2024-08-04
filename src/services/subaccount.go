package services

import (
	"TradingSystem/src/models"
	"context"
	"errors"

	"google.golang.org/api/iterator"
)

const (
	SubAccountColliectionName = "SubAccounts"
)

func GetSubaccountListByID(ctx context.Context, CustomerID string) ([]models.SubAccount, error) {
	var rtn []models.SubAccount
	client := getFirestoreClient()
	//找出所有的subaccount
	iter := client.Collection(SubAccountColliectionName).
		Where("CustomerID", "==", CustomerID).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()

	if err == iterator.Done {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var data models.SubAccount
	doc.DataTo(&data)
	data.DocumentRefID = doc.Ref.ID
	rtn = append(rtn, data)

	return rtn, nil
}

func getSubaccountByaccountname(ctx context.Context, CustomerID, SubAccountName string) (*models.SubAccount, error) {
	var dbData models.SubAccountDB
	client := getFirestoreClient()
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
	client := getFirestoreClient()
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
	client := getFirestoreClient()
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
		ref, _, err := client.Collection(SubAccountColliectionName).Add(ctx, Data.SubAccountDB)
		if err != nil {
			return rtn, err
		}

		Data.DocumentRefID = ref.ID

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
	client := getFirestoreClient()

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
