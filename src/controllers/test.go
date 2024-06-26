package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ManageAPI/src/models"
	"ManageAPI/src/services"
)

//var _APIKey = "LCHu8mULibUy3IdkIj78UgPtrfuhZCho7IeWCmHYh8JcXlwCIoJaBPosntg967PkCJ9QpHqBDSneHtLX9Sg"
//var _SecertKey = "uub6B7scfX5AVmTYHPlNbmBbf9pLJMY7lgpAq7qunDkw5gDP7xgWLBHkduESKCjlpHENsCwHPpX8EYVw"

func TEST(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": "TEST T1"})
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
