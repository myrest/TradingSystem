package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRestAdminRoutes(r *gin.Engine) {
	authRoutes := r.Group("/restadmin")
	{
		authRoutes.POST("/symbo", controllers.AddNewSymbol)
		authRoutes.PATCH("/symbo", controllers.UpdateSymbol)
		authRoutes.GET("/symbo", controllers.GetAllSymbol)
	}
}
