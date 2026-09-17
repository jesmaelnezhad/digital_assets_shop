package main

import (
	"backend/database"
	"backend/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.CreateTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	authHandler := handlers.NewAuthHandler(db)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/register", authHandler.Register)
		v1.POST("/login", authHandler.Login)
		v1.GET("/me", handlers.JwtAuthMiddleware(), authHandler.GetProfile)
		v1.PUT("/me", handlers.JwtAuthMiddleware(), authHandler.UpdateProfile)
		v1.POST("/logout", handlers.JwtAuthMiddleware(), authHandler.Logout)

		// Phase 1: Product catalog
		handlers.RegisterProductRoutes(v1)

		// Phase 2: Orders & payments
		handlers.RegisterOrderRoutes(v1)

		// Phase 3: Community center
		handlers.RegisterCommunityRoutes(v1)

		// Phase 4+: Feature routes that require JWT auth
		features := v1.Group("")
		features.Use(handlers.JwtAuthMiddleware())
		{
			handlers.RegisterReviewRoutes(features)
			handlers.RegisterCartRoutes(features)
			handlers.RegisterWishlistRoutes(features)
			handlers.RegisterRecentlyViewedRoutes(features)
			handlers.RegisterComparisonRoutes(features)
			handlers.RegisterRecommendationRoutes(features)
		}
	}

	if os.Getenv("ADMIN_ENABLED") == "true" {
		admin := v1.Group("/admin")
		admin.Use(handlers.AdminAuthMiddleware())
		{
			admin.GET("/users", authHandler.ListUsers)
			admin.DELETE("/users/:id", authHandler.DeleteUser)
			admin.POST("/users/:id/reset-password", authHandler.ResetPassword)

			// Product admin
			admin.POST("/products", handlers.CreateProduct)
			admin.PUT("/products/:id", handlers.UpdateProduct)
			admin.DELETE("/products/:id", handlers.DeleteProduct)
			admin.GET("/products/stats", handlers.GetProductStats)

			// Exchange rate admin (Phase 2)
			admin.GET("/exchange-rates", handlers.ListExchangeRates)
			admin.GET("/exchange-rates/:chain", handlers.GetExchangeRate)
			admin.PUT("/exchange-rates/:chain", handlers.SetExchangeRate)
			admin.DELETE("/exchange-rates/:chain", handlers.DeleteExchangeRate)

			// Settings admin (seller wallet, etc.)
			admin.GET("/settings", handlers.ListSettings)
			admin.PUT("/settings/:key", handlers.SetSetting)

			// Admin: bulk product actions (feature 7)
			handlers.RegisterAdminFeaturesRoutes(admin)

			// Admin: order status workflow (feature 8)
			handlers.RegisterOrderStatusRoutes(admin)

			// Admin: community moderation (feature 8)
			handlers.RegisterCommunityModerationRoutes(admin)

			// Admin: guest order creation (feature 10)
			handlers.RegisterGuestOrderRoutes(admin)
			}
	}

	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/api/v1/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
