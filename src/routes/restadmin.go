package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRestAdminRoutes(r *gin.Engine) {
	authRoutes := r.Group("/restadmin")
	{
		authRoutes.POST("/symbol", controllers.AddNewSymbol)
		authRoutes.PATCH("/symbolStatus", controllers.UpdateStatus)
		authRoutes.PATCH("/symbolMessage", controllers.UpdateMessage)
		authRoutes.GET("/symbol", controllers.GetAllSymbol)
		authRoutes.GET("/subscriber", controllers.GetSubscribeCustomerBySymbol)
	}
}
