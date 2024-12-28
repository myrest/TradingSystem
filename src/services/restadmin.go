package services

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"errors"
	"log"
	"sort"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const (
	ErrorCode_001 = "symbol not found"
	ErrorCode_002 = "incorrect Symbol cert"
)

func CreateNewSymbol(ctx context.Context, Symbol models.AdminCurrencySymbol) (*models.AdminCurrencySymbol, error) {
	client := common.GetFirestoreClient()
	Symbol.Cert = common.GenerateRandomString(8)

	//檢查有無重覆
	if _, err := GetSymbol(ctx, Symbol.Symbol, Symbol.Cert); err != nil && err.Error() != ErrorCode_001 {
		return nil, err
	}

	_, _, err := client.Collection("SymbolData").Add(ctx, Symbol)
	if err != nil {
		return nil, err
	}

	if systemSymbol, err := GetSymbol(ctx, Symbol.Symbol, Symbol.Cert); err != nil {
		return nil, err
	} else {
		//處理自動訂閱
		err = updateAutoSubscriberCustomerSymbol(ctx, *systemSymbol)
		if err != nil {
			log.Printf("updateAutoSubscriberCustomerSymbol got error=%v", err)
		}
		return systemSymbol, nil
	}
}

// 只有在Create的時候才幫有自動訂閱的客戶加上去。
func updateAutoSubscriberCustomerSymbol(ctx context.Context, AdminSymbol models.AdminCurrencySymbol) error {
	client := common.GetFirestoreClient()

	//找出有啟用自動訂閱的客戶
	iter := client.Collection("customers").Where("IsAutoSubscribe", "==", true).
		Documents(ctx)
	defer iter.Stop()

	var dbCustomer models.Customer

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		doc.DataTo(&dbCustomer)
		data := models.CustomerCurrencySymbol{
			CurrencySymbolBase: models.CurrencySymbolBase{
				Symbol: AdminSymbol.CurrencySymbolBase.Symbol,
				Status: true,
			},
			Simulation: !dbCustomer.AutoSubscribReal,
			CustomerID: dbCustomer.ID,
			Amount:     float64(dbCustomer.AutoSubscribAmount),
			Leverage:   10, //預設10倍槓桿
		}
		//第三個參數表示不改使用者原槓桿
		err = UpdateCustomerCurrency(ctx, &data, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteAdminSymbol(ctx context.Context, Symbol string) error {
	client := common.GetFirestoreClient()

	iter := client.Collection("SymbolData").Where("Symbol", "==", Symbol).Documents(ctx)
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

func DisableCustomerSymbolStatus(ctx context.Context, Symbol string) error {
	client := common.GetFirestoreClient()

	// 使用 Firestore 批量写入操作
	bulkWriter := client.BulkWriter(ctx)

	iter := client.Collection("customerssymbol").Where("Symbol", "==", Symbol).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		updates := []firestore.Update{}
		updates = append(updates, firestore.Update{ // Append update for each doc
			Path:  "Status",
			Value: false,
		})
		bulkWriter.Update(doc.Ref, updates)
	}
	bulkWriter.Flush()
	return nil
}

func getSymbolFromDB(ctx context.Context, symbol string) (*firestore.DocumentSnapshot, error) {
	client := common.GetFirestoreClient()

	iter := client.Collection("SymbolData").Where("Symbol", "==", symbol).Limit(1).Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err != nil {
		return nil, err
	}
	return doc, err
}

func UpdateSymbolStatus(ctx context.Context, Symbol models.AdminCurrencySymbol) error {
	client := common.GetFirestoreClient()

	doc, err := getSymbolFromDB(ctx, Symbol.Symbol)
	if err != nil {
		return err
	}

	var SymbolData models.AdminCurrencySymbol
	doc.DataTo(&SymbolData)
	SymbolData.Status = Symbol.Status
	_, err = client.Collection("SymbolData").Doc(doc.Ref.ID).Set(ctx, SymbolData)

	return err
}

func UpdateSymbolMessage(ctx context.Context, Symbol models.AdminCurrencySymbol) error {
	client := common.GetFirestoreClient()

	doc, err := getSymbolFromDB(ctx, Symbol.Symbol)
	if err != nil {
		return err
	}

	var SymbolData models.AdminCurrencySymbol
	doc.DataTo(&SymbolData)
	SymbolData.Message = Symbol.Message
	_, err = client.Collection("SymbolData").Doc(doc.Ref.ID).Set(ctx, SymbolData)

	return err
}

func GetAllSymbol(ctx context.Context) ([]models.AdminCurrencySymbol, error) {
	client := common.GetFirestoreClient()
	iter := client.Collection("SymbolData").Documents(ctx)
	defer iter.Stop()

	var symboList []models.AdminCurrencySymbol
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var Symbol models.AdminCurrencySymbol
		doc.DataTo(&Symbol)
		symboList = append(symboList, Symbol)
	}

	// Sort customersymboList by Symbo
	sort.Slice(symboList, func(i, j int) bool {
		return symboList[i].Symbol < symboList[j].Symbol
	})

	return symboList, nil
}

func GetSymbol(ctx context.Context, Symbol, Cert string) (*models.AdminCurrencySymbol, error) {
	client := common.GetFirestoreClient()
	var rtn *models.AdminCurrencySymbol
	iter := client.Collection("SymbolData").Where("Symbol", "==", Symbol).Limit(1).Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, errors.New(ErrorCode_001)
	}

	doc.DataTo(&rtn)
	if rtn.Cert != Cert {
		return rtn, errors.New(ErrorCode_002)
	}

	return rtn, nil
}

func GetSubscribeCustomersBySymbol(ctx context.Context, Symbol string) ([]models.CustomerCurrencySymbolUI, error) {
	client := common.GetFirestoreClient()

	MappedSubCustomerList, err := GetMappedCustomerList(ctx)
	if err != nil {
		return nil, err
	}

	iter := client.Collection("customerssymbol").Where("Symbol", "==", Symbol).Documents(ctx)
	defer iter.Stop()

	var rtn []models.CustomerCurrencySymbolUI
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
		rtn = append(rtn, models.CustomerCurrencySymbolUI{
			CustomerCurrencySymbol: data,
			CustomerRelationUI:     MappedSubCustomerList[data.CustomerID],
		})
	}

	return rtn, nil
}

func GetCustomerIDByBingxOrderID(ctx context.Context, OrderID string) (string, error) {
	client := common.GetFirestoreClient()
	var data models.Log_TvSiginalData
	iter := client.Collection("placeOrderLog").Where("Result", "==", OrderID).Limit(1).Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err == iterator.Done {
		return "", errors.New("OrderID not found")
	}

	doc.DataTo(&data)

	return data.CustomerID, nil
}
