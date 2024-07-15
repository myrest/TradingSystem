package routes

import (
	"TradingSystem/src/controllers"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterMyTestRoutes(r *gin.Engine) {
	firebaseKey := os.Getenv("ENVIRONMENT")
	if firebaseKey != "" && strings.ToLower(firebaseKey) == "dev" {
		authRoutes := r.Group("/test")
		{
			authRoutes.GET("/getbyid", controllers.GetBingxOrderByID)
			authRoutes.GET("/t3", controllers.TESTGetOpenOrder)
		}
	}
}
