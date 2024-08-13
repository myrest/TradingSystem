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
		handleCustomError(c, http.StatusBadRequest, "Invalid request", true)
		return
	}

	session := sessions.Default(c)
	var googleUser models.GoogleTokenDetail
	if tokenResult, err := services.VerifyIDTokenAndGetDetails(req.Token); err != nil {
		handleCustomError(c, http.StatusBadRequest, "Invalid ID token", true)
		return
	} else {
		googleUser = tokenResult
	}

	session.Set("uid", googleUser.UID)
	session.Set("name", googleUser.Name)
	session.Set("email", googleUser.Email)
	session.Set("photo", googleUser.Photo)

	customer, err := services.GetCustomerByEmail(c, googleUser.Email)
	if err == nil && customer != nil {
		session.Set("isadmin", customer.IsAdmin)
		session.Set("id", customer.ID)
		session.Set("parentid", customer.ID)
		services.CustomerEventLog{
			CustomerID: customer.ID,
			EventName:  services.EventNameLogin,
			Message:    googleUser.Email,
		}.Send(c)
	} else {
		services.CustomerEventLog{
			CustomerID: "NewCommer",
			EventName:  services.EventNameLogin,
			Message:    googleUser.Email,
		}.Send(c)
		session.Set("isadmin", false)
	}

	if err := session.Save(); err != nil {
		handleCustomError(c, http.StatusBadRequest, "Failed to save session", true)
		return
	}

	c.Redirect(http.StatusFound, "/customers/dashboard")
}

func GoogleLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		handleCustomError(c, http.StatusBadRequest, "Failed to save session")
		return
	}

	c.Redirect(http.StatusFound, "/")
}
