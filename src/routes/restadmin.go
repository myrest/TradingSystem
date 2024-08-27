package routes

import (
	"TradingSystem/src/controllers"
	"TradingSystem/src/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRestAdminRoutes(r *gin.Engine) {
	authRoutes := r.Group("/restadmin")
	authRoutes.Use(middleware.AdminMiddleware())
	{
		authRoutes.POST("/symbol", controllers.AddNewSymbol)
		authRoutes.DELETE("/symbol", controllers.DeleteSymbol)
		authRoutes.PATCH("/symbolStatus", controllers.UpdateStatus)
		authRoutes.PATCH("/symbolMessage", controllers.UpdateMessage)
		authRoutes.GET("/symbol", controllers.GetAllSymbol)
		authRoutes.GET("/subscriber", controllers.GetSubscribeCustomerBySymbol)
		authRoutes.GET("/customers", controllers.GetAllCustomerList)
		authRoutes.GET("/customersubscribe", controllers.GetSubscribeSymbolbyCompanyID)
		authRoutes.GET("/customerdata", controllers.GetCustomerData)
	}
}
