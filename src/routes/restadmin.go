package routes

import (
	"ManageAPI/src/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRestAdminRoutes(r *gin.Engine) {
	authRoutes := r.Group("/restadmin")
	{
		authRoutes.POST("/symbo", controllers.AddNewSymbo)
		authRoutes.PATCH("/symbo", controllers.UpdateSymbo)
		authRoutes.GET("/symbo", controllers.GetAllSymbo)
	}
}
