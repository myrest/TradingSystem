package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	//測試用，正常要關起來。
	//RegisterMyTestRoutes(r)
	RegisterAuthRoutes(r)
	RegisterCustomerRoutes(r)
	RegisterBeforeLoginRoutes(r)
	RegisterMiscRoutes(r)
	RegisterWebhookRoutes(r)
	RegisterRestAdminRoutes(r)
}
