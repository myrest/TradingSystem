package services

import (
	"TradingSystem/src/models"
	"context"
	"strings"

	"google.golang.org/api/iterator"
)

func GetMappedCustomerList(ctx context.Context) ([]models.CustomerMap, error) {
	var rtn []models.CustomerMap
	customermap := make(map[string]models.CustomerMap)
	client := getFirestoreClient()
	//取出所有的客戶資料
	iter := client.Collection("customers").Documents(ctx)
	var customers []models.Customer
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
			customermap[customer.ID] = models.CustomerMap{
				Parent_CustomerID: customer.ID,
			}
		}
		customers = append(customers, customer)
	}

	// 建立 parent-child 关系
	for _, customer := range customers {
		if !strings.Contains(customer.Email, "@") {
			mappedcustomer := customermap[customer.Email]
			mappedcustomer.Child_CustomerID = append(mappedcustomer.Child_CustomerID, customer.ID)
			customermap[customer.Email] = mappedcustomer
		}
	}

	for _, value := range customermap {
		rtn = append(rtn, value)
	}
	return rtn, nil
}
