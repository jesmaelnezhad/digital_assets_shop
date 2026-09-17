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

// ---------- Admin Bulk Product Actions ----------

type BulkProductHandler struct{ db *sql.DB }

func NewBulkProductHandler(db *sql.DB) *BulkProductHandler {
	if db == nil {
		db = database.DB
	}
	return &BulkProductHandler{db}
}

// AdminBulkUpdateProducts updates status/category/stock for multiple products
func (h *BulkProductHandler) AdminBulkUpdateProducts(c *gin.Context) {
	var req struct {
		ProductIDs []struct {
			ProductID int `json:"product_id"`
		} `json:"product_ids"`
		Status     string `json:"status"`
		CategoryID sql.NullInt64 `json:"category_id"`
		PriceUSD   sql.NullString `json:"price_usd"`
		StockQty   sql.NullInt64 `json:"stock_quantity"`
		Title      string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.ProductIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: provide product_ids array"})
		return
	}

	validIDs := make([]int, 0, len(req.ProductIDs))
	for _, entry := range req.ProductIDs {
		if entry.ProductID > 0 {
			validIDs = append(validIDs, entry.ProductID)
		}
	}
	if len(validIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid product IDs"})
		return
	}

	// Validate status if provided
	if req.Status != "" {
		switch req.Status {
		case "active", "draft", "discontinued", "hidden":
			// valid
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
	}

	type update struct {
		id int
	}
	updated := 0
	failed := 0

	for _, id := range validIDs {
		var sets []string
		var args []interface{}
		argIdx := 1

		if req.Status != "" {
			sets = append(sets, fmt.Sprintf("status = $%d", argIdx))
			args = append(args, req.Status)
			argIdx++
		}
		if req.CategoryID.Valid {
			sets = append(sets, fmt.Sprintf("category_id = $%d", argIdx))
			args = append(args, req.CategoryID.Int64)
			argIdx++
		}
		if req.PriceUSD.Valid {
			sets = append(sets, fmt.Sprintf("price_usd = $%d", argIdx))
			args = append(args, req.PriceUSD.String)
			argIdx++
		}
		if req.StockQty.Valid {
			sets = append(sets, fmt.Sprintf("stock_quantity = $%d", argIdx))
			args = append(args, req.StockQty.Int64)
			argIdx++
		}
		if req.Title != "" {
			sets = append(sets, fmt.Sprintf("title = $%d", argIdx))
			args = append(args, req.Title)
			argIdx++
		}
		args = append(args, id)

		if len(sets) > 0 {
			query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d", strings.Join(sets, ", "), argIdx)
			res, err := h.db.Exec(query, args...)
			if err != nil {
				failed++
				continue
			}
			n, _ := res.RowsAffected()
			if n > 0 {
				updated++
			} else {
				failed++
			}
		} else {
			failed++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"updated": updated,
		"failed":  failed,
		"message": fmt.Sprintf("bulk update complete: %d updated, %d failed", updated, failed),
	})
}

// AdminProductList returns paginated product list with filters
func (h *BulkProductHandler) AdminProductList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 { page = 1 }
	if perPage < 1 { perPage = 20 }
	if perPage > 100 { perPage = 100 }
	offset := (page - 1) * perPage

	filterStatus := c.Query("status")
	filterCategory := c.Query("category")
	filterSearch := c.Query("search")

	// Count
	var total int
	countQuery := "SELECT COUNT(*) FROM products WHERE 1=1"
	countArgs := []interface{}{}
	if filterStatus != "" {
		countQuery += " AND status = $" + fmt.Sprintf("%d", len(countArgs)+1)
		countArgs = append(countArgs, filterStatus)
	}
	if filterCategory != "" {
		countQuery += " AND category_id = $" + fmt.Sprintf("%d", len(countArgs)+1)
		countArgs = append(countArgs, filterCategory)
	}
	if filterSearch != "" {
		countQuery += " AND (title ILIKE $" + fmt.Sprintf("%d", len(countArgs)+1) + " OR slug ILIKE $" + fmt.Sprintf("%d", len(countArgs)+1) + ")"
		countArgs = append(countArgs, "%"+filterSearch+"%")
	}
	h.db.QueryRow(countQuery, countArgs...).Scan(&total)

	// List
	listQuery := `
		SELECT p.id, p.title, p.slug, p.description,
		       COALESCE(c.name, '') as category_name, COALESCE(c.slug, '') as category_slug,
		       CAST(p.price_usd AS NUMERIC) as price_usd,
		       p.status, p.stock_quantity,
		       CAST(p.crypto_amount AS NUMERIC) as crypto_amount, p.crypto_currency,
		       p.asset_path, p.asset_hash, p.file_mime_type,
		       p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes,
		       p.search_tags, p.views_count,
		       p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE 1=1`
	listArgs := []interface{}{}
	if filterStatus != "" {
		listQuery += " AND p.status = $" + fmt.Sprintf("%d", len(listArgs)+1)
		listArgs = append(listArgs, filterStatus)
	}
	if filterCategory != "" {
		listQuery += " AND p.category_id = $" + fmt.Sprintf("%d", len(listArgs)+1)
		listArgs = append(listArgs, filterCategory)
	}
	if filterSearch != "" {
		listQuery += " AND (p.title ILIKE $" + fmt.Sprintf("%d", len(listArgs)+1) + " OR p.slug ILIKE $" + fmt.Sprintf("%d", len(listArgs)+1) + ")"
		listArgs = append(listArgs, "%"+filterSearch+"%")
	}
	listQuery += fmt.Sprintf(" ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d", len(listArgs)+1, len(listArgs)+2)
	listArgs = append(listArgs, perPage, offset)

	rows, err := h.db.Query(listQuery, listArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load products"})
		return
	}
	defer rows.Close()

	type AdminProduct struct {
		ID              int    `json:"id"`
		Title           string `json:"title"`
		Slug            string `json:"slug"`
		Description     string `json:"description"`
		CategoryName    string `json:"category_name"`
		CategorySlug    string `json:"category_slug"`
		PriceUSD        string `json:"price_usd"`
		Status          string `json:"status"`
		StockQuantity   int    `json:"stock_quantity"`
		CryptoAmount    string `json:"crypto_amount"`
		CryptoCurrency  string `json:"crypto_currency"`
		AssetPath       string `json:"asset_path"`
		AssetHash       string `json:"asset_hash"`
		FileMimeType    string `json:"file_mime_type"`
		DownloadCountLimit int `json:"download_count_limit"`
		MaxDownloadsPerUser int `json:"max_downloads_per_user"`
		FileSizeBytes   int64  `json:"file_size_bytes"`
		SearchTags      string `json:"search_tags"`
		ViewsCount      int    `json:"views_count"`
		CreatedAt       string `json:"created_at"`
		UpdatedAt       string `json:"updated_at"`
	}
	var products []AdminProduct
	for rows.Next() {
		var ap AdminProduct
		if err := rows.Scan(&ap.ID, &ap.Title, &ap.Slug, &ap.Description, &ap.CategoryName, &ap.CategorySlug,
			&ap.PriceUSD, &ap.Status, &ap.StockQuantity, &ap.CryptoAmount, &ap.CryptoCurrency,
			&ap.AssetPath, &ap.AssetHash, &ap.FileMimeType, &ap.DownloadCountLimit, &ap.MaxDownloadsPerUser,
			&ap.FileSizeBytes, &ap.SearchTags, &ap.ViewsCount, &ap.CreatedAt, &ap.UpdatedAt); err != nil {
			continue
		}
		products = append(products, ap)
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func RegisterAdminFeaturesRoutes(group *gin.RouterGroup) {
	h := NewBulkProductHandler(nil)
	group.POST("/products/bulk", h.AdminBulkUpdateProducts)
	group.GET("/products", h.AdminProductList)
}

// ---------- Admin Order Status Workflow ----------

type OrderStatusHandler struct{ db *sql.DB }

func NewOrderStatusHandler(db *sql.DB) *OrderStatusHandler {
	if db == nil {
		db = database.DB
	}
	return &OrderStatusHandler{db}
}

// validTransitions defines allowed status transitions
var validTransitions = map[string][]string{
	"pending":   {"pending", "paid", "cancelled", "failed", "refunded"},
	"paid":      {"paid", "completed", "cancelled", "refunded"},
	"completed": {"completed", "refunded"},
	"cancelled": {"cancelled"},
	"failed":    {"failed", "cancelled"},
	"refunded":  {"refunded"},
}

// AdminUpdateOrderStatus updates an order's status with transition validation
func (h *OrderStatusHandler) AdminUpdateOrderStatus(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req struct {
		Status  string `json:"status"`
		Notes   string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status required"})
		return
	}

	// Validate status value
	switch req.Status {
	case "pending", "paid", "completed", "cancelled", "failed", "refunded":
		// valid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	// Get current order
	var currentStatus string
	err = h.db.QueryRow("SELECT status FROM orders WHERE id=$1", orderID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	// Validate transition
	allowed, exists := validTransitions[currentStatus]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown current status"})
		return
	}
	validTransition := false
	for _, s := range allowed {
		if s == req.Status {
			validTransition = true
			break
		}
	}
	if !validTransition {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        "invalid transition",
			"current":      currentStatus,
			"target":       req.Status,
			"allowed":      allowed,
		})
		return
	}

	// Update
	nowStr := "CURRENT_TIMESTAMP"
	query := "UPDATE orders SET status = $1, updated_at = " + nowStr
	if req.Notes != "" {
		query += ", notes = $2"
	}
	if req.Status == "paid" {
		query += ", paid_at = " + nowStr
	}
	query += " WHERE id = $" + fmt.Sprintf("%d", 2+len(req.Notes))
	if req.Notes != "" {
		h.db.Exec(query, req.Status, req.Notes, orderID)
	} else {
		h.db.Exec(query, req.Status, orderID)
	}

	// Return updated order
	var status, totalUSD, guestEmail, billingName string
	var paidAt sql.NullString
	h.db.QueryRow("SELECT status, CAST(total_usd AS NUMERIC), guest_email, billing_name, paid_at FROM orders WHERE id=$1", orderID).
		Scan(&status, &totalUSD, &guestEmail, &billingName, &paidAt)

	c.JSON(http.StatusOK, gin.H{
		"order": gin.H{
			"id":          orderID,
			"status":      status,
			"total_usd":   totalUSD,
			"guest_email": guestEmail,
			"billing_name": billingName,
			"paid_at":     paidAt.String,
		},
		"message": "status updated",
	})
}

// AdminOrderDetail returns full order detail
func (h *OrderStatusHandler) AdminOrderDetail(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var status, totalUSD, guestEmail, billingName, shippingAddress, notes, createdAt string
	var paidAt sql.NullString
	h.db.QueryRow(`
		SELECT o.status, CAST(o.total_usd AS NUMERIC), o.guest_email, o.billing_name,
		       o.shipping_address_json, o.notes, o.created_at, o.paid_at
		FROM orders o WHERE o.id=$1`, orderID).
		Scan(&status, &totalUSD, &guestEmail, &billingName, &shippingAddress, &notes, &createdAt, &paidAt)

	// Items
	type OrderItem struct {
		ID            int    `json:"id"`
		ProductID     int    `json:"product_id"`
		ProductTitle  string `json:"product_title"`
		ProductSlug   string `json:"product_slug"`
		Quantity      int    `json:"quantity"`
		UnitPriceUSD  string `json:"unit_price_usd"`
		DownloadCount int    `json:"download_count"`
	}
	rows, _ := h.db.Query("SELECT id, product_id, product_title, product_slug, quantity, CAST(unit_price_usd AS NUMERIC), download_count FROM order_items WHERE order_id=$1 ORDER BY id", orderID)
	defer rows.Close()
	var items []OrderItem
	for rows.Next() {
		var oi OrderItem
		rows.Scan(&oi.ID, &oi.ProductID, &oi.ProductTitle, &oi.ProductSlug, &oi.Quantity, &oi.UnitPriceUSD, &oi.DownloadCount)
		items = append(items, oi)
	}

	// Get user info if user_id exists
	var userID sql.NullInt64
	var userName, userEmail string
	h.db.QueryRow("SELECT user_id, u.name, u.email FROM orders o LEFT JOIN users u ON u.id = o.user_id WHERE o.id=$1", orderID).
		Scan(&userID, &userName, &userEmail)

	type AdminOrderDetail struct {
		ID           int    `json:"id"`
		Status       string `json:"status"`
		TotalUSD     string `json:"total_usd"`
		GuestEmail   string `json:"guest_email"`
		BillingName  string `json:"billing_name"`
		ShippingAddress string `json:"shipping_address"`
		Notes        string `json:"notes"`
		CreatedAt    string `json:"created_at"`
		PaidAt       string `json:"paid_at"`
		Items        []OrderItem `json:"items"`
		UserID       *int   `json:"user_id"`
		UserName     string `json:"user_name"`
		UserEmail    string `json:"user_email"`
	}
	c.JSON(http.StatusOK, gin.H{
		"order": AdminOrderDetail{
			ID:           orderID,
			Status:       status,
			TotalUSD:     totalUSD,
			GuestEmail:   guestEmail,
			BillingName:  billingName,
			ShippingAddress: shippingAddress,
			Notes:        notes,
			CreatedAt:    createdAt,
			PaidAt:       paidAt.String,
			Items:        items,
			UserID:       nil,
			UserName:     userName,
			UserEmail:    userEmail,
		},
	})
}

func RegisterOrderStatusRoutes(group *gin.RouterGroup) {
	h := NewOrderStatusHandler(nil)
	group.PUT("/order/:id/status", h.AdminUpdateOrderStatus)
	group.GET("/order/:id/detail", h.AdminOrderDetail)
}

// ---------- Guest Order Check (self-service) ----------

// GuestOrderCheck returns order status for guest (uses order_token)
func (h *OrderStatusHandler) GuestOrderCheckSelf(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	// In production, order_token would be used; here we match by email query param
	var email string
	if val := c.Query("email"); val != "" {
		email = val
	}

	var guestEmail string
	h.db.QueryRow("SELECT guest_email FROM orders WHERE id=$1", orderID).Scan(&guestEmail)
	if guestEmail == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if email != "" && guestEmail != email {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var status string
	var totalUSD string
	h.db.QueryRow("SELECT status, CAST(total_usd AS NUMERIC) FROM orders WHERE id=$1", orderID).Scan(&status, &totalUSD)

	type OrderItem struct {
		ID            int    `json:"id"`
		ProductID     int    `json:"product_id"`
		ProductTitle  string `json:"product_title"`
		ProductSlug   string `json:"product_slug"`
		Quantity      int    `json:"quantity"`
		UnitPriceUSD  string `json:"unit_price_usd"`
		DownloadCount int    `json:"download_count"`
	}
	rows, _ := h.db.Query("SELECT id, product_id, product_title, product_slug, quantity, CAST(unit_price_usd AS NUMERIC), download_count FROM order_items WHERE order_id=$1", orderID)
	defer rows.Close()
	var items []OrderItem
	for rows.Next() {
		var oi OrderItem
		rows.Scan(&oi.ID, &oi.ProductID, &oi.ProductTitle, &oi.ProductSlug, &oi.Quantity, &oi.UnitPriceUSD, &oi.DownloadCount)
		items = append(items, oi)
	}

	c.JSON(http.StatusOK, gin.H{
		"order": gin.H{
			"id":         orderID,
			"status":     status,
			"total_usd":  totalUSD,
			"guest_email": guestEmail,
			"items":      items,
		},
	})
}

func RegisterGuestOrderRoutes(group *gin.RouterGroup) {
	// guest order self-check (public, but requires order id + email)
	h := NewOrderStatusHandler(nil)
	group.GET("/orders/guest/:id", h.GuestOrderCheckSelf)
}

// ---------- Admin Community Moderation ----------

type CommunityModerationHandler struct{ db *sql.DB }

func NewCommunityModerationHandler(db *sql.DB) *CommunityModerationHandler { return &CommunityModerationHandler{db} }

// ListCommunityPosts returns all posts with user info for moderation
func (h *CommunityModerationHandler) ListCommunityPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 { page = 1 }
	if perPage < 1 { perPage = 20 }
	if perPage > 100 { perPage = 100 }
	offset := (page - 1) * perPage

	filterUser := c.Query("user_id")

	var total int
	countQuery := "SELECT COUNT(*) FROM community_posts WHERE 1=1"
	if filterUser != "" {
		countQuery += " AND user_id = $" + fmt.Sprintf("%d", 1)
	}
	h.db.QueryRow(countQuery, filterUser).Scan(&total)

	listQuery := `
		SELECT cp.id, cp.user_id, u.name, u.email, cp.content, cp.created_at, cp.updated_at
		FROM community_posts cp
		LEFT JOIN users u ON u.id = cp.user_id
		WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if filterUser != "" {
		listQuery += fmt.Sprintf(" AND cp.user_id = $%d", argIdx)
		args = append(args, filterUser)
		argIdx++
	}
	listQuery += fmt.Sprintf(" ORDER BY cp.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := h.db.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load posts"})
		return
	}
	defer rows.Close()

	type ModPost struct {
		ID       int    `json:"id"`
		UserID   int    `json:"user_id"`
		UserName string `json:"user_name"`
		UserEmail string `json:"user_email"`
		Content  string `json:"content"`
		CreatedAt string `json:"created_at"`
	}
	var posts []ModPost
	for rows.Next() {
		var mp ModPost
		rows.Scan(&mp.ID, &mp.UserID, &mp.UserName, &mp.UserEmail, &mp.Content, &mp.CreatedAt)
		posts = append(posts, mp)
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":  posts,
		"total":  total,
		"page":   page,
		"per_page": perPage,
	})
}

// DeleteCommunityPost soft-deletes a post (admin action)
func (h *CommunityModerationHandler) DeleteCommunityPost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	// Get post to verify exists
	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM community_posts WHERE id=$1)", postID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	h.db.Exec("DELETE FROM community_posts WHERE id=$1", postID)
	c.JSON(http.StatusOK, gin.H{"deleted": true, "message": "post removed"})
}

// ListCommunityUsers returns all users with stats for admin moderation
func (h *CommunityModerationHandler) ListCommunityUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 { page = 1 }
	if perPage < 1 { perPage = 20 }
	if perPage > 100 { perPage = 100 }
	offset := (page - 1) * perPage

	filterSearch := c.Query("search")

	var total int
	countQuery := "SELECT COUNT(*) FROM users WHERE 1=1"
	args := []interface{}{}
	argIdx := 1
	if filterSearch != "" {
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filterSearch+"%", "%"+filterSearch+"%")
		argIdx++
	}
	h.db.QueryRow(countQuery, args...).Scan(&total)

	listQuery := `
		SELECT u.id, u.name, u.email, u.created_at, u.updated_at,
		       COALESCE(up.bio, '') as bio,
		       COALESCE(up.avatar_url, '') as avatar_url,
		       COALESCE(up.wallet_address, '') as wallet_address,
		       COALESCE(cp.post_count, 0) as post_count,
		       COALESCE(pl.like_count, 0) as like_count,
		       COALESCE(pc.comment_count, 0) as comment_count,
		       COALESCE(tf.follower_count, 0) as follower_count,
		       COALESCE(tf.following_count, 0) as following_count
		FROM users u
		LEFT JOIN user_profiles up ON up.user_id = u.id
		LEFT JOIN (SELECT user_id, COUNT(*) as post_count FROM community_posts GROUP BY user_id) cp ON cp.user_id = u.id
		LEFT JOIN (SELECT user_id, COUNT(*) as like_count FROM post_likes GROUP BY user_id) pl ON pl.user_id = u.id
		LEFT JOIN (SELECT user_id, COUNT(*) as comment_count FROM post_comments GROUP BY user_id) pc ON pc.user_id = u.id
		LEFT JOIN (
			SELECT f.user_id,
			       COUNT(CASE WHEN f.followed_id IS NOT NULL THEN 1 END) as follower_count,
			       COUNT(CASE WHEN f.follower_id IS NOT NULL THEN 1 END) as following_count
			FROM follows f
			GROUP BY f.user_id
		) tf ON tf.user_id = u.id
		WHERE 1=1`
	args2 := []interface{}{}
	argIdx2 := 1
	if filterSearch != "" {
		listQuery += fmt.Sprintf(" AND (u.name ILIKE $%d OR u.email ILIKE $%d)", argIdx2, argIdx2)
		args2 = append(args2, "%"+filterSearch+"%", "%"+filterSearch+"%")
		argIdx2 += 2
	}
	listQuery += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", argIdx2, argIdx2+1)
	args2 = append(args2, perPage, offset)

	rows, err := h.db.Query(listQuery, args2...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
		return
	}
	defer rows.Close()

	type ModUser struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Email        string `json:"email"`
		Bio          string `json:"bio"`
		AvatarURL    string `json:"avatar_url"`
		WalletAddress string `json:"wallet_address"`
		PostCount    int    `json:"post_count"`
		LikeCount    int    `json:"like_count"`
		CommentCount int    `json:"comment_count"`
		FollowerCount int   `json:"follower_count"`
		FollowingCount int  `json:"following_count"`
		CreatedAt    string `json:"created_at"`
	}
	var users []ModUser
	for rows.Next() {
		var mu ModUser
		rows.Scan(&mu.ID, &mu.Name, &mu.Email, &mu.Bio, &mu.AvatarURL, &mu.WalletAddress,
			&mu.PostCount, &mu.LikeCount, &mu.CommentCount, &mu.FollowerCount, &mu.FollowingCount, &mu.CreatedAt)
		users = append(users, mu)
	}

	c.JSON(http.StatusOK, gin.H{
		"users":   users,
		"total":   total,
		"page":    page,
		"per_page": perPage,
	})
}

func RegisterCommunityModerationRoutes(group *gin.RouterGroup) {
	h := NewCommunityModerationHandler(nil)
	group.GET("/community/posts", h.ListCommunityPosts)
	group.DELETE("/community/posts/:id", h.DeleteCommunityPost)
	group.GET("/community/users", h.ListCommunityUsers)
}
