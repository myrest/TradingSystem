package controllers

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SubaccountList(c *gin.Context) {
	c.HTML(http.StatusOK, "subaccountmanagememt.html", gin.H{})
}

func GetSubaccountList(c *gin.Context) {
	session := sessions.Default(c)
	customerid := session.Get("id").(string)

	subaccounts, err := services.GetSubaccountListByID(c, customerid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": subaccounts})
}

func ModifySubAccount(c *gin.Context) {
	var req models.SubAccountUI

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	session := sessions.Default(c)
	customerid := session.Get("id").(string)

	request := models.SubAccount{
		DocumentRefID: req.DocumentRefID,
		SubAccountDB: models.SubAccountDB{
			AccountName: req.AccountName,
			CustomerID:  customerid,
		},
	}

	rtn, err := services.UpdateSubaccount(c, request)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rtn,
	})
}

func DeleteSubAccount(c *gin.Context) {
	var req models.SubAccountUI

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	session := sessions.Default(c)
	customerid := session.Get("id").(string)

	request := models.SubAccount{
		DocumentRefID: req.DocumentRefID,
		SubAccountDB: models.SubAccountDB{
			AccountName: req.AccountName,
			CustomerID:  customerid,
		},
	}

	err := services.DeleteSubaccount(c, request)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": "OK",
	})
}
