package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"backend/database"
	"github.com/gin-gonic/gin"
)

// ---------- Cart ----------

type CartHandler struct{ db *sql.DB }
type CartItemHandler struct{ db *sql.DB }

func NewCartHandler(db *sql.DB) *CartHandler {
	if db == nil {
		db = database.DB
	}
	return &CartHandler{db}
}
func NewCartItemHandler(db *sql.DB) *CartItemHandler {
	if db == nil {
		db = database.DB
	}
	return &CartItemHandler{db}
}

// GetOrCreateCart returns or creates a cart for the user
func (h *CartHandler) GetOrCreateCart(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var cartID int
	err := h.db.QueryRow("SELECT id FROM carts WHERE user_id=$1", uid).Scan(&cartID)
	if err == sql.ErrNoRows {
		result := h.db.QueryRow("INSERT INTO carts (user_id) VALUES ($1) RETURNING id", uid)
		err = result.Scan(&cartID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cart"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"cart_id": cartID,
		"message": "cart ready",
	})
}

// AddToCart adds or updates an item in the cart
func (h *CartItemHandler) AddToCart(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID == 0 || req.Quantity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var cartID int
	err := h.db.QueryRow("SELECT id FROM carts WHERE user_id=$1", uid).Scan(&cartID)
	if err == sql.ErrNoRows {
		result := h.db.QueryRow("INSERT INTO carts (user_id) VALUES ($1) RETURNING id", uid)
		err = result.Scan(&cartID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cart"})
			return
		}
	}

	var stock, priceUsd int
	err = h.db.QueryRow("SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id=$1", req.ProductID).Scan(&stock, &priceUsd)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if req.Quantity > stock {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exceeds stock"})
		return
	}

	// Upsert
	h.db.Exec("INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = $3, added_at = CURRENT_TIMESTAMP",
		cartID, req.ProductID, req.Quantity)

	// total items
	var totalItems int
	h.db.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE cart_id=$1", cartID).Scan(&totalItems)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"cart_id":     cartID,
		"total_items": totalItems,
	})
}

// GetCart returns the full cart
func (h *CartItemHandler) GetCart(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var cartID int
	err := h.db.QueryRow("SELECT id FROM carts WHERE user_id=$1", uid).Scan(&cartID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"items": []interface{}{}, "total": 0, "total_usd": "0.00"})
		return
	}

	rows, err := h.db.Query(`
		SELECT ci.product_id, ci.quantity, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd,
		       COALESCE(p.stock_quantity, 0) as stock,
		       p.asset_path, p.asset_hash, p.file_mime_type
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1`, cartID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cart"})
		return
	}
	defer rows.Close()

	type CartItem struct {
		ProductID    int    `json:"product_id"`
		Quantity     int    `json:"quantity"`
		Title        string `json:"title"`
		Slug         string `json:"slug"`
		UnitPriceUSD string `json:"unit_price_usd"`
		Stock        int    `json:"stock"`
		AssetPath   string `json:"asset_path"`
		AssetHash   string `json:"asset_hash"`
		FileMimeType string `json:"file_mime_type"`
	}
	var items []CartItem
	var totalUsd float64
	for rows.Next() {
		var ci CartItem
		if err := rows.Scan(&ci.ProductID, &ci.Quantity, &ci.Title, &ci.Slug, &ci.UnitPriceUSD, &ci.Stock, &ci.AssetPath, &ci.AssetHash, &ci.FileMimeType); err != nil {
			continue
		}
		items = append(items, ci)
		var p float64
		fmt.Sscanf(ci.UnitPriceUSD, "%f", &p)
		totalUsd += float64(ci.Quantity) * p
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     items,
		"total":     len(items),
		"total_usd": fmt.Sprintf("%.2f", totalUsd),
	})
}

// RemoveFromCart removes an item
func (h *CartItemHandler) RemoveFromCart(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		ProductID int `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var cartID int
	h.db.QueryRow("SELECT id FROM carts WHERE user_id=$1", uid).Scan(&cartID)

	res, err := h.db.Exec("DELETE FROM cart_items WHERE cart_id=$1 AND product_id=$2", cartID, req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove"})
		return
	}
	n, _ := res.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"removed": n > 0})
}

// ClearCart clears all items
func (h *CartItemHandler) ClearCart(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var cartID int
	h.db.QueryRow("SELECT id FROM carts WHERE user_id=$1", uid).Scan(&cartID)
	h.db.Exec("DELETE FROM cart_items WHERE cart_id=$1", cartID)
	c.JSON(http.StatusOK, gin.H{"cleared": true})
}

func RegisterCartRoutes(group *gin.RouterGroup) {
	h := NewCartHandler(nil)
	ci := NewCartItemHandler(nil)
	group.GET("/cart", h.GetOrCreateCart)
	group.GET("/cart/items", ci.GetCart)
	group.POST("/cart/add", ci.AddToCart)
	group.DELETE("/cart/remove", ci.RemoveFromCart)
	group.DELETE("/cart/clear", ci.ClearCart)
}

// ---------- Wishlist ----------

type WishlistHandler struct{ db *sql.DB }

func NewWishlistHandler(db *sql.DB) *WishlistHandler {
	if db == nil {
		db = database.DB
	}
	return &WishlistHandler{db}
}

// ToggleWishlist adds or removes from wishlist
func (h *WishlistHandler) ToggleWishlist(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		ProductID int `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id=$1 AND product_id=$2)", uid, req.ProductID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if exists {
		h.db.Exec("DELETE FROM wishlists WHERE user_id=$1 AND product_id=$2", uid, req.ProductID)
		c.JSON(http.StatusOK, gin.H{"in_wishlist": false})
	} else {
		h.db.Exec("INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING", uid, req.ProductID)
		c.JSON(http.StatusOK, gin.H{"in_wishlist": true})
	}
}

// WishlistList returns user's wishlist
func (h *WishlistHandler) WishlistList(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.db.Query(`
		SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity,
		       COALESCE(pi.url, '') as primary_image
		FROM wishlists w
		JOIN products p ON p.id = w.product_id
		LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load wishlist"})
		return
	}
	defer rows.Close()

	type WishlistItem struct {
		ID           int    `json:"id"`
		Title        string `json:"title"`
		Slug         string `json:"slug"`
		PriceUSD     string `json:"price_usd"`
		Stock        int    `json:"stock"`
		PrimaryImage string `json:"primary_image"`
	}
	var items []WishlistItem
	for rows.Next() {
		var wi WishlistItem
		rows.Scan(&wi.ID, &wi.Title, &wi.Slug, &wi.PriceUSD, &wi.Stock, &wi.PrimaryImage)
		items = append(items, wi)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"count": len(items),
	})
}

func RegisterWishlistRoutes(group *gin.RouterGroup) {
	h := NewWishlistHandler(nil)
	group.POST("/wishlist/toggle", h.ToggleWishlist)
	group.GET("/wishlist", h.WishlistList)
}

// ---------- Reviews ----------

type ReviewHandler struct{ db *sql.DB }

func NewReviewHandler(db *sql.DB) *ReviewHandler {
	if db == nil {
		db = database.DB
	}
	return &ReviewHandler{db}
}

// CreateReview creates or updates a review for a product
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be 1-5"})
		return
	}

	// Product must exist
	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id=$1)", productID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	// Upsert
	h.db.Exec(`INSERT INTO reviews (product_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`,
		productID, uid, req.Rating, req.Comment)

	var review struct {
		ID        int    `json:"id"`
		Rating    int    `json:"rating"`
		Comment   string `json:"comment"`
		CreatedAt string `json:"created_at"`
	}
	h.db.QueryRow(`SELECT id, rating, COALESCE(comment, ''), created_at FROM reviews WHERE product_id=$1 AND user_id=$2`,
		productID, uid).Scan(&review.ID, &review.Rating, &review.Comment, &review.CreatedAt)

	c.JSON(http.StatusCreated, gin.H{
		"review":  review,
		"message": "review saved",
	})
}

// ReviewList returns paginated reviews for a product
func (h *ReviewHandler) ReviewList(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if page < 1 { page = 1 }
	if perPage < 1 { perPage = 10 }
	if perPage > 50 { perPage = 50 }
	offset := (page - 1) * perPage

	var total int
	h.db.QueryRow("SELECT COUNT(*) FROM reviews WHERE product_id=$1", productID).Scan(&total)

	rows, err := h.db.Query(`
		SELECT r.id, r.rating, r.comment, r.created_at,
		       u.id, u.name, u.email,
		       COALESCE(up.avatar_url, '') as avatar_url
		FROM reviews r
		JOIN users u ON u.id = r.user_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		WHERE r.product_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`, productID, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load reviews"})
		return
	}
	defer rows.Close()

	type ReviewItem struct {
		ID         int    `json:"id"`
		Rating     int    `json:"rating"`
		Comment    string `json:"comment"`
		CreatedAt  string `json:"created_at"`
		User       struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
	}
	var reviews []ReviewItem
	var sumRating int
	for rows.Next() {
		var ri ReviewItem
		if err := rows.Scan(&ri.ID, &ri.Rating, &ri.Comment, &ri.CreatedAt, &ri.User.ID, &ri.User.Name, &ri.User.Email, &ri.User.AvatarURL); err != nil {
			continue
		}
		reviews = append(reviews, ri)
		sumRating += ri.Rating
	}

	avgRating := 0.0
	if total > 0 {
		avgRating = float64(sumRating) / float64(total)
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews":       reviews,
		"total":        total,
		"page":         page,
		"per_page":     perPage,
		"avg_rating":   fmt.Sprintf("%.1f", avgRating),
		"rating_count": total,
	})
}

func RegisterReviewRoutes(group *gin.RouterGroup) {
	h := NewReviewHandler(nil)
	group.POST("/reviews", h.CreateReview)
	group.GET("/reviews", h.ReviewList)
}

// ---------- Recently Viewed ----------

type RecentlyViewedHandler struct{ db *sql.DB }

func NewRecentlyViewedHandler(db *sql.DB) *RecentlyViewedHandler {
	if db == nil {
		db = database.DB
	}
	return &RecentlyViewedHandler{db}
}

// RecordView records a product view
func (h *RecentlyViewedHandler) RecordView(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		ProductID int `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Upsert (replaces old entry with new timestamp)
	h.db.Exec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`,
		uid, req.ProductID)

	// Keep only last 20
	h.db.Exec(`DELETE FROM recently_viewed WHERE user_id=$1 AND id NOT IN
		(SELECT id FROM recently_viewed WHERE user_id=$1 ORDER BY viewed_at DESC LIMIT 20)`, uid)

	c.JSON(http.StatusCreated, gin.H{"recorded": true})
}

// RecentlyViewedList returns last viewed products
func (h *RecentlyViewedHandler) RecentlyViewedList(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 { limit = 1 }
	if limit > 50 { limit = 50 }

	rows, err := h.db.Query(`
		SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity,
		       COALESCE(pi.url, '') as primary_image, rv.viewed_at
		FROM recently_viewed rv
		JOIN products p ON p.id = rv.product_id
		LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true
		WHERE rv.user_id = $1
		ORDER BY rv.viewed_at DESC
		LIMIT $2`, uid, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load history"})
		return
	}
	defer rows.Close()

	type ViewedItem struct {
		ID           int    `json:"id"`
		Title        string `json:"title"`
		Slug         string `json:"slug"`
		PriceUSD     string `json:"price_usd"`
		Stock        int    `json:"stock"`
		PrimaryImage string `json:"primary_image"`
		ViewedAt     string `json:"viewed_at"`
	}
	var items []ViewedItem
	for rows.Next() {
		var vi ViewedItem
		rows.Scan(&vi.ID, &vi.Title, &vi.Slug, &vi.PriceUSD, &vi.Stock, &vi.PrimaryImage, &vi.ViewedAt)
		items = append(items, vi)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"count": len(items),
	})
}

func RegisterRecentlyViewedRoutes(group *gin.RouterGroup) {
	h := NewRecentlyViewedHandler(nil)
	group.POST("/recently-viewed", h.RecordView)
	group.GET("/recently-viewed", h.RecentlyViewedList)
}

// ---------- Product Comparison ----------

type ComparisonHandler struct{ db *sql.DB }

func NewComparisonHandler(db *sql.DB) *ComparisonHandler {
	if db == nil {
		db = database.DB
	}
	return &ComparisonHandler{db}
}

// ToggleComparison adds or removes from comparison list
func (h *ComparisonHandler) ToggleComparison(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		ProductID int `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id=$1 AND product_id=$2)", uid, req.ProductID).Scan(&exists)

	if exists {
		h.db.Exec("DELETE FROM product_comparison WHERE user_id=$1 AND product_id=$2", uid, req.ProductID)
		c.JSON(http.StatusOK, gin.H{"in_comparison": false})
	} else {
		h.db.Exec("INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING", uid, req.ProductID)
		c.JSON(http.StatusOK, gin.H{"in_comparison": true})
	}
}

// ComparisonList returns comparison table
func (h *ComparisonHandler) ComparisonList(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	rows, err := h.db.Query(`
		SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd,
		       p.stock_quantity, p.description,
		       COALESCE(p.asset_path, '') as asset_path, COALESCE(p.asset_hash, '') as asset_hash,
		       COALESCE(p.file_mime_type, '') as file_mime_type,
		       COALESCE(created_at::text, '') as created_at
		FROM product_comparison pc
		JOIN products p ON p.id = pc.product_id
		WHERE pc.user_id = $1
		ORDER BY pc.added_at DESC`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comparison"})
		return
	}
	defer rows.Close()

	type CompProduct struct {
		ID            int      `json:"id"`
		Title         string   `json:"title"`
		Slug          string   `json:"slug"`
		PriceUSD      string   `json:"price_usd"`
		Stock         int      `json:"stock"`
		Description   string   `json:"description"`
		AssetPath     string   `json:"asset_path"`
		AssetHash     string   `json:"asset_hash"`
		FileMimeType  string   `json:"file_mime_type"`
		CreatedAt     string   `json:"created_at"`
		PrimaryImage  string   `json:"primary_image"`
		FeatureList   []string `json:"feature_list"`
	}
	var items []CompProduct
	for rows.Next() {
		var cp CompProduct
		if err := rows.Scan(&cp.ID, &cp.Title, &cp.Slug, &cp.PriceUSD, &cp.Stock, &cp.Description,
			&cp.AssetPath, &cp.AssetHash, &cp.FileMimeType, &cp.CreatedAt); err != nil {
			continue
		}
		cp.FeatureList = []string{}
		if cp.Description != "" {
			// Simple feature extraction: split by common delimiters
		}
		// Get primary image
		var imgURL string
		h.db.QueryRow("SELECT COALESCE(url, '') FROM product_images WHERE product_id=$1 AND is_primary=true LIMIT 1", cp.ID).Scan(&imgURL)
		cp.PrimaryImage = imgURL
		items = append(items, cp)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"count": len(items),
	})
}

func RegisterComparisonRoutes(group *gin.RouterGroup) {
	h := NewComparisonHandler(nil)
	group.POST("/compare/toggle", h.ToggleComparison)
	group.GET("/compare", h.ComparisonList)
}

// ---------- Product Recommendations ----------

type RecommendationHandler struct{ db *sql.DB }

func NewRecommendationHandler(db *sql.DB) *RecommendationHandler {
	if db == nil {
		db = database.DB
	}
	return &RecommendationHandler{db}
}

// Recommend returns products related to a given product
func (h *RecommendationHandler) Recommend(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "8"))
	if page < 1 { page = 1 }
	if perPage < 1 { perPage = 1 }
	if perPage > 20 { perPage = 20 }
	offset := (page - 1) * perPage

	// Get the product's category
	var catID sql.NullInt64
	h.db.QueryRow("SELECT category_id FROM products WHERE id=$1", productID).Scan(&catID)

	var rows *sql.Rows
	var err2 error
	if catID.Valid && catID.Int64 > 0 {
		// Same category, different product
		rows, err2 = h.db.Query(`
			SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd,
			       p.stock_quantity, p.status,
			       COALESCE(pi.url, '') as primary_image
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true
			WHERE p.category_id = $1 AND p.id != $2 AND p.status = 'active'
			ORDER BY p.views_count DESC, p.created_at DESC
			LIMIT $3 OFFSET $4`, catID.Int64, productID, perPage, offset)
	} else {
		// Global popular products
		rows, err2 = h.db.Query(`
			SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd,
			       p.stock_quantity, p.status,
			       COALESCE(pi.url, '') as primary_image
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true
			WHERE p.id != $1 AND p.status = 'active'
			ORDER BY p.views_count DESC, p.created_at DESC
			LIMIT $2 OFFSET $3`, productID, perPage, offset)
	}
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load recommendations"})
		return
	}
	defer rows.Close()

	type RecProduct struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		PriceUSD    string `json:"price_usd"`
		Stock       int    `json:"stock"`
		Status      string `json:"status"`
		PrimaryImage string `json:"primary_image"`
	}
	var recs []RecProduct
	for rows.Next() {
		var rp RecProduct
		if err := rows.Scan(&rp.ID, &rp.Title, &rp.Slug, &rp.PriceUSD, &rp.Stock, &rp.Status, &rp.PrimaryImage); err != nil {
			continue
		}
		recs = append(recs, rp)
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recs,
		"total":          len(recs),
		"page":           page,
	})
}

func RegisterRecommendationRoutes(group *gin.RouterGroup) {
	h := NewRecommendationHandler(nil)
	group.GET("/recommendations/:id", h.Recommend)
}

// ---------- Guest Order (public) ----------

// CreateGuestOrder creates a guest order (no auth required)
func CreateGuestOrder(c *gin.Context) {
	var req struct {
		Email     string `json:"email"`
		ProductIDs string `json:"product_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.ProductIDs == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and product_ids are required"})
		return
	}

	// Calculate total from products
	var totalUSD string
	err := database.DB.QueryRow(
		"SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = 'active'",
		parseIntArray(req.ProductIDs),
	).Scan(&totalUSD)
	if err != nil || totalUSD == "0.00" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid products or total is zero"})
		return
	}

	// Insert guest order
	var orderID int
	err = database.DB.QueryRow(
		`INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order)
		 VALUES ($1, $2, $3, 'pending', true) RETURNING id`,
		req.Email, totalUSD, "BSC",
	).Scan(&orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"order_id": orderID, "status": "pending", "total_usd": totalUSD})
}

// GetGuestOrder returns a guest order by ID (public, email must match)
func GetGuestOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var email, status, totalUSD, cryptoChain string
	err = database.DB.QueryRow(
		"SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true",
		id,
	).Scan(&email, &status, &totalUSD, &cryptoChain)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id, "email": email, "status": status,
		"total_usd": totalUSD, "crypto_chain": cryptoChain,
	})
}

func parseIntArray(s string) []int {
	// Simple comma-separated integer array parser
	var result []int
	for _, part := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil {
			result = append(result, n)
		}
	}
	return result
}
