package controllers

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AddNewSymbol(c *gin.Context) {
	var data models.AdminCurrencySymbol

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if data.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid symbo"})
		return
	}

	rtn, err := services.CreateNewSymbol(context.Background(), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error createing symbol"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rtn})
}

func DeleteSymbol(c *gin.Context) {
	symbol := c.Query("symbol")
	cert := c.Query("cert")
	adminSymbol, err := services.GetSymbol(c, symbol, cert)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = services.DeleteAdminSymbol(c, adminSymbol.Symbol)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//把customer的訂閱停掉。
	err = services.DisableCustomerSymbolStatus(c, adminSymbol.Symbol)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ""})
}

type updateStatusRequest struct {
	Symbol string `json:"symbol"`
	Status string `json:"status"`
}

func UpdateStatus(c *gin.Context) {
	var data models.AdminCurrencySymbol
	var req updateStatusRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if req.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid symbo"})
		return
	}

	data = models.AdminCurrencySymbol{
		CurrencySymbolBase: models.CurrencySymbolBase{
			Symbol: req.Symbol,
			Status: req.Status == "true",
		},
		//Cert不能改
	}

	if err := services.UpdateSymbolStatus(context.Background(), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updateing Symbol."})
		return
	}
}

type updateMessageRequest struct {
	Symbol  string `json:"symbol"`
	Message string `json:"message"`
}

func UpdateMessage(c *gin.Context) {
	var data models.AdminCurrencySymbol
	var req updateMessageRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if req.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid symbo"})
		return
	}

	data = models.AdminCurrencySymbol{
		CurrencySymbolBase: models.CurrencySymbolBase{
			Symbol: req.Symbol,
		},
		Message: req.Message,
	}

	if err := services.UpdateSymbolMessage(context.Background(), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updateing Symbol."})
		return
	}
}

func GetAllSymbol(c *gin.Context) {
	var rtn []models.AdminSymboListUI

	symboList, err := services.GetAllSymbol(context.Background())
	if err != nil {
		c.JSON(http.StatusFound, gin.H{"error": err.Error()})
		return
	}

	webhooklist, err := services.GetLatestWebhook(context.Background())
	if err != nil {
		c.JSON(http.StatusFound, gin.H{"error": err.Error()})
		return
	}

	// 将webhooklist的数据合并到symboList中
	webhookMap := make(map[string]models.TvWebhookData)
	for _, webhook := range webhooklist {
		webhookMap[webhook.Symbol] = webhook
	}

	for _, Symbol := range symboList {
		positionSize := ""
		if webhook, exists := webhookMap[Symbol.Symbol]; exists {
			positionSize = webhook.Data.PositionSize
		}
		rtn = append(rtn, models.AdminSymboListUI{
			AdminCurrencySymbol: Symbol,
			PositionSize:        positionSize,
		})
	}

	c.JSON(http.StatusOK, rtn)
}

func GetSubscribeCustomerBySymbol(c *gin.Context) {
	//page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	symbol := c.DefaultQuery("symbol", "")

	customerSymbolList, err := services.GetSubscribeCustomersBySymbol(context.Background(), symbol)
	if err != nil {
		panic(err)
	}

	c.HTML(http.StatusOK, "adminviewcustomersymbolist.html", gin.H{
		"data":              customerSymbolList,
		"symbol":            symbol,
		"StaticFileVersion": common.GetEnvironmentSetting().StartTimestemp,
	})
}

func GetAllCustomerList(c *gin.Context) {
	MappedSubCustomerList, err := services.GetMappedCustomerList(c)
	if err != nil {
		return
	}

	rtn := sortCustomerRelationsByMainSub(MappedSubCustomerList)

	c.HTML(http.StatusOK, "adminviewcustomers.html", gin.H{
		"data":              rtn,
		"StaticFileVersion": common.GetEnvironmentSetting().StartTimestemp,
	})
}

func GetSubscribeSymbolbyCompanyID(c *gin.Context) {
	customerid := c.Query("cid")
	rtn, err := getAllCustomerSymbolByCustomerID(customerid)
	if err != nil {
		return
	}
	c.HTML(http.StatusOK, "adminviewcustomersubscribe.html", gin.H{
		"data":              rtn,
		"cid":               customerid,
		"StaticFileVersion": common.GetEnvironmentSetting().StartTimestemp,
	})
}

func GetCustomerData(c *gin.Context) {
	customerid := c.Query("cid")
	data := services.GetCustomerData(c, customerid)
	customer, _ := services.GetCustomer(c, customerid)

	c.HTML(http.StatusOK, "adminviewcustomerssum.html", gin.H{
		"data":              data,
		"customer":          customer,
		"StaticFileVersion": common.GetEnvironmentSetting().StartTimestemp,
	})
}

// 按照主账号和子账号的关系进行排序
func sortCustomerRelationsByMainSub(mappedCustomer map[string]models.CustomerRelationUI) []models.CustomerRelationUI {
	mainAccounts := make(map[string]models.CustomerRelationUI)
	var rtn []models.CustomerRelationUI

	//先處理Main account，以確保都能找到parent資料
	for _, value := range mappedCustomer {
		if strings.Contains(value.Customer.Email, "@") {
			mainAccounts[value.Customer.ID] = value
		}
	}

	//開始Map有sub的資料
	for _, mainAcc := range mainAccounts {
		//新增main
		rtn = append(rtn, mainAcc)
		//新增sub accounts under that main.
		for _, value := range mappedCustomer {
			if value.Parent_CustomerID == mainAcc.Customer.ID {
				rtn = append(rtn, value)
			}
		}
	}
	return rtn
}
