package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterWebhookRoutes(r *gin.Engine) {
	authRoutes := r.Group("/webhook")
	{
		authRoutes.POST("/tradingview", controllers.TradingViewWebhook)
	}
}
