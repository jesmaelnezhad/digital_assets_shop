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
		"github.com/store4bots/shared/auth"
		"github.com/store4bots/payment-service/handlers"
	)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8086" }
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "payment_db" }

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, dbPort, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil { log.Fatalf("[payment] db: %v", err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatalf("[payment] ping: %v", err) }
	auth.SetRevocationDB(db)
	log.Println("[payment] connected")

	h := handlers.NewPaymentHandler(db)

	r := gin.New()
	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "payment-service"}) })

	// Public
	r.GET("/api/v1/exchange-rates", h.GetExchangeRates)
	r.GET("/api/v1/exchange-rates/:chain", h.GetExchangeRate)
	r.GET("/api/v1/settings/:key", h.GetSetting)

	// Auth - payment ops
	authGroup := r.Group("/api/v1")
	authGroup.Use(middleware.JwtAuthMiddleware())
	{
		authGroup.GET("/payments/order/:orderId", h.GetPaymentDetails)
		authGroup.GET("/payments/order/:orderId/status", h.CheckPaymentStatus)
		authGroup.POST("/payments/order/:orderId/confirm", h.ConfirmPayment)
	}

	// Admin
	admin := r.Group("/api/v1")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		admin.PUT("/exchange-rates/:chain", h.SetExchangeRate)
		admin.DELETE("/exchange-rates/:chain", h.DeleteExchangeRate)
		admin.GET("/settings", h.GetSettings)
		admin.PUT("/settings/:key", h.SetSetting)
	}

	log.Printf("[payment] starting on :%s", port)
	if err := r.Run(":" + port); err != nil { log.Fatalf("[payment] start: %v", err) }
}
