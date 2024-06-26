package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	RegisterAuthRoutes(r)
	RegisterCustomerRoutes(r)
	RegisterBeforeLoginRoutes(r)
	RegisterMiscRoutes(r)
	RegisterWebhookRoutes(r)
	RegisterRestAdminRoutes(r)
	RegisterMyTestRoutes(r)
}
