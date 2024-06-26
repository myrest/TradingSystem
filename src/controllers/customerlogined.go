package controllers

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type updateCustomerSymboRequest struct {
	Symbo  string `json:"symbo"`
	Status string `json:"status"`
	Amount string `json:"amount"`
}

func UpdateCustomerSymbo(c *gin.Context) {
	var input models.CustomerCurrencySymbo
	var req updateCustomerSymboRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	input.Symbo = req.Symbo
	input.Status = req.Status == "true"
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	} else {
		input.Amount = amount
	}

	session := sessions.Default(c)
	input.CustomerID = session.Get("id").(string)
	APIkey := session.Get("apikey").(string)
	SecretKey := session.Get("secertkey").(string)

	//檢查餘額
	freeamount, err := services.GetAccountBalance(APIkey, SecretKey)

	var errormessage string

	if err != nil || input.Amount > freeamount {
		if err != nil {
			errormessage = err.Error()
		} else {
			errormessage = "Balance not enough. Balance: " + strconv.FormatFloat(freeamount, 'f', -1, 64)
		}
	}

	err = services.UpdateCustomerCurrency(context.Background(), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Update customer symbo failed. " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, errormessage)
}

func GetAllCustomerSymbo(c *gin.Context) {
	session := sessions.Default(c)
	customerid := session.Get("id").(string)

	symboList, err := services.GetAllSymbo(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	customersymboList, err := services.GetCustomerCurrency(context.Background(), customerid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mergedList := mergeSymboLists(symboList, customersymboList)

	c.JSON(http.StatusOK, mergedList)
}

func mergeSymboLists(symboList []models.CurrencySymbo, customersymboList []models.CustomerCurrencySymbo) []models.CustomerCurrencySymbo {
	customerSymboMap := make(map[string]*models.CustomerCurrencySymbo)
	for i := range customersymboList {
		customerSymboMap[customersymboList[i].Symbo] = &customersymboList[i]
	}

	// Iterate through symboList and add to customersymboList if not already present
	for _, symbo := range symboList {
		if _, exists := customerSymboMap[symbo.Symbo]; !exists {
			symbo.Status = false //預設為不啟用，不能被系統的啟用影響
			newCustomerSymbo := models.CustomerCurrencySymbo{
				CurrencySymbo: symbo,
				Amount:        0,
			}
			customersymboList = append(customersymboList, newCustomerSymbo)
		}
	}
	// Sort customersymboList by Symbo
	sort.Slice(customersymboList, func(i, j int) bool {
		return customersymboList[i].Symbo < customersymboList[j].Symbo
	})

	return customersymboList
}
