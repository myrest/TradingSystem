package controllers

import (
	"TradingSystem/src/models"
	"TradingSystem/src/services"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GoogleAuthCallback(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	uid, email, name, err := services.VerifyIDTokenAndGetDetails(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid ID token"})
		return
	}

	session := sessions.Default(c)
	session.Set("uid", uid)
	session.Set("name", name)
	session.Set("email", email)

	customer, err := services.GetCustomerByEmail(email)
	if err == nil && customer != nil {
		session.Set("isadmin", customer.IsAdmin)
		session.Set("id", customer.ID)
		session.Set("apikey", customer.APIKey)
		session.Set("secertkey", customer.SecretKey)
	} else {
		session.Set("isadmin", false)
	}

	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.Redirect(http.StatusFound, "/customers/dashboard")
}

func CreateAccount(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")

	customer := &models.Customer{
		Name:  name,
		Email: email,
	}

	err := services.CreateCustomerAccount(customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	c.Redirect(http.StatusFound, "/customers/dashboard")
}

func GoogleLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.Redirect(http.StatusFound, "/")
}
