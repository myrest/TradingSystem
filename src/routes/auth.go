package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine) {
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/google", controllers.GoogleAuthCallback)
		authRoutes.GET("/google", controllers.GoogleLogout)
	}
}
