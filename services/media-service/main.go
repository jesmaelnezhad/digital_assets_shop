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
	"github.com/pawradise/media-service/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8088" }
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "media_db" }

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, dbPort, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil { log.Fatalf("[media] db: %v", err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatalf("[media] ping: %v", err) }
	log.Println("[media] connected")

	r := gin.New()
	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "media-service"}) })

	// Admin uploads
	admin := r.Group("/api/v1")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		admin.POST("/media/upload", handlers.UploadFile)
		admin.DELETE("/media/files/:path", handlers.DeleteFile)
		admin.POST("/media/images/:productId/previews", handlers.GeneratePreview)
		admin.POST("/media/images/:productId/thumbnails", handlers.CreateThumbnail)
	}

	// File access (auth required)
	auth := r.Group("/api/v1")
	auth.Use(middleware.JwtAuthMiddleware())
	{
		auth.GET("/media/files/:path", handlers.DownloadFile)
		auth.GET("/media/files/:path/metadata", handlers.GetFileMetadata)
		auth.POST("/media/verify-access", handlers.VerifyAccess)
		auth.POST("/media/validate-type", handlers.ValidateFileType)
	}

	log.Printf("[media] starting on :%s", port)
	if err := r.Run(":" + port); err != nil { log.Fatalf("[media] start: %v", err) }
}
