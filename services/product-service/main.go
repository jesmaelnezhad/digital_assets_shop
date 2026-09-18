package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/store4bots/shared/middleware"
	"github.com/store4bots/product-service/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8082" }
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "product_db" }

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, dbPort, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil { log.Fatalf("[product-service] db: %v", err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatalf("[product-service] ping: %v", err) }
	log.Println("[product-service] connected")

	h := handlers.NewHandlers(db)

	r := gin.New()
	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "product-service"}) })

	// Public
	r.GET("/api/v1/products/appearance", h.GetAppearance)
	r.GET("/api/v1/products", h.ListProducts)
	r.GET("/api/v1/products/:slug/tiers", h.GetProductTiers)
	r.GET("/api/v1/products/:slug", h.GetProduct)
	r.GET("/api/v1/categories", h.ListCategories)
	r.GET("/api/v1/bundles", h.ListBundles)
	r.GET("/api/v1/bundles/:id", h.GetBundle)
	r.GET("/api/v1/recommendations/:productId", h.GetRecommendations)
	r.POST("/api/v1/product-requests", h.CreateProductRequest)
	r.GET("/api/v1/product-requests", h.ListProductRequests)

	// Admin
	admin := r.Group("/api/v1")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		admin.POST("/products", h.CreateProduct)
		admin.PUT("/products/appearance", h.SetAppearance)
		admin.PUT("/products/banner", h.SetBanner)
		admin.PUT("/products/:id", h.UpdateProduct)
		admin.DELETE("/products/:id", h.DeleteProduct)
		admin.POST("/products/:id/images", h.AddProductImage)
		admin.DELETE("/products/:id/images/:imageId", h.DeleteProductImage)
		admin.POST("/products/:id/previews", h.GeneratePreviews)
		admin.POST("/products/:id/pin", h.PinProduct)
		admin.POST("/products/:id/unpin", h.UnpinProduct)
		admin.DELETE("/products/:id/pin", h.UnpinProduct)
		admin.POST("/products/:id/tiers", h.AddProductTier)
		admin.PUT("/products/:id/tiers/:tierId", h.UpdateProductTier)
		admin.DELETE("/products/:id/tiers/:tierId", h.DeleteProductTier)
		admin.POST("/categories", h.CreateCategory)
		admin.PUT("/categories/:id", h.UpdateCategory)
		admin.DELETE("/categories/:id", h.DeleteCategory)
		admin.POST("/bundles", h.CreateBundle)
		admin.PUT("/bundles/:id", h.UpdateBundle)
		admin.DELETE("/bundles/:id", h.DeleteBundle)
		admin.GET("/admin/product-requests", h.ListProductRequests)
		admin.PUT("/admin/product-requests/:id", h.UpdateProductRequest)
	}

	log.Printf("[product-service] starting on :%s", port)
	if err := r.Run(":" + port); err != nil { log.Fatalf("[product-service] start: %v", err) }
}
