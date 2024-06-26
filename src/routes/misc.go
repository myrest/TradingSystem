package routes

import (
	"ManageAPI/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterMiscRoutes(r *gin.Engine) {
	authRoutes := r.Group("/misc")
	{
		authRoutes.GET("/fireAuthConfig", controllers.FireAuthConfig)
	}
}
