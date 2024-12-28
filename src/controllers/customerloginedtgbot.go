package controllers

import (
	"TradingSystem/src/services"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetTGBot(c *gin.Context) {
	isLinked := false
	session := sessions.Default(c)
	customerid := session.Get("id").(string)
	customer, err := services.GetCustomer(c, customerid)

	if err != nil {
		c.Error(err) // 將錯誤添加到上下文中
		return
	}

	if customer.TgChatID != 0 {
		isLinked = true
	}

	if customer.TgIdentifyKey == "" {
		//如果沒有TgIdentifyKey，就生一個
		customer.TgIdentifyKey, err = services.SetTGIdentifyKey(c, customerid)
		if err != nil {
			c.Error(err) // 將錯誤添加到上下文中
			return
		}
	}

	c.HTML(http.StatusOK, "tgbot.html", gin.H{
		"tgidentifykey":     customer.TgIdentifyKey,
		"islinked":          isLinked,
		"StaticFileVersion": systemsettings.StartTimestemp,
	})
}
