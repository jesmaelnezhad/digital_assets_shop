package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"backend/database"
)

func setupTestEnv(t *testing.T) (func(), sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	database.DB = mockDB
	cleanup := func() {
		mockDB.Close()
	}
	return cleanup, mock
}

func R(method, path string, body interface{}, headers map[string]string) (int, []byte, map[string]interface{}) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Public auth routes (no JWT needed)
	r.POST("/api/v1/register", NewAuthHandler(nil).Register)
	r.POST("/api/v1/login", NewAuthHandler(nil).Login)

	// Protected routes with JWT middleware
	v1 := r.Group("/api/v1")
	v1.Use(JwtAuthMiddleware())
	{
		// Auth
		v1.GET("/me", NewAuthHandler(nil).GetProfile)
		v1.PUT("/me", NewAuthHandler(nil).UpdateProfile)
		v1.POST("/logout", NewAuthHandler(nil).Logout)

		// Community
		community := v1.Group("/community")
		{
			community.GET("/feed", GetFeed)
			community.POST("/posts", CreatePost)
			community.GET("/posts/:id", GetPost)
			community.POST("/posts/:id/like", LikePost)
			community.DELETE("/posts/:id/like", UnlikePost)
			community.POST("/posts/:id/comments", AddComment)
			community.DELETE("/posts/:id/comments/:commentId", DeleteComment)
			community.POST("/follow/:userId", FollowUser)
			community.DELETE("/follow/:userId", UnfollowUser)
			community.GET("/users/:id", GetPublicProfile)
		}

		// Profile
		profile := v1.Group("/profile")
		{
			profile.GET("", GetMyProfile)
			profile.PUT("", UpdateProfile)
		}

		// Orders
		v1.POST("/orders", CreateOrder)
		v1.GET("/orders", GetUserOrders)
		v1.GET("/orders/:id", GetOrder)
		v1.GET("/orders/:id/payment", GetOrderPayment)
		v1.GET("/orders/:id/status/check", CheckPaymentStatus)

		// Settings (public)
		v1.GET("/settings/:key", GetSetting)

		// Exchange rates (public)
		v1.GET("/exchange-rates", ListExchangeRates)
		v1.GET("/exchange-rates/:chain", GetExchangeRate)

		// Guest orders (public)
		v1.POST("/guest-orders", CreateGuestOrder)
		v1.GET("/guest-orders/:id", GetGuestOrder)

		// Feature routes (cart, wishlist, reviews, recently-viewed, compare, recommendations)
		features := v1.Group("")
		{
			RegisterReviewRoutes(features)
			RegisterCartRoutes(features)
			RegisterWishlistRoutes(features)
			RegisterRecentlyViewedRoutes(features)
			RegisterComparisonRoutes(features)
			RegisterRecommendationRoutes(features)
		}
	}

	// Admin routes with admin auth
	admin := r.Group("/api/v1/admin")
	admin.Use(AdminAuthMiddleware())
	{
		admin.GET("/users", NewAuthHandler(nil).ListUsers)
		admin.DELETE("/users/:id", NewAuthHandler(nil).DeleteUser)
		admin.POST("/users/:id/reset-password", NewAuthHandler(nil).ResetPassword)

		// Product admin
		admin.POST("/products", CreateProduct)
		admin.PUT("/products/:id", UpdateProduct)
		admin.DELETE("/products/:id", DeleteProduct)
		admin.GET("/products/stats", GetProductStats)

		// Exchange rate admin
		admin.GET("/exchange-rates", ListExchangeRates)
		admin.GET("/exchange-rates/:chain", GetExchangeRate)
		admin.PUT("/exchange-rates/:chain", SetExchangeRate)
		admin.DELETE("/exchange-rates/:chain", DeleteExchangeRate)

		// Settings admin
		admin.GET("/settings", ListSettings)
		admin.PUT("/settings/:key", SetSetting)

		// Bulk, order status, community moderation, guest orders
		RegisterAdminFeaturesRoutes(admin)
		RegisterOrderStatusRoutes(admin)
		RegisterCommunityModerationRoutes(admin)
		RegisterGuestOrderRoutes(admin)

		// Admin order listing
		admin.GET("/orders", ListAllOrders)
		admin.GET("/orders/:id/payment", GetOrderPayment)
	}

	// Public routes via Register* functions
	RegisterProductRoutes(r.Group("/api/v1"))
	RegisterExchangeRateRoutes(r.Group("/api/v1"))
	RegisterGuestOrderRoutes(r.Group("/api/v1"))

	// Health
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, map[string]string{"status": "ok"})
	})
	r.GET("/api/v1/", func(c *gin.Context) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	var bodyBytes []byte
	if body != nil {
		switch v := body.(type) {
		case []byte:
			bodyBytes = v
		case string:
			bodyBytes = []byte(v)
		default:
			var err error
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				panic(err)
			}
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	result := make(map[string]interface{})
	if rec.Body.Len() > 0 {
		json.Unmarshal(rec.Body.Bytes(), &result)
	}
	return rec.Code, rec.Body.Bytes(), result
}

func makeJWT(userID int, email string) string {
	secret := "your-secret-key-change-in-production"
	if v := os.Getenv("JWT_SECRET"); v != "" {
		secret = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func makeAdminToken() string {
	return "admin-secret-token-change-in-production"
}

func mockTime(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func expectExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
