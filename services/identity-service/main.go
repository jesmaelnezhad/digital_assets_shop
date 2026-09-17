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
	"github.com/pawradise/shared/auth"
	"github.com/pawradise/identity-service/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8081" }

	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "identity_db" }

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, dbPort, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("[identity-service] db open: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("[identity-service] db ping: %v", err)
	}
	log.Println("[identity-service] connected to database")

	auth.SetRevocationDB(db)

	r := gin.New()
	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "identity-service"})
	})
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "identity-service"})
	})

	ah := handlers.NewAuthHandler(db)

	// Public routes
	r.POST("/api/v1/register", ah.Register)
	r.POST("/api/v1/login", ah.Login)

	// Auth-required routes
	authGroup := r.Group("/api/v1")
	authGroup.Use(middleware.JwtAuthMiddleware())
	{
		authGroup.POST("/logout", ah.Logout)
		authGroup.GET("/me", ah.GetProfile)
		authGroup.PUT("/me", ah.UpdateProfile)
		authGroup.GET("/profile/:id", ah.GetUserByID)
		authGroup.GET("/referrals", ah.GetReferrals)
		authGroup.GET("/commissions", ah.GetCommissions)
		authGroup.GET("/referrals/earnings", ah.GetCommissions)
		authGroup.POST("/referrals/track", ah.TrackReferral)
		authGroup.POST("/commissions/record", ah.RecordCommission)
	}

	// Admin routes
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(middleware.AdminAuthMiddleware())
	{
		adminGroup.GET("/users", ah.ListUsers)
		adminGroup.DELETE("/users/:id", ah.DeleteUser)
		adminGroup.POST("/users/:id/reset-password", ah.ResetPassword)
	}

	log.Printf("[identity-service] starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("[identity-service] failed to start: %v", err)
	}
}
