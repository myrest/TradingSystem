package services

import (
	"TradingSystem/src/models"
	"context"

	"google.golang.org/api/iterator"
)

func GetCustomerCurrencySymbosBySymbol(ctx context.Context, symbol string) ([]models.CustomerCurrencySymboWithCustomer, error) {
	client := getFirestoreClient()

	// 查询 CustomerCurrencySymbol 集合
	iter := client.Collection("customerssymbo").Where("Symbol", "==", symbol).Where("Status", "==", true).Documents(ctx)
	defer iter.Stop()

	var customerCurrencySymbos []models.CustomerCurrencySymbol
	var customerIDs []string

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var customerCurrencySymbol models.CustomerCurrencySymbol
		doc.DataTo(&customerCurrencySymbol)
		customerCurrencySymbos = append(customerCurrencySymbos, customerCurrencySymbol)
		customerIDs = append(customerIDs, customerCurrencySymbol.CustomerID)
	}

	// 分批次查询 Customer 记录
	customers := make(map[string]models.Customer)
	batchSize := 10
	for i := 0; i < len(customerIDs); i += batchSize {
		end := i + batchSize
		if end > len(customerIDs) {
			end = len(customerIDs)
		}
		batchIDs := customerIDs[i:end]

		customerIter := client.Collection("customers").Where("ID", "in", batchIDs).Documents(ctx)
		defer customerIter.Stop()

		for {
			doc, err := customerIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, err
			}

			var customer models.Customer
			doc.DataTo(&customer)
			customer.ID = doc.Ref.ID
			customers[customer.ID] = customer
		}
	}

	// 合并结果
	var results []models.CustomerCurrencySymboWithCustomer
	for _, ccs := range customerCurrencySymbos {
		if customer, found := customers[ccs.CustomerID]; found {
			result := models.CustomerCurrencySymboWithCustomer{
				CustomerCurrencySymbol: ccs,
				Customer:               customer,
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// SaveWebhookData saves the webhook data to Firestore
func SaveWebhookData(ctx context.Context, webhookData models.TvWebhookData) error {
	client := getFirestoreClient()
	_, _, err := client.Collection("webhookData").Add(ctx, webhookData)
	return err
}

func SaveCustomerPlaceOrderResultLog(ctx context.Context, placeorderlog models.Log_TvSiginalData) error {
	client := getFirestoreClient()
	_, _, err := client.Collection("placeOrderLog").Add(ctx, placeorderlog)
	return err
}
