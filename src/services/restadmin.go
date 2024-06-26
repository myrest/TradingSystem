package services

import (
	"TradingSystem/src/models"
	"context"
	"sort"

	"google.golang.org/api/iterator"
)

func CreateNewSymbo(ctx context.Context, Symbo models.CurrencySymbo) error {
	client := getFirestoreClient()

	_, _, err := client.Collection("SymboData").Add(ctx, Symbo)
	return err
}

func UpdateSymbo(ctx context.Context, Symbo models.CurrencySymbo) error {
	client := getFirestoreClient()

	iter := client.Collection("SymboData").Where("Symbo", "==", Symbo.Symbo).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err != nil {
		return err
	}

	var Symbodata models.CurrencySymbo
	doc.DataTo(&Symbodata)
	Symbodata.Status = Symbo.Status
	_, err = client.Collection("SymboData").Doc(doc.Ref.ID).Set(ctx, Symbodata)

	return err
}

func GetAllSymbo(ctx context.Context) ([]models.CurrencySymbo, error) {
	client := getFirestoreClient()
	iter := client.Collection("SymboData").Documents(ctx)

	var symboList []models.CurrencySymbo
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var symbo models.CurrencySymbo
		doc.DataTo(&symbo)
		symboList = append(symboList, symbo)
	}

	// Sort customersymboList by Symbo
	sort.Slice(symboList, func(i, j int) bool {
		return symboList[i].Symbo < symboList[j].Symbo
	})

	return symboList, nil
}
