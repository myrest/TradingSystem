package controllers

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddNewSymbol(c *gin.Context) {
	var data models.CurrencySymbol

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	rtn, err := services.CreateNewSymbol(context.Background(), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error createing symbo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rtn.Cert})
}

type updateStatusRequest struct {
	Symbol string `json:"symbol"`
	Status string `json:"status"`
}

func UpdateSymbol(c *gin.Context) {
	var data models.CurrencySymbol
	var req updateStatusRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	data = models.CurrencySymbol{
		AdminCurrencySymbol: models.AdminCurrencySymbol{
			Symbol: req.Symbol,
			Status: req.Status == "true",
		},
		//Cert不能改
	}

	if err := services.UpdateSymbol(context.Background(), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updateing Symbol."})
		return
	}
}

func GetAllSymbol(c *gin.Context) {
	var rtn []models.AdminSymboListUI

	symboList, err := services.GetAllSymbol(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	webhooklist, err := services.GetLatestWebhook(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			CurrencySymbol: Symbol,
			PositionSize:   positionSize,
		})
	}

	c.JSON(http.StatusOK, rtn)
}
