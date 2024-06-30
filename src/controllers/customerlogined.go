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

type CustomerCurrencySymboResponse struct {
	models.CustomerCurrencySymbo
	SystemStatus string
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Update customer symbo failed. " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, errormessage)
}

func GetAllCustomerSymbo(c *gin.Context) {
	session := sessions.Default(c)
	customerid := session.Get("id").(string)

	systemSymboList, err := services.GetAllSymbo(context.Background())
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

func mergeSymboLists(systemSymboList []models.CurrencySymbo, customersymboList []models.CustomerCurrencySymbo) []CustomerCurrencySymboResponse {
	customerSymboMap := make(map[string]models.CustomerCurrencySymbo)
	for _, symbo := range customersymboList {
		customerSymboMap[symbo.Symbo] = symbo
	}

	var result []CustomerCurrencySymboResponse

	// Iterate through systemSymboList
	for _, symbo := range systemSymboList {
		systemStatus := "Disabled"
		if symbo.Status {
			systemStatus = "Enabled"
		}
		if customerSymbo, exists := customerSymboMap[symbo.Symbo]; exists {
			// 如果 systemSymboList 中的 Symbo 存在于 customerSymboMap 中
			result = append(result, CustomerCurrencySymboResponse{
				CustomerCurrencySymbo: customerSymbo,
				SystemStatus:          systemStatus,
			})
		} else {
			// 如果 systemSymboList 中的 Symbo 不存在于 customerSymboMap 中，创建一个新的
			newCustomerSymbo := models.CustomerCurrencySymbo{
				CurrencySymbo: models.CurrencySymbo{
					AdminCurrencySymbo: models.AdminCurrencySymbo{
						Symbo:  symbo.Symbo,
						Status: false,
					},
					//Cert不需顯示給用戶
				},
				Amount: 0,
			}
			result = append(result, CustomerCurrencySymboResponse{
				CustomerCurrencySymbo: newCustomerSymbo,
				SystemStatus:          systemStatus,
			})
		}
	}

	// Sort the result by Symbo
	sort.Slice(result, func(i, j int) bool {
		return result[i].Symbo < result[j].Symbo
	})

	return result
}
