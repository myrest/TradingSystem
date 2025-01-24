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
	systemSymbol, err := GetAllSymbol(ctx)
	if err != nil {
		return err
	}

	//將systemSymbol中的資料轉成map
	systemSymbolMap := make(map[string]bool)
	for _, v := range systemSymbol {
		systemSymbolMap[v.Symbol] = v.Status
	}

	// 查詢 customerssymbol 中的所有資料
	client := common.GetFirestoreClient()
	iter := client.Collection("customerssymbol").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		var customerSymbol models.CustomerCurrencySymbol
		doc.DataTo(&customerSymbol)
		//如果systemSymbolMap中存在customerSymbol.Symbol，且customerSymbol.Status為true，就跳過
		if systemSymbolMap[customerSymbol.Symbol] {
			continue
		}

		// 檢查 placeOrder 集合中是否存在相同的 CustomerID 和 Symbol
		exists, err := checkPlaceOrderExists(ctx, client, customerSymbol.CustomerID, customerSymbol.Symbol)
		if err != nil {
			return err
		}

		// 如果在 placeOrder 中沒有找到，就刪除 customerssymbol 裡的資料
		if !exists {
			if _, err := client.Collection("customerssymbol").Doc(doc.Ref.ID).Delete(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// 檢查 placeOrder 中是否存在相同的 CustomerID 和 Symbol
func checkPlaceOrderExists(ctx context.Context, client *firestore.Client, customerID, symbol string) (bool, error) {
	logsymbol := common.FormatSymbol(symbol)
	iter := client.Collection("placeOrderLog").
		Where("CustomerID", "==", customerID).
		Where("Symbol", "==", logsymbol).
		Limit(1).Documents(ctx)
	defer iter.Stop()

	_, err := iter.Next()
	if err == iterator.Done {
		// 找不到資料
		return false, nil
	}

	if err != nil {
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
