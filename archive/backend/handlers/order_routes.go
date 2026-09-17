package handlers

import "github.com/gin-gonic/gin"

// RegisterOrderRoutes registers order-related routes
func RegisterOrderRoutes(r *gin.RouterGroup) {
	user := r.Group("/orders")
	user.Use(JwtAuthMiddleware())
	{
		user.POST("", CreateOrder)
		user.GET("", GetUserOrders)
		user.GET("/:id", GetOrder)
		user.GET("/:id/payment", GetOrderPayment)
		user.GET("/:id/status", CheckPaymentStatus)
		user.GET("/download/:id", GetOrderDownload)
	}
}
