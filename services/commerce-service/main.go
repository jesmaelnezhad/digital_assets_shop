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
	"github.com/store4bots/commerce-service/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8083" }
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "commerce_db" }

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, dbPort, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil { log.Fatalf("[commerce] db: %v", err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatalf("[commerce] ping: %v", err) }
	log.Println("[commerce] connected")

	h := handlers.NewCommerceHandler(db)

	r := gin.New()
	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "commerce-service"}) })

	// Public
	r.POST("/api/v1/guest-orders", h.CreateGuestOrder)
	r.GET("/api/v1/guest-orders", h.GetGuestOrders)
	r.GET("/api/v1/guest-orders/:id", h.GetGuestOrder)
	r.POST("/api/v1/coupons/validate", h.ValidateCoupon)

	// Auth
	auth := r.Group("/api/v1")
	auth.Use(middleware.JwtAuthMiddleware())
	{
		auth.POST("/orders", h.CreateOrder)
		auth.GET("/orders", h.GetUserOrders)
		auth.GET("/orders/:id", h.GetOrder)
		auth.GET("/orders/:id/payment", h.GetOrderPayment)
		auth.GET("/orders/:id/status", h.CheckPaymentStatus)
		auth.GET("/orders/:id/download/:itemId", h.GetOrderDownload)
		auth.POST("/orders/:id/confirm", h.ConfirmPayment)
		auth.GET("/cart", h.GetCart)
		auth.POST("/cart/items", h.AddCartItem)
		auth.PUT("/cart/items/:id", h.UpdateCartItem)
		auth.DELETE("/cart/items/:id", h.RemoveCartItem)
		auth.GET("/wishlist", h.GetWishlist)
		auth.POST("/wishlist/toggle", h.ToggleWishlist)
		auth.POST("/recently-viewed", h.RecordView)
		auth.GET("/recently-viewed", h.GetRecentlyViewed)
		auth.POST("/compare/toggle", h.ToggleCompare)
		auth.GET("/compare", h.GetCompare)
		auth.POST("/coupons", h.CreateCoupon)
		auth.PUT("/coupons/:id", h.UpdateCoupon)
		auth.DELETE("/coupons/:id", h.DeleteCoupon)
	}

	log.Printf("[commerce] starting on :%s", port)
	if err := r.Run(":" + port); err != nil { log.Fatalf("[commerce] start: %v", err) }
}
