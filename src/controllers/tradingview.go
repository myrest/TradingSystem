package controllers

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

	err := preProcessPlaceOrder(c, WebhookData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func preProcessPlaceOrder(c *gin.Context, WebhookData models.TvWebhookData) error {
	// 寫入 WebhookData 到 Firestore
	TvWebHookLog, err := services.SaveWebhookData(context.Background(), WebhookData)
	if err != nil {
		log.Printf("Failed to save webhook data: %v", err.Error())
	}

	//檢查Cert
	_, err = services.GetSymbol(c, WebhookData.Symbol, WebhookData.Cert)
	if err != nil {
		//Todo:要寫Log
		return err
	}

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
			processPlaceOrder(customer, tvData, TvWebHookLog, customer.APIKey, customer.SecretKey, c)
		}(customerList[i])
	}
	wg.Wait()
	//要更新績效的cache
	go func() {
		//清暫存檔
		services.RemoveLog_TVExpiredCacheFiles()
	}()
	return nil
}

func processPlaceOrder(Customer models.CustomerCurrencySymboWithCustomer, tv models.TvSiginalData, TvWebHookLog, APIKey, SecertKey string, c *gin.Context) {
	client := bingx.NewClient(APIKey, SecertKey, Customer.Simulation)
	placeOrderLog, isTowWayPositionOnHand, AlertMessageModel, err := client.CreateOrder(c, tv, Customer)
	placeOrderLog.WebHookRefID = TvWebHookLog
	if err != nil {
		placeOrderLog.Result = placeOrderLog.Result + "\nPlace order get exception:" + err.Error()
		services.SystemEventLog{
			EventName:  services.PlaceOrder,
			CustomerID: Customer.CustomerID,
			Message:    fmt.Sprintf("Place order failed:%s\nsymbol:%s\ncustomerId:%s", err.Error(), tv.TVData.Symbol, Customer.ID),
		}.Send()
	}

	asyncWriteTVsignalData(AlertMessageModel, &Customer.Customer, placeOrderLog, c, isTowWayPositionOnHand)
}

// 寫log
func asyncWriteTVsignalData(alertType models.AlertMessageModel, customer *models.Customer, tvdata models.Log_TvSiginalData, c *gin.Context, isTwoWayPosition bool) {
	if isTwoWayPosition {
		tvdata.Result = tvdata.Result + "\n⚠️⚠️偵測到雙向持倉情況，請立即檢查倉位。"
	}
	go func(data models.Log_TvSiginalData) {
		_, err := services.SaveCustomerPlaceOrderResultLog(c, data)
		if err != nil {
			log.Printf("Failed to save webhook data: %v", err)
		}

		customerAlertLevel := customer.AlertMessageType.GetPriority()
		systmeAlertLevel := alertType.GetPriority()
		//有綁定，且訊息等級要夠才發
		if customer.TgChatID > 0 && (customerAlertLevel >= systmeAlertLevel) {
			positionside := "多"
			side := "開"
			if tvdata.PositionSideType == models.ShortPositionSideType {
				positionside = "空"
			}
			if (tvdata.PositionSideType == models.ShortPositionSideType && tvdata.Side == models.BuySideType) ||
				(tvdata.PositionSideType == models.LongPositionSideType && tvdata.Side == models.SellSideType) {
				side = "平"
			}
			tgMessage := fmt.Sprintf("幣種：%s \n方向：%s%s\n結果/單號：%s", tvdata.Symbol, side, positionside, tvdata.Result)
			if tvdata.Profit != 0 {
				tgMessage = fmt.Sprintf("%s\n盈虧：%s", tgMessage, formatFloat64(6, tvdata.Profit))
			}
			if data.Simulation {
				tgMessage = fmt.Sprintf("%s\n【***模擬交易單***】", tgMessage)
			}
			//發送TG訊號
			err := services.TGSendMessage(customer.TgChatID, tgMessage)
			if err != nil {
				services.CustomerEventLog{
					CustomerID: customer.ID,
					EventName:  services.PlaceOrder,
					Message:    err.Error(),
				}.Send(c)
			}
		}
	}(tvdata)
}

func formatFloat64(round int, f float64) string {
	value := common.Decimal(f, round)
	return strconv.FormatFloat(value, 'f', -1, 64)
}
