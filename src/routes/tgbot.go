package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterTGRoutes(r *gin.Engine) {
	authRoutes := r.Group("/tg")
	{
		authRoutes.POST("/whxuygsg", controllers.TGbot)
	}

}
