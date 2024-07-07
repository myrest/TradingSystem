package services

import (
	"TradingSystem/src/models"
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
)

func GetCustomerCurrencySymbosBySymbol(ctx context.Context, symbol string) ([]models.CustomerCurrencySymboWithCustomer, error) {
	client := getFirestoreClient()

	// 查询 CustomerCurrencySymbol 集合
	iter := client.Collection("customerssymbol").Where("Symbol", "==", symbol).Where("Status", "==", true).Documents(ctx)
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
func SaveWebhookData(ctx context.Context, webhookData models.TvWebhookData) (string, error) {
	client := getFirestoreClient()
	doc, _, err := client.Collection("webhookData").Add(ctx, webhookData)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}

func SaveCustomerPlaceOrderResultLog(ctx context.Context, placeorderlog models.Log_TvSiginalData) (string, error) {
	client := getFirestoreClient()
	doc, _, err := client.Collection("placeOrderLog").Add(ctx, placeorderlog)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}

func GetPlaceOrderHistory(ctx context.Context, Symbol, CustomerID string, page, pageSize int) ([]models.Log_TvSiginalData, int, error) {
	client := getFirestoreClient()

	query := client.Collection("placeOrderLog").
		Where("Symbol", "==", Symbol).
		Where("CustomerID", "==", CustomerID).
		OrderBy("Time", firestore.Desc).
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var rtn []models.Log_TvSiginalData

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return rtn, 0, err
		}

		var history models.Log_TvSiginalData
		doc.DataTo(&history)
		rtn = append(rtn, history)
	}

	totalpage, err := getTotalPages(ctx, Symbol, CustomerID, pageSize)

	return rtn, totalpage, err
}

func getTotalPages(ctx context.Context, Symbol, CustomerID string, pageSize int) (int, error) {
	client := getFirestoreClient()

	// Firestore COUNT query
	query := client.Collection("placeOrderLog").
		Where("Symbol", "==", Symbol).
		Where("CustomerID", "==", CustomerID)

	countQuery := query.NewAggregationQuery().WithCount("all")
	results, err := countQuery.Get(ctx)
	if err != nil {
		return 0, err
	}
	count, ok := results["all"]
	if !ok {
		return 0, errors.New("firestore: couldn't get alias for COUNT from results")
	}

	countValue := count.(*firestorepb.Value).GetIntegerValue()

	totalPages := (int(countValue) + pageSize - 1) / pageSize

	return totalPages, nil
}
