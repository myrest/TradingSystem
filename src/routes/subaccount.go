package routes

import (
	"TradingSystem/src/controllers"
	"TradingSystem/src/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterSubaccountRoutes(r *gin.Engine) {
	customerRoutes := r.Group("/subaccount")
	customerRoutes.Use(middleware.CustomerMiddleware())
	{
		customerRoutes.GET("/", controllers.SubaccountList)
		customerRoutes.POST("/update", controllers.ModifySubAccount)
	}

}
