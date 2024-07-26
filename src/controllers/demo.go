package controllers

import (
	"TradingSystem/src/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DemoList(c *gin.Context) {
	customerid := "8LcgaKWUvn1LyUQQj3oP"

	systemSymboList, err := services.GetAllSymbol(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	customersymboList, err := services.GetAllCustomerCurrency(c, customerid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mergedList := mergeSymboLists(systemSymboList, customersymboList)
	c.HTML(http.StatusOK, "demosymbolist.html", gin.H{
		"data": mergedList,
	})
}
