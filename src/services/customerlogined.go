package services

import (
	"TradingSystem/src/models"
	"context"
	"errors"

	"google.golang.org/api/iterator"
)

func UpdateCustomerCurrency(ctx context.Context, customercurrency *models.CustomerCurrencySymbo) error {
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
	iter := client.Collection("SymboData").Where("Symbo", "==", customercurrency.Symbo).Limit(1).Documents(ctx)
	_, err = iter.Next()
	if err == iterator.Done {
		return errors.New("system symbo (" + customercurrency.Symbo + ") not found")
	}
	if err != nil {
		return err
	}

	iter = client.Collection("customerssymbo").Where("Symbo", "==", customercurrency.Symbo).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		// data not found
		_, _, err := client.Collection("customerssymbo").Add(ctx, customercurrency)
		return err
	}

	if err != nil {
		return err
	}

	var data models.CustomerCurrencySymbo
	doc.DataTo(&data)
	data.Status = customercurrency.Status
	data.Amount = customercurrency.Amount

	_, err = client.Collection("customerssymbo").Doc(doc.Ref.ID).Set(ctx, data)
	return err
}

func GetCustomerCurrency(ctx context.Context, customerID string) ([]models.CustomerCurrencySymbo, error) {
	client := getFirestoreClient()

	iter := client.Collection("customerssymbo").Where("CustomerID", "==", customerID).Documents(ctx)
	defer iter.Stop()

	var customerCurrencySymbos []models.CustomerCurrencySymbo
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var data models.CustomerCurrencySymbo
		doc.DataTo(&data)
		customerCurrencySymbos = append(customerCurrencySymbos, data)
	}

	return customerCurrencySymbos, nil
}
