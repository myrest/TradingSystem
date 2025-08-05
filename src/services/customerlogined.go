package services

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"errors"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// 清理相關常數
const (
	DefaultBatchSize          = 500  // 預設批次大小
	DefaultProgressReportStep = 1000 // 預設進度報告間隔
	MaxFirestoreBatchSize     = 500  // Firestore 批次操作最大限制
)

// 預設會幫客戶自動修改槓桿及全倉，傳入false可以不修改，目前是固定自動改
func UpdateCustomerCurrency(ctx context.Context, customercurrency *models.CustomerCurrencySymbol, flag ...bool) error {
	//依據ExchangeSystemName來判斷使用哪一個client
	autoUpdateBingXLeverage := true
	if len(flag) > 0 {
		autoUpdateBingXLeverage = flag[0]
	}

	client := common.GetFirestoreClient()
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

	//模擬盤固定為10倍
	if customercurrency.Simulation {
		customercurrency.Leverage = 10
	}

	var data models.CustomerCurrencySymbol
	doc.DataTo(&data)
	data.Status = customercurrency.Status
	data.Amount = customercurrency.Amount
	data.Leverage = customercurrency.Leverage
	data.Simulation = customercurrency.Simulation

	_, err = client.Collection("customerssymbol").Doc(doc.Ref.ID).Set(ctx, data)

	if err != nil {
		return err
	}

	//該幣種有啟用，有自動修改，且為實盤才改設定
	if customercurrency.Status && autoUpdateBingXLeverage && !customercurrency.Simulation {
		//幫客戶改槓桿及持倉模式
		err = UpdateLeverage(ctx, customer.APIKey, customer.SecretKey, customer.ExchangeSystemName, customercurrency.Symbol, int64(customercurrency.Leverage))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetAllCustomerCurrency(ctx context.Context, customerID string) ([]models.CustomerCurrencySymbol, error) {
	client := common.GetFirestoreClient()

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

func GetCustomerCurrency(ctx context.Context, customerID, symbol string) (*models.CustomerCurrencySymbol, error) {
	client := common.GetFirestoreClient()

	iter := client.Collection("customerssymbol").Where("Symbol", "==", symbol).
		Where("CustomerID", "==", customerID).
		Limit(1).Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err == iterator.Done {
		// data not found
		return nil, nil
	}

	var data models.CustomerCurrencySymbol
	doc.DataTo(&data)

	return &data, nil
}

func CleanCustomerCurrency(ctx context.Context) error {
	return cleanCustomerCurrencyWithOptions(ctx, CleanOptions{
		BatchSize:    DefaultBatchSize,
		LogProgress:  true,
		ProgressStep: DefaultProgressReportStep,
	})
}

// CleanOptions 清理選項
type CleanOptions struct {
	BatchSize    int  // 批次處理大小
	LogProgress  bool // 是否記錄進度
	ProgressStep int  // 進度報告間隔
}

// cleanCustomerCurrencyWithOptions 帶有選項的清理函數
func cleanCustomerCurrencyWithOptions(ctx context.Context, options CleanOptions) error {
	systemSymbol, err := GetAllSymbol(ctx)
	if err != nil {
		return err
	}

	// 將 systemSymbol 中啟用的資料轉成 map
	enabledSystemSymbolMap := make(map[string]bool)
	for _, v := range systemSymbol {
		if v.Status {
			enabledSystemSymbolMap[v.Symbol] = true
		}
	}

	client := common.GetFirestoreClient()

	// 批次收集要刪除的文檔
	var docsToDelete []*firestore.DocumentRef
	var deletedCount int
	var processedCount int

	if options.LogProgress {
		log.Printf("Starting CleanCustomerCurrency and there are %d enabled in %d system symbols", len(enabledSystemSymbolMap), len(systemSymbol))
	}

	// 查詢 customerssymbol 中的所有資料
	iter := client.Collection("customerssymbol").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error reading customerssymbol document: %v", err)
			return err
		}

		processedCount++

		// 定期報告進度
		if options.LogProgress && processedCount%options.ProgressStep == 0 {
			log.Printf("Processed %d records, deleted %d records so far", processedCount, deletedCount)
		}

		var customerSymbol models.CustomerCurrencySymbol
		if err := doc.DataTo(&customerSymbol); err != nil {
			log.Printf("Error converting document data: %v", err)
			continue
		}

		// 如果 systemSymbol 中該貨幣符號仍然啟用，就跳過
		if enabledSystemSymbolMap[customerSymbol.Symbol] {
			continue
		}

		// 檢查 placeOrderLog 集合中是否存在相同的 CustomerID 和 Symbol
		exists, err := checkPlaceOrderExists(ctx, client, customerSymbol.CustomerID, customerSymbol.Symbol)
		if err != nil {
			log.Printf("Error checking placeOrder existence for customer %s, symbol %s: %v",
				customerSymbol.CustomerID, customerSymbol.Symbol, err)
			continue // 繼續處理其他記錄，而不是完全失敗
		}

		// 如果在 placeOrderLog 中沒有找到，標記為需要刪除
		if !exists {
			docsToDelete = append(docsToDelete, doc.Ref)
		}

		// 當累積到一定數量時，執行批次刪除
		if len(docsToDelete) >= options.BatchSize {
			if err := batchDeleteDocuments(ctx, client, docsToDelete); err != nil {
				log.Printf("Error in batch delete: %v", err)
				return err
			}
			deletedCount += len(docsToDelete)
			docsToDelete = docsToDelete[:0] // 清空切片
		}
	}

	// 處理剩餘的文檔
	if len(docsToDelete) > 0 {
		if err := batchDeleteDocuments(ctx, client, docsToDelete); err != nil {
			log.Printf("Error in final batch delete: %v", err)
			return err
		}
		deletedCount += len(docsToDelete)
	}

	if options.LogProgress {
		log.Printf("CleanCustomerCurrency completed: processed %d records, deleted %d records",
			processedCount, deletedCount)
	}

	return nil
}

// batchDeleteDocuments 批次刪除文檔，使用 BulkWriter
func batchDeleteDocuments(ctx context.Context, client *firestore.Client, docRefs []*firestore.DocumentRef) error {
	if len(docRefs) == 0 {
		return nil
	}

	bulkWriter := client.BulkWriter(ctx)
	defer bulkWriter.End()

	for _, docRef := range docRefs {
		_, err := bulkWriter.Delete(docRef)
		if err != nil {
			return err
		}
	}

	bulkWriter.Flush()
	return nil
}

// 檢查 placeOrderLog 中是否存在相同的 CustomerID 和 Symbol
func checkPlaceOrderExists(ctx context.Context, client *firestore.Client, customerID, symbol string) (bool, error) {
	logsymbol := common.FormatSymbol(symbol)

	// 使用 Select() 只查詢文檔 ID，減少數據傳輸
	iter := client.Collection("placeOrderLog").
		Where("CustomerID", "==", customerID).
		Where("Symbol", "==", logsymbol).
		Select(). // 只查詢文檔 ID，不查詢完整數據
		Limit(1).Documents(ctx)
	defer iter.Stop()

	_, err := iter.Next()
	if err == iterator.Done {
		// 找不到資料
		return false, nil
	}

	if err != nil {
		log.Printf("Error checking placeOrderLog for customer %s, symbol %s: %v",
			customerID, symbol, err)
		return false, err
	}

	// 找到資料
	return true, nil
}

// 給自動測試刪除使用
func DeleteCustomerCurrency(ctx context.Context, CustomerID, Symbol string) error {
	client := common.GetFirestoreClient()

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

func GetCustomerByTGChatID(ctx context.Context, ChatID int64) (*models.Customer, error) {
	client := common.GetFirestoreClient()

	iter := client.Collection("customers").Where("TgChatID", "==", ChatID).Limit(1).Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // Customer not found
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var customer models.Customer
	doc.DataTo(&customer)
	customer.ID = doc.Ref.ID
	return &customer, nil
}
