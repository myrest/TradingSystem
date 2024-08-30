package services

import (
	"TradingSystem/src/models"
	"context"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
)

func GetMappedCustomerList(ctx context.Context) (map[string]models.CustomerRelationUI, error) {
	customermap := make(map[string]models.CustomerRelationUI)
	client := getFirestoreClient()
	//取出所有的客戶資料
	iter := client.Collection("customers").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var customer models.Customer
		if err := doc.DataTo(&customer); err != nil {
			return nil, err
		}

		//先建出Parent customer資料
		if strings.Index(customer.Email, "@") > 0 {
			//因為是本尊，所以Parent不會有資料
			customermap[customer.ID] = models.CustomerRelationUI{
				Customer: customer,
			}
		} else {
			//是sub-customer，等所資料找出來，要補上Parent資料
			customermap[customer.ID] = models.CustomerRelationUI{
				Customer: customer,
			}
		}
	}

	for _, customer := range customermap {
		if !strings.Contains(customer.Customer.Email, "@") {
			//找出parent資料。
			parent := customermap[customer.Customer.Email]
			//只處理sub customer
			customermap[customer.Customer.ID] = models.CustomerRelationUI{
				Parent_CustomerID: parent.Customer.ID,
				Parent_Email:      parent.Customer.Email,
				Parent_Name:       parent.Customer.Name,
				Customer:          customer.Customer,
			}
		}
	}

	return customermap, nil
}

func getCount(ctx context.Context, client *firestore.Client, query *firestore.AggregationQuery) int64 {
	results, err := query.Get(ctx)
	if err != nil {
		return -1
	}
	count, ok := results["all"]
	if !ok {
		return -1
	}
	return count.(*firestorepb.Value).GetIntegerValue()
}

// 取得各資料表的筆數
func GetCustomerData(ctx context.Context, customerID string) map[string]int64 {
	rtn := make(map[string]int64)
	client := getFirestoreClient()

	//placeOrderLog
	query := client.Collection("placeOrderLog").
		Where("CustomerID", "==", customerID)
	aggregationQuery := query.NewAggregationQuery().WithCount("all") //一定要用all，因為取得的function是用all當key
	rtn["placeOrderLog"] = getCount(ctx, client, aggregationQuery)

	//customerssymbol
	query = client.Collection("customerssymbol").
		Where("CustomerID", "==", customerID)
	aggregationQuery = query.NewAggregationQuery().WithCount("all")
	rtn["customerssymbol"] = getCount(ctx, client, aggregationQuery)

	//DBCustomerWeeklyReport
	query = client.Collection("DBCustomerWeeklyReport").
		Where("CustomerID", "==", customerID)
	aggregationQuery = query.NewAggregationQuery().WithCount("all")
	rtn["DBCustomerWeeklyReport"] = getCount(ctx, client, aggregationQuery)

	return rtn

}

// 刪除資料
func DeleteCustomerData(ctx context.Context, customerID string, cutoffDatetime time.Time) error {
	client := getFirestoreClient()
	query := client.Collection("placeOrderLog").
		Where("Time", "<", cutoffDatetime).
		Where("CustomerID", "==", customerID)

	iter := query.Documents(ctx)
	defer iter.Stop()

	// 創建 BulkWriter
	bulkWriter := client.BulkWriter(ctx)
	defer bulkWriter.Flush() // 確保在結束時將所有操作提交

	// 遍歷查詢結果
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break // 沒有更多文件
		}
		if err != nil {
			return err
		}

		// 在 BulkWriter 中添加刪除操作
		if _, err := bulkWriter.Delete(doc.Ref); err != nil {
			return err
		}
	}
	// 提交所有操作
	bulkWriter.Flush()
	return nil
}
