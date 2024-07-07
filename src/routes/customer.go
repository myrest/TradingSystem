package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterCustomerRoutes(r *gin.Engine) {
	customerRoutes := r.Group("/customers")
	{
		customerRoutes.GET("/placeorderhistory", controllers.PlaceOrderHistory)
		customerRoutes.GET("/getplaceorderhistory", controllers.GetPlaceOrderHistoryBySymbol)
		customerRoutes.GET("/dashboard", controllers.ShowDashboardPage)
		customerRoutes.GET("/symbol", controllers.GetAllCustomerSymbol)
		customerRoutes.POST("", controllers.CreateCustomer)
		customerRoutes.POST("/update", controllers.UpdateCustomer)
		customerRoutes.PATCH("/symbol", controllers.UpdateCustomerSymbol)
	}
}
