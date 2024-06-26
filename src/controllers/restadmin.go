package controllers

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddNewSymbo(c *gin.Context) {
	var data models.CurrencySymbo

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := services.CreateNewSymbo(context.Background(), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error createing symbo."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

type updateStatusRequest struct {
	Symbo  string `json:"symbo"`
	Status string `json:"status"`
}

func UpdateSymbo(c *gin.Context) {
	var data models.CurrencySymbo
	var req updateStatusRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	data = models.CurrencySymbo{
		Symbo:  req.Symbo,
		Status: req.Status == "true",
	}

	if err := services.UpdateSymbo(context.Background(), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updateing symbo."})
		return
	}
}

func GetAllSymbo(c *gin.Context) {
	symboList, err := services.GetAllSymbo(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, symboList)
}
