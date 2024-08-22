package services

import (
	"TradingSystem/src/models"
	"context"
	"strings"

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
		if customer.IsAdmin { //跳過管理員，管理員不算客戶
			continue
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
