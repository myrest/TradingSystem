package services

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
)

const deletion_batchSize = 500 // 每個批次的最大文件數量

func GetCustomerCurrencySymbosBySymbol(ctx context.Context, symbol string) ([]models.CustomerCurrencySymboWithCustomer, error) {
	client := common.GetFirestoreClient()

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
	client := common.GetFirestoreClient()
	doc, _, err := client.Collection("webhookData").Add(ctx, webhookData)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}

func SaveCustomerPlaceOrderResultLog(ctx context.Context, placeorderlog models.Log_TvSiginalData) (string, error) {
	client := common.GetFirestoreClient()
	doc, _, err := client.Collection("placeOrderLog").Add(ctx, placeorderlog)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}

func GetPlaceOrderHistory(ctx context.Context, Symbol, CustomerID string, sdt, edt time.Time, page, pageSize int) ([]models.Log_TvSiginalData, int, error) {
	client := common.GetFirestoreClient()

	query := client.Collection("placeOrderLog").
		Where("CustomerID", "==", CustomerID).
		Where("Symbol", "==", Symbol).
		Where("Time", ">=", common.FormatDate(sdt)).
		Where("Time", "<", common.FormatTime(edt)).
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

	totalpage, err := getTotalPages(ctx, Symbol, CustomerID, sdt, edt, pageSize)

	return rtn, totalpage, err
}

func getTotalPages(ctx context.Context, Symbol, CustomerID string, sdt, edt time.Time, pageSize int) (int, error) {
	client := common.GetFirestoreClient()

	query := client.Collection("placeOrderLog").
		Where("CustomerID", "==", CustomerID).
		Where("Symbol", "==", Symbol).
		Where("Time", ">=", common.FormatDate(sdt)).
		Where("Time", "<", common.FormatTime(edt)).
		OrderBy("Time", firestore.Desc)
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

// ClearPlaceOrderHistory 清除過期的 placeOrderLog
func ClearPlaceOrderHistory(ctx context.Context, edt time.Time) error {
	client := common.GetFirestoreClient()
	// 使用 Firestore 批量写入操作
	bulkWriter := client.BulkWriter(ctx)

	query := client.Collection("placeOrderLog").
		Where("Time", "<", common.FormatTime(edt))

	iter := query.Documents(ctx)
	defer iter.Stop()

	count := 0 // 計數器

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		bulkWriter.Delete(doc.Ref)
		count++
		if count >= deletion_batchSize {
			bulkWriter.Flush()
			count = 0 // 重置計數器
		}
	}
	bulkWriter.Flush()
	return nil
}

func ClearCustomerReportHistory(ctx context.Context, edt time.Time) error {
	//Clear weekly report
	err := clearReportHistory(ctx, common.GetWeeksByDate(edt), "CustomerWeeklyReport")
	if err != nil {
		return err
	}

	//clear monthly report
	err = clearReportHistory(ctx, common.GetMonthsInRange(edt)[0], "CustomerMonthlyReport")
	if err != nil {
		return err
	}
	return nil

}

// 清除過期的 CustomerWeeklyReport
func clearReportHistory(ctx context.Context, edt, reportName string) error {
	client := common.GetFirestoreClient()
	// 使用 Firestore 批量写入操作
	bulkWriter := client.BulkWriter(ctx)

	query := client.Collection(reportName).
		Where("YearUnit", "<", edt)

	iter := query.Documents(ctx)
	defer iter.Stop()

	count := 0 // 計數器

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		bulkWriter.Delete(doc.Ref)
		count++
		if count >= deletion_batchSize {
			bulkWriter.Flush()
			count = 0 // 重置計數器
		}
	}
	bulkWriter.Flush()
	return nil
}
