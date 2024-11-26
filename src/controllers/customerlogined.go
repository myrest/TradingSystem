package controllers

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type updateCustomerSymboRequest struct {
	Symbol         string `json:"symbol"`
	Status         string `json:"status"`
	Amount         string `json:"amount"`
	Leverage       string `json:"leverage"`
	Simulation     string `json:"simulation"`
	UpdateLeverage string `json:"updateleverage"`
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
			"Name":              name,
			"Email":             email,
			"StaticFileVersion": common.GetEnvironmentSetting().StartTimestemp,
		})
		return
	}

	SubCustomerID := session.Get("id").(string)
	//MainCustomerID := session.Get("parentid").(string)
	//情境2,3

	customer, err := services.GetCustomer(c, SubCustomerID)
	//為了向下相容
	if customer.ExchangeSystemName == "" {
		customer.ExchangeSystemName = models.ExchangeBingx
	}

	//Binance只有實盤
	if customer.ExchangeSystemName == models.ExchangeBinance_N ||
		customer.ExchangeSystemName == models.ExchangeBinance_P {
		customer.AutoSubscribReal = true
	}

	if err == nil {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"Name":                name,
			"Email":               email,
			"ApiKey":              customer.APIKey,
			"SecretKey":           customer.SecretKey,
			"IsAdmin":             c.GetBool("IsAdmin"),
			"AutoSubscribeStatus": customer.IsAutoSubscribe,
			"AutoSubscribeType":   customer.AutoSubscribReal,
			"AutoSubscribeAmount": customer.AutoSubscribAmount,
			"AlertMessageType":    customer.AlertMessageType,
			"StaticFileVersion":   common.GetEnvironmentSetting().StartTimestemp,
			"ExchangeSystemName":  customer.ExchangeSystemName,
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
		freeamount, err = services.GetAccountBalance(c, dbCustomer.APIKey, dbCustomer.SecretKey, dbCustomer.ExchangeSystemName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
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

	err := services.UpdateCustomerCurrency(context.Background(), &input, req.UpdateLeverage == "1")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Update customer Symbol failed. " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

func GetAllCustomerSymbol(c *gin.Context) {
	session := sessions.Default(c)
	customerid := session.Get("id").(string)
	rtn, err := getAllCustomerSymbolByCustomerID(customerid)

	if err != nil {
		return
	}

	c.JSON(http.StatusOK, rtn)
}

func getAllCustomerSymbolByCustomerID(CustomerID string) ([]models.CustomerCurrencySymboResponse, error) {
	systemSymboList, err := services.GetAllSymbol(context.Background())
	if err != nil {
		return nil, err
	}

	customersymboList, err := services.GetAllCustomerCurrency(context.Background(), CustomerID)
	if err != nil {
		return nil, err
	}
	mergedList := mergeSymboLists(systemSymboList, customersymboList)
	return mergedList, nil
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
				Message:      Symbol.Message,
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
	sdt, edt := common.GetReportStartEndDate(session)
	if sdt == common.TimeMax() || edt == common.TimeMax() {
		sdt = common.GetMonthlyDay1(1)[0]
		edt = time.Now().UTC()
		common.SetReportStartEndDate(session, sdt, edt)
	}

	list, totalPages, err := services.GetPlaceOrderHistory(c, symbol, customerid, sdt, edt, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := 0; i < len(list); i++ {
		positionside := "多"
		side := "開"
		if list[i].PositionSideType == models.ShortPositionSideType {
			positionside = "空"
		}
		if (list[i].PositionSideType == models.ShortPositionSideType && list[i].Side == models.BuySideType) ||
			(list[i].PositionSideType == models.LongPositionSideType && list[i].Side == models.SellSideType) {
			side = "平"
		}
		rtn = append(rtn, Log_PlaceBetHistoryUI{
			Log_TvSiginalData: list[i],
			Position:          side + positionside,
		})
	}

	c.HTML(http.StatusOK, "placeorderhistory.html", gin.H{
		"data":              rtn,
		"page":              page,
		"pageSize":          pageSize,
		"totalPages":        totalPages,
		"symbol":            symbol,
		"cid":               c.Query("cid"),
		"StaticFileVersion": common.GetEnvironmentSetting().StartTimestemp,
	})
}
