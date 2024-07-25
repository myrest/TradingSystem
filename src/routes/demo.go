package routes

import (
	"TradingSystem/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterDemoRoutes(r *gin.Engine) {
	authRoutes := r.Group("/demo")
	{
		authRoutes.GET("/", controllers.DemoList)
	}
}
