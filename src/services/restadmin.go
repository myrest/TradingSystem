package services

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"errors"
	"log"
	"sort"
	"sync"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func CreateNewSymbol(ctx context.Context, Symbol models.AdminCurrencySymbol) (models.AdminCurrencySymbol, error) {
	client := getFirestoreClient()
	Symbol.Cert = common.GenerateRandomString(8)
	_, _, err := client.Collection("SymbolData").Add(ctx, Symbol)
	return Symbol, err
}

func getSymbolFromDB(ctx context.Context, symbol string) (*firestore.DocumentSnapshot, error) {
	client := getFirestoreClient()

	iter := client.Collection("SymbolData").Where("Symbol", "==", symbol).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err != nil {
		return nil, err
	}
	return doc, err
}

func UpdateSymbolStatus(ctx context.Context, Symbol models.AdminCurrencySymbol) error {
	client := getFirestoreClient()

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
	client := getFirestoreClient()

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
	client := getFirestoreClient()
	iter := client.Collection("SymbolData").Documents(ctx)

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

func GetSymbol(ctx context.Context, Symbol, Cert string) (models.AdminCurrencySymbol, error) {
	client := getFirestoreClient()
	var rtn models.AdminCurrencySymbol
	iter := client.Collection("SymbolData").Where("Symbol", "==", Symbol).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return rtn, errors.New("symbol not found")
	}

	doc.DataTo(&rtn)
	if rtn.Cert != Cert {
		return rtn, errors.New("incorrect Symbol cert")
	}

	return rtn, nil
}

func GetLatestWebhook(ctx context.Context) ([]models.TvWebhookData, error) {
	client := getFirestoreClient()

	allAdminSymbol, err := GetAllSymbol(ctx)
	if err != nil {
		return nil, err
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	var rtn []models.TvWebhookData
	symboMap := make(map[string]models.TvWebhookData)

	for i := range allAdminSymbol {
		wg.Add(1)
		go func(Symbol string) {
			defer wg.Done()
			var webhookdata models.TvWebhookData
			webhookdata.Symbol = Symbol
			iter := client.Collection("webhookData").
				Where("Symbol", "==", Symbol).
				//OrderBy("Time", firestore.Desc).
				Limit(1).
				Documents(ctx)

			doc, err := iter.Next()
			if err == iterator.Done {
				mu.Lock()
				symboMap[Symbol] = webhookdata
				mu.Unlock()
				return
			}
			if err != nil {
				log.Printf("Failed to iterate documents for Symbol %s: %v", Symbol, err)
				mu.Lock()
				symboMap[Symbol] = webhookdata
				mu.Unlock()
				return
			}

			doc.DataTo(&webhookdata)

			mu.Lock()
			symboMap[Symbol] = webhookdata
			mu.Unlock()
		}(allAdminSymbol[i].Symbol)
	}

	wg.Wait()

	// 将map转换为slice
	for _, value := range symboMap {
		rtn = append(rtn, value)
	}

	if len(rtn) == 0 {
		return nil, errors.New("no data found")
	}

	return rtn, nil
}
