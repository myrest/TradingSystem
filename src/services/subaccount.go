package services

import (
	"TradingSystem/src/models"
	"context"

	"google.golang.org/api/iterator"
)

func GetSubaccountListByID(ctx context.Context, CustomerID string) ([]models.SubAccount, error) {
	var rtn []models.SubAccount
	client := getFirestoreClient()
	//找出所有的subaccount
	iter := client.Collection("subaccount").
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
	rtn = append(rtn, data)

	return rtn, nil
}
