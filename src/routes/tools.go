package routes

import (
	"TradingSystem/src/common"
	"TradingSystem/src/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterMyTestRoutes(r *gin.Engine) {
	settings := common.GetEnvironmentSetting()

	authRoutes := r.Group(settings.SectestWord)
	{
		authRoutes.GET("/getbyid", controllers.GetBingxOrderByID)
		authRoutes.GET("/getavailablebalance", controllers.GetAvailableAmountByID)
		authRoutes.GET("/systemsettings", controllers.SystemSettings)
		authRoutes.POST("/savesystemsettings", controllers.SaveSystemSettings)
		authRoutes.POST("/tradingview", controllers.PlaceOrderManually)
	}

	specialrouter := r.Group("/resthome")
	{
		specialrouter.PATCH("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": settings.SectestWord})
		})
	}

}
