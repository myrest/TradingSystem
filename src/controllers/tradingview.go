package controllers

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func TradingViewWebhook(c *gin.Context) {
	var WebhookData models.TvWebhookData
	if err := c.ShouldBindJSON(&WebhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var tvData models.TvSiginalData
	tvData.Convert(WebhookData)

	//取出有訂閱的人
	customerList, err := services.GetCustomerCurrencySymbosBySymbol(c, WebhookData.Symbol)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": err})
	}
	for i := 0; i < len(customerList); i++ {
		processPlaceOrder(customerList[i].APIKey, customerList[i].SecretKey, customerList[i].Amount, tvData)
	}

}

type openPosition struct {
	AvailableAmt float64
	PositionSide bingx.PositionSideType
}

func processPlaceOrder(APIKey, SecertKey string, amount float64, tv models.TvSiginalData) {
	client := bingx.NewClient(APIKey, SecertKey, true)

	//查出目前持倉情況
	positions, err := client.NewGetOpenPositionsService().Symbol(tv.TVData.Symbol).Do(context.Background())
	if err != nil {
		//Todo:要寫Log
		return
	}

	var oepntrade openPosition //目前持倉
	if len(*positions) > 0 {
		amount, _ := strconv.ParseFloat((*positions)[0].AvailableAmt, 64)
		oepntrade.AvailableAmt = amount
		if strings.ToLower((*positions)[0].PositionSide) == "long" {
			oepntrade.PositionSide = bingx.LongPositionSideType
		} else {
			oepntrade.PositionSide = bingx.ShortPositionSideType
		}
	}
	//計算下單數量
	placeAmount := tv.TVData.Contracts * amount / 100
	if tv.PlaceOrderType.Side == bingx.SellSideType && oepntrade.AvailableAmt < placeAmount { //要防止平太多，變反向持倉
		placeAmount = oepntrade.AvailableAmt
	}

	order, err := client.NewCreateOrderService().
		PositionSide(tv.PlaceOrderType.PositionSideType). //bingx.LongPositionSideType
		Symbol(tv.TVData.Symbol).
		Quantity(placeAmount).
		Type(bingx.MarketOrderType).
		Side(tv.PlaceOrderType.Side).
		Do(context.Background())
	if err != nil {
		log.Println(err)
	}
	log.Printf("Limit order created: %+v", order)
	//Todo:要寫Log
}
