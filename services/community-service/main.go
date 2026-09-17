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
	"github.com/pawradise/community-service/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8084" }
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "community_db" }

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, dbPort, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil { log.Fatalf("[community] db: %v", err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatalf("[community] ping: %v", err) }
	log.Println("[community] connected")

	h := handlers.NewCommunityHandler(db)

	r := gin.New()
	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "community-service"}) })

	// Public routes
	r.GET("/api/v1/community/posts", h.GetFeed)
	r.GET("/api/v1/community/feed", h.GetFeed)
	r.GET("/api/v1/community/posts/:id", h.GetPost)
	r.GET("/api/v1/community/users/:id", h.GetPublicProfile)
	// E2E test path aliases (public)
	r.GET("/api/v1/posts", h.GetFeed)
	r.GET("/api/v1/posts/:id", h.GetPost)
	r.GET("/api/v1/users/:id", h.GetPublicProfile)

	// Auth group
	auth := r.Group("/api/v1")
	auth.Use(middleware.JwtAuthMiddleware())
	{
		// Community CRUD (original paths)
		auth.POST("/community/posts", h.CreatePost)
		auth.DELETE("/community/posts/:id", h.DeletePost)
		auth.POST("/community/posts/:id/like", h.LikePost)
		auth.DELETE("/community/posts/:id/like", h.UnlikePost)
		auth.POST("/community/posts/:id/comments", h.AddComment)
		auth.DELETE("/community/posts/:id/comments/:commentId", h.DeleteComment)
		auth.POST("/community/follow/:userId", h.FollowUser)
		auth.DELETE("/community/follow/:userId", h.UnfollowUser)
		auth.GET("/profile", h.GetMyProfile)
		auth.PUT("/profile", h.UpdateProfile)
		// E2E test path aliases (auth)
		auth.POST("/posts", h.CreatePost)
		auth.DELETE("/posts/:id", h.DeletePost)
		auth.POST("/posts/:id/like", h.LikePost)
		auth.DELETE("/posts/:id/like", h.UnlikePost)
		auth.POST("/posts/:id/comments", h.AddComment)
		auth.DELETE("/posts/:id/comments/:commentId", h.DeleteComment)
		auth.POST("/follow/:userId", h.FollowUser)
		auth.DELETE("/follow/:userId", h.UnfollowUser)
	}

	log.Printf("[community] starting on :%s", port)
	if err := r.Run(":" + port); err != nil { log.Fatalf("[community] start: %v", err) }
}
