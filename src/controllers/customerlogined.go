package controllers

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/common"
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
	Leverage   string `json:"leverage"`
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
	//有三種情況
	//1. 己登Google，但系統還沒建帳號
	//2. 己登入，用本尊身份
	//3. 己登入，用分身身份

	//情境1
	CustomerByEmail, err := services.GetCustomerByEmail(c, email.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if CustomerByEmail == nil || CustomerByEmail.ID == "" {
		//帳號不存在，要建一個新的
		c.HTML(http.StatusOK, "iscreatenew.html", gin.H{
			"Name":  name,
			"Email": email,
		})
		return
	}

	SubCustomerID := session.Get("id").(string)
	//MainCustomerID := session.Get("parentid").(string)
	//情境2,3

	customer, err := services.GetCustomer(c, SubCustomerID)
	if err == nil {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"Name":                name,
			"Email":               email,
			"ApiKey":              customer.APIKey,
			"SecretKey":           customer.SecretKey,
			"IsAdmin":             customer.IsAdmin,
			"AutoSubscribeStatus": customer.IsAutoSubscribe,
			"AutoSubscribeType":   customer.AutoSubscribReal,
			"AutoSubscribeAmount": customer.AutoSubscribAmount,
		})
	} else {
		session := sessions.Default(c)
		session.Clear()
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}
		c.Redirect(http.StatusFound, "/login?GotError")
	}
}

func GetCustomerBalance(c *gin.Context) {
	session := sessions.Default(c)
	customerid := session.Get("id").(string)
	getcustomerbalance(c, customerid)
}

func getcustomerbalance(c *gin.Context, customerid string) {
	var freeamount float64
	dbCustomer, err := services.GetCustomer(c, customerid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	//有key，啟用時要檢查餘額
	if dbCustomer.APIKey != "" && dbCustomer.SecretKey != "" {
		freeamount, err = services.GetAccountBalance(dbCustomer.APIKey, dbCustomer.SecretKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
	errmsg := ""
	if dbCustomer.APIKey == "" || dbCustomer.SecretKey == "" {
		errmsg = "Missing API, Secert Key."
	}
	c.JSON(http.StatusOK, gin.H{
		"error":  errmsg,
		"amount": freeamount,
	})
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
	amount := common.Decimal(req.Amount)
	leverage := common.Decimal(req.Leverage)
	if leverage == 0 {
		leverage = 1
	}
	input.Amount = amount
	input.Leverage = leverage

	session := sessions.Default(c)
	input.CustomerID = session.Get("id").(string)

	err := services.UpdateCustomerCurrency(context.Background(), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Update customer Symbol failed. " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

func GetAllCustomerSymbol(c *gin.Context) {
	session := sessions.Default(c)
	customerid := session.Get("id").(string)

	systemSymboList, err := services.GetAllSymbol(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	customersymboList, err := services.GetAllCustomerCurrency(context.Background(), customerid)
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
					Symbol: customerSymbol.Symbol,
					Status: customerSymbol.Status,
				},
				SystemStatus: systemStatus,
				Amount:       customerSymbol.Amount,
				Leverage:     customerSymbol.Leverage,
				Simulation:   customerSymbol.Simulation,
				Message:      Symbol.Message,
			})
		} else {
			// 如果 systemSymboList 中的 Symbol 不存在于 customerSymboMap 中，创建一个新的
			newCustomerSymbol := models.CustomerCurrencySymbol{
				CurrencySymbolBase: models.CurrencySymbolBase{
					Symbol: Symbol.Symbol,
					Status: false,
				},
				Amount:     0,
				Leverage:   1,
				Simulation: false,
			}
			result = append(result, models.CustomerCurrencySymboResponse{
				CurrencySymbolBase: models.CurrencySymbolBase{
					Symbol: newCustomerSymbol.Symbol,
					Status: newCustomerSymbol.Status,
				},
				SystemStatus: systemStatus,
				Amount:       newCustomerSymbol.Amount,
				Leverage:     newCustomerSymbol.Leverage,
				Simulation:   newCustomerSymbol.Simulation,
				Message:      "The symbol do not exist in the system.",
			})
		}
	}

	// Sort the result by Symbo
	sort.Slice(result, func(i, j int) bool {
		return result[i].Symbol < result[j].Symbol
	})

	return result
}

type Log_PlaceBetHistoryUI struct {
	models.Log_TvSiginalData
	Position string
}

func PlaceOrderHistory(c *gin.Context) {
	symbol := c.Query("symbol")
	customerid := c.Query("cid")
	session := sessions.Default(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	symbol = common.FormatSymbol(symbol)

	var rtn []Log_PlaceBetHistoryUI

	if customerid == "" {
		cid := session.Get("id")
		if cid != nil {
			customerid = cid.(string)
		}
	}
	if customerid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Customer Data."})
		return
	}

	list, totalPages, err := services.GetPlaceOrderHistory(c, symbol, customerid, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := 0; i < len(list); i++ {
		positionside := "多"
		side := "開"
		if list[i].PositionSideType == bingx.ShortPositionSideType {
			positionside = "空"
		}
		if (list[i].PositionSideType == bingx.ShortPositionSideType && list[i].Side == bingx.BuySideType) ||
			(list[i].PositionSideType == bingx.LongPositionSideType && list[i].Side == bingx.SellSideType) {
			side = "平"
		}
		rtn = append(rtn, Log_PlaceBetHistoryUI{
			Log_TvSiginalData: list[i],
			Position:          side + positionside,
		})
	}

	c.HTML(http.StatusOK, "placeorderhistory.html", gin.H{
		"data":       rtn,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
		"symbol":     symbol,
	})
}

func GetPlaceOrderHistoryBySymbol(c *gin.Context) {
	symbol := c.Query("symbol")
	cid := c.DefaultQuery("cid", "")
	session := sessions.Default(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	symbol = common.FormatSymbol(symbol)

	sessioncid := session.Get("id")
	var customerid string
	if sessioncid != nil {
		customerid = sessioncid.(string)
	}

	//只有管理員可以看到其它人的記錄。
	if customerid != "" && session.Get("isadmin") != nil && session.Get("isadmin").(bool) {
		customerid = cid
	}

	if customerid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Customer Data."})
		return
	}

	list, totalPages, err := services.GetPlaceOrderHistory(c, symbol, customerid, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       list,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
		"symbol":     symbol,
	})
}

func GetTGBot(c *gin.Context) {
	session := sessions.Default(c)
	customerid := session.Get("id").(string)
	customer, err := services.GetCustomer(c, customerid)
	if customer.TgIdentifyKey == "" {
		//如果沒有TgIdentifyKey，就生一個
		customer.TgIdentifyKey, err = services.SetTGIdentifyKey(c, customerid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "tgbot.html", gin.H{
		"tgidentifykey": customer.TgIdentifyKey,
	})
}
