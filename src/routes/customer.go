package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterCustomerRoutes(r *gin.Engine) {
	customerRoutes := r.Group("/customers")
	{
		customerRoutes.POST("", controllers.CreateCustomer)
		customerRoutes.POST("/update", controllers.UpdateCustomer)
		customerRoutes.GET("/dashboard", controllers.ShowDashboardPage)
		customerRoutes.GET("/symbo", controllers.GetAllCustomerSymbol)
		customerRoutes.PATCH("/symbo", controllers.UpdateCustomerSymbol)
	}
}
