package routes

import (
	"TradingSystem/src/common"
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterMyTestRoutes(r *gin.Engine) {
	settings := common.GetEnvironmentSetting()
	if settings.Env == common.Dev {
		authRoutes := r.Group("/test")
		{
			authRoutes.GET("/getbyid", controllers.GetBingxOrderByID)
			authRoutes.GET("/getavailablebalance", controllers.GetAvailableAmountByID)
			authRoutes.GET("/t2", controllers.TEST2)
			authRoutes.GET("/t3", controllers.TEST3)
		}
	}
}
