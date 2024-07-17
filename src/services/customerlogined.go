package services

import (
	"TradingSystem/src/models"
	"context"
	"errors"

	"google.golang.org/api/iterator"
)

func UpdateCustomerCurrency(ctx context.Context, customercurrency *models.CustomerCurrencySymbol) error {
	client := getFirestoreClient()
	customer, err := GetCustomer(ctx, customercurrency.CustomerID)
	if err != nil || customer == nil || customer.Email == "" {
		if err == nil {
			return err
		} else {
			return errors.New("customer not found")
		}
	}
	//檢查系統symbo是否存在
	iter := client.Collection("SymbolData").Where("Symbol", "==", customercurrency.Symbol).Limit(1).Documents(ctx)
	defer iter.Stop()
	_, err = iter.Next()
	if err == iterator.Done {
		return errors.New("system Symbol (" + customercurrency.Symbol + ") not found")
	}
	if err != nil {
		return err
	}

	iter = client.Collection("customerssymbol").Where("Symbol", "==", customercurrency.Symbol).
		Where("CustomerID", "==", customercurrency.CustomerID).
		Limit(1).Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err == iterator.Done {
		// data not found
		_, _, err := client.Collection("customerssymbol").Add(ctx, customercurrency)
		return err
	}

	if err != nil {
		return err
	}

	var data models.CustomerCurrencySymbol
	doc.DataTo(&data)
	data.Status = customercurrency.Status
	data.Amount = customercurrency.Amount
	data.Simulation = customercurrency.Simulation

	_, err = client.Collection("customerssymbol").Doc(doc.Ref.ID).Set(ctx, data)
	return err
}

func GetCustomerCurrency(ctx context.Context, customerID string) ([]models.CustomerCurrencySymbol, error) {
	client := getFirestoreClient()

	iter := client.Collection("customerssymbol").Where("CustomerID", "==", customerID).Documents(ctx)
	defer iter.Stop()

	var customerCurrencySymbos []models.CustomerCurrencySymbol
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var data models.CustomerCurrencySymbol
		doc.DataTo(&data)
		customerCurrencySymbos = append(customerCurrencySymbos, data)
	}

	return customerCurrencySymbos, nil
}

func DeleteCustomerCurrency(ctx context.Context, CustomerID, Symbol string) error {
	client := getFirestoreClient()

	iter := client.Collection("customerssymbol").Where("Symbol", "==", Symbol).
		Where("CustomerID", "==", CustomerID).
		Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		// 删除文档
		_, err = doc.Ref.Delete(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
