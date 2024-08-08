package routes

import (
	"TradingSystem/src/common"
	"TradingSystem/src/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

var sectestword = common.GenerateRandomString(8)

func RegisterMyTestRoutes(r *gin.Engine) {

	authRoutes := r.Group(sectestword)
	{
		authRoutes.GET("/getbyid", controllers.GetBingxOrderByID)
		authRoutes.GET("/getavailablebalance", controllers.GetAvailableAmountByID)
	}

	specialrouter := r.Group("/resthome")
	{
		specialrouter.PATCH("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": sectestword})
		})
	}

}
