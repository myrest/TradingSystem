package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"TradingSystem/src/bingx"
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

func GetAvailableAmountByID(c *gin.Context) {
	CustomerID := c.Query("cid")
	account, err := services.GetCustomer(c, CustomerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	freeamount, err := services.GetAccountBalance(account.APIKey, account.SecretKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"amount": freeamount})
}

func TEST3(c *gin.Context) {
	Symbol := c.Query("symbol")
	CustomerID := c.Query("cid")

	//依CustomerID取得Cert資料
	customer, err := services.GetCustomer(c, CustomerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if customer == nil || customer.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer not exist or have no API Key."})
		return
	}

	client := bingx.NewClient(customer.APIKey, customer.SecretKey, true)
	tradleverage, err := client.NewGetTradService().
		Symbol(Symbol).
		Do(c)

	//無法取得下單的資料
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"data": tradleverage})

}

func TEST2(c *gin.Context) {
	Symbol := c.Query("symbol")
	CustomerID := c.Query("cid")
	Leverage, _ := strconv.ParseInt(c.Query("leverage"), 10, 64)

	if Leverage == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leverage不得為0"})
		return
	}

	//依CustomerID取得Cert資料
	customer, err := services.GetCustomer(c, CustomerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if customer == nil || customer.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer not exist or have no API Key."})
		return
	}

	client := bingx.NewClient(customer.APIKey, customer.SecretKey, true)
	client.Debug = true
	//改多單槓桿
	_, err = client.NewSetTradService().
		Symbol(Symbol).
		PositionSide(bingx.LongPositionSideType).
		Leverage(Leverage).
		Do(c)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	}

	//改空單槓桿
	tradleverage, err := client.NewSetTradService().
		Symbol(Symbol).
		PositionSide(bingx.ShortPositionSideType).
		Leverage(Leverage).
		Do(c)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"data": tradleverage})

}
