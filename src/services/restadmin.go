package services

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"errors"
	"log"
	"sort"
	"sync"

	"google.golang.org/api/iterator"
)

func CreateNewSymbol(ctx context.Context, Symbol models.CurrencySymbol) (models.CurrencySymbol, error) {
	client := getFirestoreClient()
	Symbol.Cert = common.GenerateRandomString(8)
	_, _, err := client.Collection("SymbolData").Add(ctx, Symbol)
	return Symbol, err
}

func UpdateSymbol(ctx context.Context, Symbol models.CurrencySymbol) error {
	client := getFirestoreClient()

	iter := client.Collection("SymbolData").Where("Symbol", "==", Symbol.Symbol).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err != nil {
		return err
	}

	var SymbolData models.CurrencySymbol
	doc.DataTo(&SymbolData)
	SymbolData.Status = Symbol.Status
	if SymbolData.Cert == "" {
		SymbolData.Cert = common.GenerateRandomString(8)
	}
	_, err = client.Collection("SymbolData").Doc(doc.Ref.ID).Set(ctx, SymbolData)

	return err
}

func GetAllSymbol(ctx context.Context) ([]models.CurrencySymbol, error) {
	client := getFirestoreClient()
	iter := client.Collection("SymbolData").Documents(ctx)

	var symboList []models.CurrencySymbol
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var Symbol models.CurrencySymbol
		doc.DataTo(&Symbol)
		symboList = append(symboList, Symbol)
	}

	// Sort customersymboList by Symbo
	sort.Slice(symboList, func(i, j int) bool {
		return symboList[i].Symbol < symboList[j].Symbol
	})

	return symboList, nil
}

func GetSymbol(ctx context.Context, Symbol, Cert string) (models.CurrencySymbol, error) {
	client := getFirestoreClient()
	var rtn models.CurrencySymbol
	iter := client.Collection("SymbolData").Where("Symbol", "==", Symbol).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return rtn, errors.New("Symbol not found")
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
