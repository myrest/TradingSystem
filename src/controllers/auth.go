package controllers

import (
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

	uid, email, name, photo, err := services.VerifyIDTokenAndGetDetails(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid ID token"})
		return
	}

	session := sessions.Default(c)
	session.Set("uid", uid)
	session.Set("name", name)
	session.Set("email", email)
	session.Set("photo", photo)

	customer, err := services.GetCustomerByEmail(c, email)
	if err == nil && customer != nil {
		session.Set("isadmin", customer.IsAdmin)
		session.Set("id", customer.ID)
		session.Set("parentid", customer.ID)
		services.CustomerEventLog{
			CustomerID: customer.ID,
			EventName:  services.EventNameLogin,
			Message:    email,
		}.Send(c)
	} else {
		services.CustomerEventLog{
			CustomerID: "NewCommer",
			EventName:  services.EventNameLogin,
			Message:    email,
		}.Send(c)
		session.Set("isadmin", false)
	}

	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
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
