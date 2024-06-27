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
	"sync"

	"github.com/gin-gonic/gin"
)

type openPosition struct {
	AvailableAmt float64
	PositionSide bingx.PositionSideType
}

func TradingViewWebhook(c *gin.Context) {
	var WebhookData models.TvWebhookData
	if err := c.ShouldBindJSON(&WebhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := preProcessPlaceOrder(c, WebhookData, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
}

func TradingViewWebhookTEST(c *gin.Context) {
	var WebhookData models.TvWebhookData
	if err := c.ShouldBindJSON(&WebhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := preProcessPlaceOrder(c, WebhookData, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
}

func preProcessPlaceOrder(c *gin.Context, WebhookData models.TvWebhookData, isTEST bool) error {
	var tvData models.TvSiginalData
	tvData.Convert(WebhookData)

	//取出有訂閱的人
	customerList, err := services.GetCustomerCurrencySymbosBySymbol(c, WebhookData.Symbol)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < len(customerList); i++ {
		wg.Add(1)
		go func(customer models.CustomerCurrencySymboWithCustomer) {
			defer wg.Done()
			processPlaceOrder(customer.APIKey, customer.SecretKey, customer.Amount, tvData, isTEST)
		}(customerList[i])
	}
	wg.Wait()

	return nil
}

func processPlaceOrder(APIKey, SecertKey string, amount float64, tv models.TvSiginalData, isTEST bool) {
	client := bingx.NewClient(APIKey, SecertKey, isTEST)

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
	if ((tv.PlaceOrderType.Side == bingx.BuySideType && tv.PlaceOrderType.PositionSideType == bingx.ShortPositionSideType) ||
		(tv.PlaceOrderType.Side == bingx.SellSideType && tv.PlaceOrderType.PositionSideType == bingx.LongPositionSideType)) &&
		oepntrade.AvailableAmt < placeAmount { //要防止平太多，變反向持倉
		placeAmount = oepntrade.AvailableAmt
	}

	if tv.TVData.PositionSize == 0 {
		//全部平倉
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
