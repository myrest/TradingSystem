package routes

import (
	"ManageAPI/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterCustomerRoutes(r *gin.Engine) {
	customerRoutes := r.Group("/customers")
	{
		customerRoutes.POST("", controllers.CreateCustomer)
		customerRoutes.POST("/update", controllers.UpdateCustomer)
		customerRoutes.GET("/dashboard", controllers.ShowDashboardPage)
		customerRoutes.GET("/symbo", controllers.GetAllCustomerSymbo)
		customerRoutes.PATCH("/symbo", controllers.UpdateCustomerSymbo)
	}
}
