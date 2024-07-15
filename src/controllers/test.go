package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"TradingSystem/src/bingx"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
)

func GetBingxOrderByID(c *gin.Context) {
	id := c.Query("id")
	Symbol := c.Query("symbol")
	idint, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input id"})
		return
	}

	//依訂單編號取得CustomerID
	customerid, err := services.GetCustomerIDByBingxOrderID(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//依CustomerID取得Cert資料
	customer, err := services.GetCustomer(c, customerid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if customer == nil || customer.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer not exist or have no API Key."})
		return
	}

	client := bingx.NewClient(customer.APIKey, customer.SecretKey, false)
	placedOrder, err := client.NewGetOrderService().
		Symbol(Symbol).
		OrderId(idint).
		Do(c)

	//無法取得下單的資料
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"data": placedOrder})
}

func TESTGetOpenOrder(c *gin.Context) {
	var data models.TvWebhookData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var tvData models.TvSiginalData
	tvData.Convert(data)

	//取出有訂閱的人
	customerList, err := services.GetCustomerCurrencySymbosBySymbol(c, "BTCUSDT")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": err})
	}
	for i := 0; i < len(customerList); i++ {
		//placeOrder(customerList[i].APIKey, customerList[i].SecretKey, tvData)

		c.JSON(http.StatusOK, gin.H{"data": "XX"})
	}

}
