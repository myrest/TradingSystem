package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterMyTestRoutes(r *gin.Engine) {
	authRoutes := r.Group("/test")
	{
		authRoutes.GET("/getbyid", controllers.GetBingxOrderByID)
		authRoutes.GET("/t1", controllers.TEST)
		authRoutes.GET("/t2", controllers.TESTBet)
		authRoutes.GET("/t3", controllers.TESTGetOpenOrder)
	}
}
