package routes

import (
	"TradingSystem/src/controllers"
	"TradingSystem/src/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterCustomerRoutes(r *gin.Engine) {
	customerRoutes := r.Group("/customers")
	customerRoutes.Use(middleware.CustomerMiddleware())
	{
		customerRoutes.GET("/availableamount", controllers.GetCustomerBalance)
		customerRoutes.GET("/placeorderhistory", controllers.PlaceOrderHistory) //下單記錄細項, Symbol, Cid,
		customerRoutes.GET("/getplaceorderhistory", controllers.GetPlaceOrderHistoryBySymbol)
		customerRoutes.GET("/dashboard", controllers.ShowDashboardPage)
		customerRoutes.GET("/symbol", controllers.GetAllCustomerSymbol)
		customerRoutes.POST("", controllers.CreateCustomer)
		customerRoutes.POST("/update", controllers.UpdateCustomer)
		customerRoutes.PATCH("/symbol", controllers.UpdateCustomerSymbol)
		customerRoutes.GET("/linktg", controllers.GetTGBot)
		customerRoutes.GET("/weeklyreportlist", controllers.CustomerWeeklyReportList)
		customerRoutes.GET("/weeklyreportlistsummary", controllers.CustomerWeeklyReportSummaryList) //週報表, Cid
		customerRoutes.GET("/monthlyreportlistsummary", controllers.CustomerMonthlyReportSummaryList)
		customerRoutes.GET("/monthlyreportlist", controllers.CustomerMonthlyReportList)
	}
}
