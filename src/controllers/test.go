package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"TradingSystem/src/bingx"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
)

var APIKey = "LCHu8mULibUy3IdkIj78UgPtrfuhZCho7IeWCmHYh8JcXlwCIoJaBPosntg967PkCJ9QpHqBDSneHtLX9Sg"
var SecertKey = "uub6B7scfX5AVmTYHPlNbmBbf9pLJMY7lgpAq7qunDkw5gDP7xgWLBHkduESKCjlpHENsCwHPpX8EYVw"

func GetBingxOrderByID(c *gin.Context) {
	id := c.Query("id")
	Symbol := c.Query("symbol")
	idint, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input id"})
	}

	client := bingx.NewClient(APIKey, SecertKey, false)
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

func TEST(c *gin.Context) {
	client := bingx.NewClient(APIKey, SecertKey, true)
	order, err := client.NewGetOrderService().ClientOrderId("1805636386072563712").
		Symbol("BTC-USDT").
		Do(context.Background())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": err.Error()})
		return
	}
	profit, _ := strconv.ParseFloat(order.Profit, 64)
	c.JSON(http.StatusOK, gin.H{"data": profit})
}

func TESTBet(c *gin.Context) {
	rtn, err := services.GetCustomerCurrencySymbosBySymbol(c, "BTCUSDT")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": err})
	}
	c.JSON(http.StatusOK, gin.H{"data": rtn})
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
