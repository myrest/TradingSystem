package services

import (
	"TradingSystem/src/models"
	"context"
	"errors"
	"log"

	"google.golang.org/api/iterator"
)

func CreateCustomer(ctx context.Context, customer *models.Customer) (string, error) {
	client := getFirestoreClient()
	dbCustomer, _ := GetCustomerByEmail(ctx, customer.Email)
	if dbCustomer != nil {
		return "", errors.New("account exist")
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

func CreateCustomerAccountxxxxx(customer *models.Customer) error {
	ctx := context.Background()
	client := getFirestoreClient()

	_, _, err := client.Collection("customers").Add(ctx, customer)
	return err
}
