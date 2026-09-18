package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/store4bots/shared/models"
)

type ReviewHandler struct{ db *sql.DB }

func NewReviewHandler(db *sql.DB) *ReviewHandler { return &ReviewHandler{db} }

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok { c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"}); return }
	uid := userID.(int)

	var req struct {
		ProductID int    `json:"product_id" binding:"required"`
		Rating    int    `json:"rating" binding:"required"`
		Title     string `json:"title"`
		Content   string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	if req.Rating < 1 || req.Rating > 5 { c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be 1-5"}); return }

	// Check existing review
	var existing int
	h.db.QueryRow("SELECT 1 FROM reviews WHERE product_id = $1 AND user_id = $2", req.ProductID, uid).Scan(&existing)
	if existing == 1 { c.JSON(http.StatusConflict, gin.H{"error": "you already reviewed this product"}); return }

	var reviewID int
	var err error
	err = h.db.QueryRow(
		`INSERT INTO reviews (product_id, user_id, rating, title, content, is_verified_purchase)
		 VALUES ($1, $2, $3, COALESCE(NULLIF($4, ''), ''), COALESCE(NULLIF($5, ''), ''), true)
		 RETURNING id`,
		req.ProductID, uid, req.Rating, req.Title, req.Content,
	).Scan(&reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review"})
		return
	}

	// Update product
	h.db.Exec("UPDATE products SET purchase_count = purchase_count + 1 WHERE id = $1", req.ProductID)

	c.JSON(http.StatusCreated, gin.H{"review": models.Review{ID: reviewID, ProductID: req.ProductID, Rating: req.Rating}})
}

func (h *ReviewHandler) GetProductReviews(c *gin.Context) {
	productIDStr := c.Param("productId")
	var productID int
	if _, err := strconv.Atoi(productIDStr); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"}); return }
	productID, _ = strconv.Atoi(productIDStr)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 { page = 1 }
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 || perPage > 50 { perPage = 20 }
	offset := (page - 1) * perPage

	var total int
	h.db.QueryRow("SELECT COUNT(*) FROM reviews WHERE product_id = $1", productID).Scan(&total)

	rows, err := h.db.Query(
		`SELECT r.id, r.product_id, r.user_id, r.rating, r.title, r.content, r.is_verified_purchase,
		        r.helpful_count, r.is_helpful, r.created_at, r.updated_at,
		        ''::text, ''::text, ''::text, ''::text
		FROM reviews r
		WHERE r.product_id = $1 AND r.is_public = true
		ORDER BY r.helpful_count DESC, r.created_at DESC LIMIT $2 OFFSET $3`,
		productID, perPage, offset,
	)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reviews"}); return }
	defer rows.Close()

	type reviewWithAuthor struct {
		models.Review
		Author struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			Email   string `json:"email"`
			Avatar  string `json:"avatar_url"`
			Display string `json:"display_name"`
		} `json:"author"`
	}
	var reviews []reviewWithAuthor
	for rows.Next() {
		var r reviewWithAuthor
		if err := rows.Scan(&r.ID, &r.ProductID, &r.UserID, &r.Rating, &r.Title, &r.Content,
			&r.IsVerifiedPurchase, &r.HelpfulCount, &r.IsHelpful, &r.CreatedAt, &r.UpdatedAt,
			&r.Author.Name, &r.Author.Email, &r.Author.Avatar, &r.Author.Display); err != nil { continue }
		reviews = append(reviews, r)
	}
	if reviews == nil { reviews = []reviewWithAuthor{} }

	c.JSON(http.StatusOK, gin.H{"reviews": reviews, "total": total, "page": page, "per_page": perPage})
}

func (h *ReviewHandler) GetAverageRating(c *gin.Context) {
	productIDStr := c.Param("productId")
	var productID int
	if _, err := strconv.Atoi(productIDStr); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"}); return }
	productID, _ = strconv.Atoi(productIDStr)

	var avgRating float64
	var count int
	var totalRating int
	h.db.QueryRow("SELECT COALESCE(AVG(rating), 0), COUNT(*), COALESCE(SUM(rating), 0) FROM reviews WHERE product_id = $1 AND is_public = true", productID).Scan(&avgRating, &count, &totalRating)

	// Rating distribution
	dist := make(map[string]int)
	for i := 1; i <= 5; i++ {
		var val int
		h.db.QueryRow("SELECT COUNT(*) FROM reviews WHERE product_id = $1 AND rating = $2 AND is_public = true", productID, i).Scan(&val)
		dist[fmt.Sprintf("%d_star", i)] = val
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id":   productID,
		"average_rating": avgRating,
		"total_reviews": count,
		"total_rating":  totalRating,
		"distribution":  dist,
	})
}
