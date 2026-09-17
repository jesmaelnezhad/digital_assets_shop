// Admin service needs additional routes for bundles, coupons, stats, exchange-rates
// Adding these to main.go

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/pawradise/shared/middleware"
	adminhandlers "github.com/pawradise/admin-service/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8087" }
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "admin_db" }

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, dbPort, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil { log.Fatalf("[admin] db: %v", err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatalf("[admin] ping: %v", err) }
	log.Println("[admin] connected")

	// Also connect to identity DB for user queries
	identityConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, dbPort, user, password, "appdb_identity_staging")
	identityDB, err := sql.Open("postgres", identityConnStr)
	if err != nil { log.Fatalf("[admin] identity db: %v", err) }
	defer identityDB.Close()
	if err := identityDB.Ping(); err != nil { log.Fatalf("[admin] identity ping: %v", err) }
	log.Println("[admin] connected to identity db")

	h := adminhandlers.NewAdminHandler(db, identityDB)

	r := gin.New()

	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "admin-service"}) })

	// Admin-only routes
	admin := r.Group("/api/v1")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		// Users
		admin.GET("/admin/users", h.ListUsers)
		admin.DELETE("/admin/users/:id", h.DeleteUser)
		admin.POST("/admin/users/:id/reset-password", h.ResetPassword)

		// Products
		admin.GET("/admin/products", h.ListAllProducts)
		admin.POST("/admin/products", h.CreateProduct)
		admin.PUT("/admin/products/:id", h.UpdateProduct)
		admin.DELETE("/admin/products/:id", h.DeleteProduct)
		admin.POST("/admin/products/:id/tiers", h.AddProductTier)
		admin.PUT("/admin/products/:id/tiers/:tierId", h.UpdateTier)
		admin.DELETE("/admin/products/:id/tiers/:tierId", h.DeleteTier)
		admin.POST("/admin/products/:id/images", h.AddProductImage)
		admin.DELETE("/admin/products/:id/images/:imageId", h.DeleteProductImage)
		admin.POST("/admin/products/:id/generate-previews", h.GeneratePreviews)
		admin.POST("/admin/products/:id/pin", h.PinProduct)
		admin.DELETE("/admin/products/:id/pin", h.UnpinProduct)
		admin.POST("/admin/products/bulk", h.BulkUpdateProducts)

		// Orders
		admin.GET("/admin/orders", h.ListAllOrders)
		admin.GET("/admin/orders/:id", h.GetOrderDetail)
		admin.PUT("/admin/orders/:id/status", h.UpdateOrderStatus)

		// Community
		admin.GET("/admin/community/posts", h.ListCommunityPosts)
		admin.DELETE("/admin/community/posts/:id", h.DeleteCommunityPost)

		// Settings
		admin.GET("/admin/settings", h.ListSettings)
		admin.PUT("/admin/settings/:key", h.SetSetting)

		// Referrals
		admin.GET("/admin/referrals", h.ListReferrals)

		// Export
		admin.POST("/admin/export/emails", h.ExportEmails)

		// Bundles
		admin.GET("/admin/bundles", h.ListBundles)
		admin.POST("/admin/bundles", h.CreateBundle)
		admin.PUT("/admin/bundles/:id", h.UpdateBundle)
		admin.DELETE("/admin/bundles/:id", h.DeleteBundle)

		// Coupons
		admin.GET("/admin/coupons", h.ListCoupons)
		admin.POST("/admin/coupons", h.CreateCoupon)
		admin.PUT("/admin/coupons/:id", h.UpdateCoupon)
		admin.DELETE("/admin/coupons/:id", h.DeleteCoupon)

		// Stats
		admin.GET("/admin/stats", h.GetStats)
		admin.GET("/admin/products/stats", h.GetStats)

		// Exchange rates
		admin.GET("/admin/exchange-rates", h.ListExchangeRates)
		admin.PUT("/admin/exchange-rates/:chain", h.SetExchangeRate)
		admin.DELETE("/admin/exchange-rates/:chain", h.DeleteExchangeRate)
	}

	log.Printf("[admin] starting on :%s", port)
	if err := r.Run(":" + port); err != nil { log.Fatalf("[admin] start: %v", err) }
}
