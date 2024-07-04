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
	Symbol     string `json:"symbol"`
	Status     string `json:"status"`
	Amount     string `json:"amount"`
	Simulation string `json:"simulation"`
}

func ShowDashboardPage(c *gin.Context) {
	session := sessions.Default(c)
	name := session.Get("name")
	email := session.Get("email")

	if name == nil || email == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	//todo:新建好像有點問題？
	customer, err := services.GetCustomerByEmail(email.(string))
	if err == nil {
		if customer == nil {
			//帳號不存在，要建立一個新
			c.HTML(http.StatusOK, "iscreatenew.html", gin.H{
				"Name":  name,
				"Email": email,
			})
		} else {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Name":      name,
				"Email":     email,
				"ApiKey":    customer.APIKey,
				"SecretKey": customer.SecretKey,
				"IsAdmin":   customer.IsAdmin,
			})
		}
	} else {
		c.Redirect(http.StatusFound, "/login?GotError")
	}
}

func UpdateCustomerSymbol(c *gin.Context) {
	var input models.CustomerCurrencySymbol
	var req updateCustomerSymboRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	input = models.CustomerCurrencySymbol{
		CurrencySymbolBase: models.CurrencySymbolBase{
			Symbol: req.Symbol,
			Status: req.Status == "true",
		},
		Simulation: req.Simulation == "true",
	}
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	} else {
		input.Amount = amount
	}

	session := sessions.Default(c)
	input.CustomerID = session.Get("id").(string)
	var APIkey, SecretKey string

	iAPIkey := session.Get("apikey")
	iSecretKey := session.Get("secertkey")
	if iAPIkey != nil {
		APIkey = iAPIkey.(string)
	}
	if iSecretKey != nil {
		SecretKey = iSecretKey.(string)
	}

	var errormessage string

	//有key，啟用時要檢查餘額
	var freeamount float64
	if APIkey != "" && SecretKey != "" && input.Status {
		freeamount, err = services.GetAccountBalance(APIkey, SecretKey)
		if err != nil || input.Amount > freeamount {
			if err != nil {
				errormessage = err.Error()
			} else {
				errormessage = "Balance not enough. Balance: " + strconv.FormatFloat(freeamount, 'f', -1, 64)
			}
		}
	}

	err = services.UpdateCustomerCurrency(context.Background(), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Update customer Symbol failed. " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, errormessage)
}

func GetAllCustomerSymbol(c *gin.Context) {
	session := sessions.Default(c)
	customerid := session.Get("id").(string)

	systemSymboList, err := services.GetAllSymbol(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	customersymboList, err := services.GetCustomerCurrency(context.Background(), customerid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mergedList := mergeSymboLists(systemSymboList, customersymboList)

	c.JSON(http.StatusOK, mergedList)
}

func mergeSymboLists(systemSymboList []models.AdminCurrencySymbol, customersymboList []models.CustomerCurrencySymbol) []models.CustomerCurrencySymboResponse {
	customerSymboMap := make(map[string]models.CustomerCurrencySymbol)
	for _, Symbol := range customersymboList {
		customerSymboMap[Symbol.Symbol] = Symbol
	}

	var result []models.CustomerCurrencySymboResponse

	// Iterate through systemSymboList
	for _, Symbol := range systemSymboList {
		systemStatus := "Disabled"
		if Symbol.Status {
			systemStatus = "Enabled"
		}
		if customerSymbol, exists := customerSymboMap[Symbol.Symbol]; exists {
			// 如果 systemSymboList 中的 Symbol 存在于 customerSymboMap 中
			result = append(result, models.CustomerCurrencySymboResponse{
				CurrencySymbolBase: models.CurrencySymbolBase{
					Symbol:  customerSymbol.Symbol,
					Status:  customerSymbol.Status,
					Message: Symbol.Message,
				},
				SystemStatus: systemStatus,
				Amount:       customerSymbol.Amount,
				Simulation:   customerSymbol.Simulation,
			})
		} else {
			// 如果 systemSymboList 中的 Symbol 不存在于 customerSymboMap 中，创建一个新的
			newCustomerSymbol := models.CustomerCurrencySymbol{
				CurrencySymbolBase: models.CurrencySymbolBase{
					Symbol: Symbol.Symbol,
					Status: false,
				},
				Amount:     0,
				Simulation: false,
			}
			result = append(result, models.CustomerCurrencySymboResponse{
				CurrencySymbolBase: models.CurrencySymbolBase{
					Symbol:  newCustomerSymbol.Symbol,
					Status:  newCustomerSymbol.Status,
					Message: newCustomerSymbol.Message,
				},
				SystemStatus: systemStatus,
				Amount:       newCustomerSymbol.Amount,
				Simulation:   newCustomerSymbol.Simulation,
			})
		}
	}

	// Sort the result by Symbo
	sort.Slice(result, func(i, j int) bool {
		return result[i].Symbol < result[j].Symbol
	})

	return result
}
