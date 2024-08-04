package controllers

import (
	"TradingSystem/src/services"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SubaccountList(c *gin.Context) {
	session := sessions.Default(c)
	customerid := session.Get("id").(string)

	subaccounts, err := services.GetSubaccountListByID(c, customerid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "subaccountmanagememt.html", gin.H{
		"data": subaccounts,
	})
}

func ModifySubAccount(c *gin.Context) {
	d := c.Query("d")
	days, _ := strconv.Atoi(d)
	if days == 0 {
		days = 7
	} else if days > 30 {
		days = 30
	}

	systemSymboList, err := services.GetDemoCurrencyList(c, days, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "demosymbolist.html", gin.H{
		"data": systemSymboList,
		"days": days,
	})
}
