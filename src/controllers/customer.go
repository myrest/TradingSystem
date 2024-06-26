package controllers

import (
	"ManageAPI/src/models"
	"ManageAPI/src/services"
	"context"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ShowDashboardPage(c *gin.Context) {
	session := sessions.Default(c)
	name := session.Get("name")
	email := session.Get("email")

	if name == nil || email == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	customer, err := services.GetCustomerByEmail(email.(string))
	if err == nil {
		if customer == nil {
			//帳號不存在，要建立一個新
			c.HTML(http.StatusOK, "iscreatenew.html", gin.H{
				"Name":  name,
				"Email": email,
			})
		} else {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Name":      name,
				"Email":     email,
				"ApiKey":    customer.APIKey,
				"SecretKey": customer.SecretKey,
				"IsAdmin":   customer.IsAdmin,
			})
		}
	} else {
		c.Redirect(http.StatusFound, "/login?GotError")
	}
}

func CreateCustomer(c *gin.Context) {
	session := sessions.Default(c)
	var customer = models.Customer{
		Name:  session.Get("name").(string),
		Email: session.Get("email").(string),
	}
	//先查該Email是否有被用掉。
	dbCustomer, err := services.GetCustomerByEmail(customer.Email)
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
	email := session.Get("email")
	dbCustomer, err := services.GetCustomerByEmail(email.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer data not exist."})
		return
	}

	dbCustomer.APIKey = customer.APIKey
	dbCustomer.SecretKey = customer.SecretKey

	if err := services.UpdateCustomer(context.Background(), dbCustomer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating customer"})
		return
	}
	session.Set("apikey", customer.APIKey)
	session.Set("secertkey", customer.SecretKey)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, customer)
}
