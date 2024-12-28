package controllers

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ShowLoginPage(c *gin.Context) {
	currentHost, err := common.GetHostName(c)
	if err != nil {
		//c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Println(err)
		//return //先不處理Host問題
	}
	session := sessions.Default(c)
	name := session.Get("name")
	email := session.Get("email")

	if name == nil || email == nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"StaticFileVersion": systemsettings.StartTimestemp,
			"host":              fmt.Sprintf("%s (%s)", systemsettings.Env.String(), currentHost),
		})
		return
	}
	c.Redirect(http.StatusFound, "/customers/dashboard")
}

func CreateCustomer(c *gin.Context) {
	session := sessions.Default(c)

	var customer = models.Customer{
		Name:  session.Get("name").(string),
		Email: session.Get("email").(string),
	}
	//先查該Email是否有被用掉。
	dbCustomer, err := services.GetCustomerByEmail(c, customer.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting customer"})
		return
	}

	if dbCustomer != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account: " + dbCustomer.Email + "is exist."})
		return
	}

	id, err := services.CreateCustomer(context.Background(), &customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating customer"})
		return
	}
	session.Set("id", id)
	session.Set("parentid", id)

	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	customer.ID = id
	c.JSON(http.StatusOK, customer)
}

func UpdateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	session := sessions.Default(c)
	id := session.Get("id").(string)
	dbCustomer, err := services.GetCustomer(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer data not exist."})
		return
	}

	customer.APIKey = strings.TrimSpace(customer.APIKey)
	customer.SecretKey = strings.TrimSpace(customer.SecretKey)
	//因為ID, Name, Email不可改，所以拿原來的套回去
	customer.ID = dbCustomer.ID
	customer.Name = dbCustomer.Name
	customer.Email = dbCustomer.Email

	if err := services.UpdateCustomer(context.Background(), &customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating customer"})
		return
	}

	c.JSON(http.StatusOK, dbCustomer)
}
