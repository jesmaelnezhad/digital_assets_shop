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
	"github.com/pawradise/review-service/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8085" }
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "review_db" }

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, dbPort, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil { log.Fatalf("[review] db: %v", err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatalf("[review] ping: %v", err) }
	log.Println("[review] connected")

	h := handlers.NewReviewHandler(db)

	r := gin.New()
	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "review-service"}) })

	// Public
	r.GET("/api/v1/reviews/:productId", h.GetProductReviews)
	r.GET("/api/v1/reviews/:productId/average", h.GetAverageRating)

	// Auth
	auth := r.Group("/api/v1")
	auth.Use(middleware.JwtAuthMiddleware())
	{
		auth.POST("/reviews", h.CreateReview)
	}

	log.Printf("[review] starting on :%s", port)
	if err := r.Run(":" + port); err != nil { log.Fatalf("[review] start: %v", err) }
}
