package services

import (
	"TradingSystem/src/models"
	"context"
	"errors"
	"log"
	"strings"

	"google.golang.org/api/iterator"
)

func CreateCustomer(ctx context.Context, customer *models.Customer) (string, error) {
	client := getFirestoreClient()

	//只有Email格式的要驗Email是否存在，Email為ID型式的為subaccount
	i := strings.Index(customer.Email, "@")
	if i > -1 {
		dbCustomer, _ := GetCustomerByEmail(ctx, customer.Email)
		if dbCustomer != nil {
			return "", errors.New("account exist")
		}
	}

	doc, _, err := client.Collection("customers").Add(ctx, customer)
	if err != nil {
		return "", err
	}

	customer.ID = doc.ID
	_, err = client.Collection("customers").Doc(customer.ID).Set(ctx, customer)
	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

func GetCustomer(ctx context.Context, id string) (*models.Customer, error) {
	client := getFirestoreClient()

	doc, err := client.Collection("customers").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var customer models.Customer
	doc.DataTo(&customer)
	customer.ID = doc.Ref.ID

	return &customer, nil
}

func UpdateCustomer(ctx context.Context, customer *models.Customer) error {
	client := getFirestoreClient()

	_, err := client.Collection("customers").Doc(customer.ID).Set(ctx, customer)
	return err
}

func DeleteCustomer(ctx context.Context, id string) error {
	client := getFirestoreClient()

	_, err := client.Collection("customers").Doc(id).Delete(ctx)
	return err
}

func GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error) {
	client := getFirestoreClient()

	iter := client.Collection("customers").Where("Email", "==", email).Limit(1).Documents(ctx)
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

func GetCustomerByTgIdentifyKey(ctx context.Context, key string) (*models.Customer, error) {
	client := getFirestoreClient()

	iter := client.Collection("customers").Where("TgIdentifyKey", "==", key).Limit(1).Documents(ctx)
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

func GetCustomerByTgChatID(ctx context.Context, key int64) (*[]models.Customer, error) {
	client := getFirestoreClient()

	iter := client.Collection("customers").Where("TgChatID", "==", key).Documents(ctx)
	defer iter.Stop()

	rtn := []models.Customer{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var data models.Customer
		doc.DataTo(&data)
		rtn = append(rtn, data)
	}

	return &rtn, nil
}
