package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterBeforeLoginRoutes(r *gin.Engine) {
	authRoutes := r.Group("/")
	{
		authRoutes.GET("/", controllers.ShowLoginPage)
		authRoutes.GET("/login", controllers.ShowLoginPage)
	}
}
