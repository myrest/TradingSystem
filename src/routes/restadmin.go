package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRestAdminRoutes(r *gin.Engine) {
	authRoutes := r.Group("/restadmin")
	{
		authRoutes.POST("/symbol", controllers.AddNewSymbol)
		authRoutes.PATCH("/symbol", controllers.UpdateSymbol)
		authRoutes.GET("/symbol", controllers.GetAllSymbol)
	}
}
