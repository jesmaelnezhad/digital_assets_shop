package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/events-service/handlers"
	"github.com/pawradise/shared/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://127.0.0.1:27017"
	}
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "events"
	}

	var client *mongo.Client
	var err error
	for i := 0; i < 12; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err == nil {
			err = client.Ping(ctx, nil)
		}
		cancel()
		if err == nil {
			break
		}
		log.Printf("[events-service] waiting for mongodb (%d/12): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("[events-service] mongo connect: %v", err)
	}
	log.Println("[events-service] connected to mongodb")

	store := handlers.NewMongoStore(client.Database(dbName))
	h := handlers.New(store)

	r := gin.New()
	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "events-service"})
	})
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "events-service"})
	})

	ingest := r.Group("/api/v1")
	ingest.Use(middleware.JwtAuthMiddleware())
	{
		ingest.POST("/events", h.Ingest)
	}

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		admin.GET("/events", h.List)
		admin.GET("/events/ttl", h.GetTTL)
		admin.PUT("/events/ttl", h.SetTTL)
	}

	log.Printf("[events-service] starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("[events-service] failed to start: %v", err)
	}
}
