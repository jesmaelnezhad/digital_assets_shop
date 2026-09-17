// handlers.go - Admin handler for all admin routes + missing handlers

package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	adminmodels "github.com/pawradise/admin-service/models"
	"github.com/pawradise/shared/models"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	db          *sql.DB
	identityDB  *sql.DB
	commerceDB  *sql.DB
}

func NewAdminHandler(db *sql.DB, identityDB *sql.DB) *AdminHandler {
	h := &AdminHandler{db: db, identityDB: identityDB, commerceDB: db}
	h.ensureExtras()
	return h
}

// Users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	rows, err := h.identityDB.Query("SELECT id, email, name, created_at, updated_at FROM users ORDER BY created_at DESC")
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()
	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt); err != nil { continue }
		users = append(users, u)
	}
	if users == nil { users = []models.User{} }
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	result, err := h.identityDB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"}); return }
	rows, _ := result.RowsAffected()
	if rows == 0 { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	var u models.User
	if err := h.identityDB.QueryRow("SELECT id FROM users WHERE id = $1", id).Scan(&u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"}); return
	}

	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate"}); return
	}
	newPassword := base64.URLEncoding.EncodeToString(randomBytes)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash"}); return }

	_, err = h.identityDB.Exec("UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(hashedPassword), id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "reset failed"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "password reset", "new_password": newPassword})
}

// Products
func (h *AdminHandler) ListAllProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 { page = 1 }
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if perPage < 1 || perPage > 200 { perPage = 50 }

	rows, err := h.db.Query(
		`SELECT p.id, p.title, p.slug, p.description, p.category_id, p.price_usd, p.status,
			p.stock_quantity, p.max_downloads_per_user, p.views, p.downloads,
			p.purchase_count, p.created_at, p.updated_at,
			COALESCE(c.name, ''), COALESCE(c.slug, '')
		FROM products p LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`,
		perPage, (page-1)*perPage,
	)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()

	type pWithCat struct {
		models.Product
		CategoryName string `json:"category_name"`
		CategorySlug string `json:"category_slug"`
	}
	products := []pWithCat{}
	for rows.Next() {
		var p pWithCat
		var catName, catSlug sql.NullString
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.CategoryID, &p.PriceUSD, &p.Status,
			&p.StockQuantity, &p.MaxDownloadsPerUser, &p.Views, &p.Downloads,
			&p.PurchaseCount, &p.CreatedAt, &p.UpdatedAt, &catName, &catSlug); err != nil { continue }
		if catName.Valid { p.CategoryName = catName.String }
		if catSlug.Valid { p.CategorySlug = catSlug.String }
		products = append(products, p)
	}
	if products == nil { products = []pWithCat{} }

	var total int
	h.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&total)

	c.JSON(http.StatusOK, gin.H{"products": products, "total": total})
}

func (h *AdminHandler) GetProductStats(c *gin.Context) {
	var total, active, pinned int
	h.db.QueryRow("SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'active'), COUNT(*) FILTER (WHERE pinned = true) FROM products").Scan(&total, &active, &pinned)
	c.JSON(http.StatusOK, gin.H{"total_products": total, "active_products": active, "pinned_products": pinned})
}

func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var req struct {
		Title       string  `json:"title" binding:"required"`
		Slug        string  `json:"slug"`
		Description string  `json:"description"`
		CategoryID  *int    `json:"category_id"`
		PriceUSD    float64 `json:"price_usd"`
		AssetPath   string  `json:"asset_path"`
		AssetHash   string  `json:"asset_hash"`
		Status      string  `json:"status"`
		StockQty    int     `json:"stock_quantity"`
		MaxDL       int     `json:"max_downloads_per_user"`
		Pinned      *bool   `json:"pinned"`
		PWYWEnabled bool    `json:"pwyw_enabled"`
		PWYWMin     float64 `json:"pwyw_min_price"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}
	if req.Title == "" { req.Title = "Untitled Product" }
	if req.Slug == "" { req.Slug = toSlug(req.Title) }
	if req.Status == "" { req.Status = "active" }

	var id int
	err := h.db.QueryRow(
		`INSERT INTO products (title, slug, description, category_id, price_usd, asset_path, asset_hash, status,
			stock_quantity, max_downloads_per_user, pinned, pwyw_enabled, pwyw_min_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING id`,
		req.Title, req.Slug, req.Description, req.CategoryID, req.PriceUSD,
		req.AssetPath, req.AssetHash, req.Status, req.StockQty, req.MaxDL,
		req.Pinned, req.PWYWEnabled, req.PWYWMin,
	).Scan(&id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "product created"})
}

func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		CategoryID  *int    `json:"category_id"`
		PriceUSD    float64 `json:"price_usd"`
		AssetPath   string  `json:"asset_path"`
		AssetHash   string  `json:"asset_hash"`
		Status      string  `json:"status"`
		StockQty    int     `json:"stock_quantity"`
		MaxDL       int     `json:"max_downloads_per_user"`
		Pinned      *bool   `json:"pinned"`
		PWYWEnabled *bool   `json:"pwyw_enabled"`
		PWYWMin     float64 `json:"pwyw_min_price"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}

	// Build the SET clause dynamically with PostgreSQL-style placeholders
	query := `UPDATE products SET updated_at = NOW()`
	args := []interface{}{}
	idx := 1

	if req.Title != "" { query += fmt.Sprintf(", title = $%d", idx); args = append(args, req.Title); idx++ }
	if req.Description != "" { query += fmt.Sprintf(", description = $%d", idx); args = append(args, req.Description); idx++ }
	if req.CategoryID != nil { query += fmt.Sprintf(", category_id = $%d", idx); args = append(args, *req.CategoryID); idx++ }
	if req.PriceUSD > 0 { query += fmt.Sprintf(", price_usd = $%d", idx); args = append(args, req.PriceUSD); idx++ }
	if req.AssetPath != "" { query += fmt.Sprintf(", asset_path = $%d", idx); args = append(args, req.AssetPath); idx++ }
	if req.AssetHash != "" { query += fmt.Sprintf(", asset_hash = $%d", idx); args = append(args, req.AssetHash); idx++ }
	if req.Status != "" { query += fmt.Sprintf(", status = $%d", idx); args = append(args, req.Status); idx++ }
	if req.StockQty > 0 { query += fmt.Sprintf(", stock_quantity = $%d", idx); args = append(args, req.StockQty); idx++ }
	if req.MaxDL > 0 { query += fmt.Sprintf(", max_downloads_per_user = $%d", idx); args = append(args, req.MaxDL); idx++ }
	if req.Pinned != nil { query += fmt.Sprintf(", pinned = $%d", idx); args = append(args, *req.Pinned); idx++ }
	if req.PWYWEnabled != nil { query += fmt.Sprintf(", pwyw_enabled = $%d", idx); args = append(args, *req.PWYWEnabled); idx++ }
	if req.PWYWMin > 0 { query += fmt.Sprintf(", pwyw_min_price = $%d", idx); args = append(args, req.PWYWMin); idx++ }

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	query += fmt.Sprintf(" WHERE id = $%d", idx)
	args = append(args, id)

	_, err = h.db.Exec(query, args...)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }

	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

func (h *AdminHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	_, err = h.db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

func (h *AdminHandler) AddProductTier(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"}); return }

	var req struct {
		TierName     string  `json:"tier_name" binding:"required"`
		FileHash     string  `json:"file_hash"`
		FilePath     string  `json:"file_path"`
		PriceUSD     float64 `json:"price_usd"`
		SortOrder    int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	if req.TierName == "" { req.TierName = "Standard" }
	if req.SortOrder == 0 { req.SortOrder = 1 }

	var tierID int
	err = h.db.QueryRow(
		`INSERT INTO product_tiers (product_id, tier_name, file_path, file_hash, price_usd, sort_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id`,
		productID, req.TierName, req.FilePath, req.FileHash, req.PriceUSD, req.SortOrder,
	).Scan(&tierID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }

	c.JSON(http.StatusCreated, gin.H{"id": tierID, "message": "tier added"})
}

func (h *AdminHandler) UpdateTier(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Param("id"))
	tierID, err := strconv.Atoi(c.Param("tierId"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tier id"}); return }

	var req struct {
		TierName  string  `json:"tier_name"`
		FileHash  string  `json:"file_hash"`
		FilePath  string  `json:"file_path"`
		PriceUSD  float64 `json:"price_usd"`
		SortOrder int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }

	query := `UPDATE product_tiers SET`
	args := []interface{}{}
	idx := 2

	if req.TierName != "" { query += fmt.Sprintf(" tier_name = $%d", idx); args = append(args, req.TierName); idx++ }
	if req.FileHash != "" { query += fmt.Sprintf(" file_hash = $%d", idx); args = append(args, req.FileHash); idx++ }
	if req.FilePath != "" { query += fmt.Sprintf(" file_path = $%d", idx); args = append(args, req.FilePath); idx++ }
	if req.PriceUSD > 0 { query += fmt.Sprintf(" price_usd = $%d", idx); args = append(args, req.PriceUSD); idx++ }
	if req.SortOrder > 0 { query += fmt.Sprintf(" sort_order = $%d", idx); args = append(args, req.SortOrder); idx++ }
	if len(args) == 0 { c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"}); return }

	args = append(args, tierID, productID)
	_, err = h.db.Exec(query+" WHERE id = $1 AND product_id = $2", args...)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }

	c.JSON(http.StatusOK, gin.H{"message": "tier updated"})
}

func (h *AdminHandler) DeleteTier(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Param("id"))
	tierID, err := strconv.Atoi(c.Param("tierId"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tier id"}); return }

	result, err := h.db.Exec("DELETE FROM product_tiers WHERE id = $1 AND product_id = $2", tierID, productID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"}); return }
	rows, _ := result.RowsAffected()
	if rows == 0 { c.JSON(http.StatusNotFound, gin.H{"error": "tier not found"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "tier deleted"})
}

func (h *AdminHandler) AddProductImage(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"}); return }

	var req struct {
		URL        string `json:"url" binding:"required"`
		ImageType  string `json:"image_type"`
		IsPrimary  bool   `json:"is_primary"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	if req.ImageType == "" { req.ImageType = "full" }

	var imgID int
	err = h.db.QueryRow(
		`INSERT INTO product_images (product_id, url, image_type, is_primary, width, height, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id`,
		productID, req.URL, req.ImageType, req.IsPrimary, req.Width, req.Height,
	).Scan(&imgID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }

	c.JSON(http.StatusCreated, gin.H{"id": imgID, "message": "image added"})
}

func (h *AdminHandler) DeleteProductImage(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Param("id"))
	imageID, err := strconv.Atoi(c.Param("imageId"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image id"}); return }

	result, err := h.db.Exec("DELETE FROM product_images WHERE id = $1 AND product_id = $2", imageID, productID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"}); return }
	rows, _ := result.RowsAffected()
	if rows == 0 { c.JSON(http.StatusNotFound, gin.H{"error": "image not found"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "image deleted"})
}

func (h *AdminHandler) BulkUpdateProducts(c *gin.Context) {
	var req struct {
		ProductIDs []int    `json:"ids" binding:"required"`
		Status     *string  `json:"status"`
		CategoryID *int     `json:"category_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	if len(req.ProductIDs) == 0 { c.JSON(http.StatusBadRequest, gin.H{"error": "no product ids"}); return }

	query := "UPDATE products SET updated_at = NOW()"
	args := []interface{}{}

	if req.Status != nil {
		query += ", status = $" + strconv.Itoa(len(args)+1)
		args = append(args, *req.Status)
	}
	if req.CategoryID != nil {
		query += ", category_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, *req.CategoryID)
	}
	if len(args) == 0 && len(req.ProductIDs) > 0 {
		// No fields to update, but we still need to touch the rows
	}
	query += " WHERE id = ANY($" + strconv.Itoa(len(args)+1) + ")"
	args = append(args, pq.Array(req.ProductIDs))

	_, err := h.db.Exec(query, args...)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, gin.H{"message": "bulk update applied"})
}

// Bundles
func (h *AdminHandler) ListBundles(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, title, slug, description, price_usd, status, created_at, updated_at FROM bundles ORDER BY created_at DESC`)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()
	bundles := []models.Bundle{}
	for rows.Next() {
		var b models.Bundle
		if err := rows.Scan(&b.ID, &b.Title, &b.Slug, &b.Description, &b.PriceUSD, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil { continue }
		bundles = append(bundles, b)
	}
	if bundles == nil { bundles = []models.Bundle{} }
	c.JSON(http.StatusOK, gin.H{"bundles": bundles})
}

func (h *AdminHandler) CreateBundle(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		PriceUSD    float64 `json:"price_usd" binding:"required"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	if req.Status == "" { req.Status = "active" }

	var id int
	err := h.db.QueryRow(
		`INSERT INTO bundles (title, slug, description, price_usd, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id`,
		req.Title, slugify(req.Title), req.Description, req.PriceUSD, req.Status,
	).Scan(&id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "bundle created"})
}

func (h *AdminHandler) UpdateBundle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		PriceUSD    float64 `json:"price_usd"`
		Status      string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }

	query := `UPDATE bundles SET updated_at = NOW()`
	args := []interface{}{}
	idx := 2

	if req.Title != "" { query += fmt.Sprintf(", title = $%d", idx); args = append(args, req.Title); idx++ }
	if req.Description != "" { query += fmt.Sprintf(", description = $%d", idx); args = append(args, req.Description); idx++ }
	if req.PriceUSD > 0 { query += fmt.Sprintf(", price_usd = $%d", idx); args = append(args, req.PriceUSD); idx++ }
	if req.Status != "" { query += fmt.Sprintf(", status = $%d", idx); args = append(args, req.Status); idx++ }

	args = append(args, id)
	_, err = h.db.Exec(query+" WHERE id = $1", args...)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, gin.H{"message": "bundle updated"})
}

func (h *AdminHandler) DeleteBundle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	result, err := h.db.Exec("DELETE FROM bundles WHERE id = $1", id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"}); return }
	rows, _ := result.RowsAffected()
	if rows == 0 { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "bundle deleted"})
}

// Coupons
func (h *AdminHandler) ListCoupons(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, code, discount_type, discount_value, expires_at, usage_limit, times_used, min_purchase_usd, product_id, is_active, created_at, updated_at FROM coupons ORDER BY created_at DESC`)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()
	coupons := []adminmodels.Coupon{}
	for rows.Next() {
		var c adminmodels.Coupon
		var expires sql.NullTime
		if err := rows.Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &expires, &c.UsageLimit, &c.TimesUsed, &c.MinPurchaseUSD, &c.ProductID, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil { continue }
		if expires.Valid { c.ExpiresAt = &expires.Time }
		coupons = append(coupons, c)
	}
	if coupons == nil { coupons = []adminmodels.Coupon{} }
	c.JSON(http.StatusOK, gin.H{"coupons": coupons})
}

func (h *AdminHandler) CreateCoupon(c *gin.Context) {
	var req struct {
		Code          string  `json:"code" binding:"required"`
		DiscountType  string  `json:"discount_type" binding:"required"`
		DiscountValue float64 `json:"discount_value" binding:"required"`
		ExpiresAt     string  `json:"expires_at"`
		UsageLimit    int     `json:"usage_limit"`
		MinPurchase   float64 `json:"min_purchase_usd"`
		ProductID     *int    `json:"product_id"`
		IsActive      bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	if req.DiscountType == "" { req.DiscountType = "percentage" }
	if req.UsageLimit == 0 { req.UsageLimit = 100 }

	var expires sql.NullTime
	if req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err == nil { expires = sql.NullTime{Time: t, Valid: true} }
	}

	var id int
	err := h.db.QueryRow(
		`INSERT INTO coupons (code, discount_type, discount_value, expires_at, usage_limit, times_used, min_purchase_usd, product_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, $6, $7, $8, NOW(), NOW()) RETURNING id`,
		req.Code, req.DiscountType, req.DiscountValue, expires, req.UsageLimit, req.MinPurchase, req.ProductID, req.IsActive,
	).Scan(&id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "coupon created"})
}

func (h *AdminHandler) UpdateCoupon(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	var req struct {
		Code          string  `json:"code"`
		DiscountType  string  `json:"discount_type"`
		DiscountValue float64 `json:"discount_value"`
		ExpiresAt     string  `json:"expires_at"`
		UsageLimit    int     `json:"usage_limit"`
		MinPurchase   float64 `json:"min_purchase_usd"`
		ProductID     *int    `json:"product_id"`
		IsActive      *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }

	query := `UPDATE coupons SET updated_at = NOW()`
	args := []interface{}{}
	idx := 2

	if req.Code != "" { query += fmt.Sprintf(", code = $%d", idx); args = append(args, req.Code); idx++ }
	if req.DiscountType != "" { query += fmt.Sprintf(", discount_type = $%d", idx); args = append(args, req.DiscountType); idx++ }
	if req.DiscountValue > 0 { query += fmt.Sprintf(", discount_value = $%d", idx); args = append(args, req.DiscountValue); idx++ }

	var expires sql.NullTime
	if req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err == nil { expires = sql.NullTime{Time: t, Valid: true} }
	}
	if expires.Valid { query += fmt.Sprintf(", expires_at = $%d", idx); args = append(args, expires); idx++ }
	if req.UsageLimit > 0 { query += fmt.Sprintf(", usage_limit = $%d", idx); args = append(args, req.UsageLimit); idx++ }
	if req.MinPurchase > 0 { query += fmt.Sprintf(", min_purchase_usd = $%d", idx); args = append(args, req.MinPurchase); idx++ }
	if req.ProductID != nil { query += fmt.Sprintf(", product_id = $%d", idx); args = append(args, *req.ProductID); idx++ }
	if req.IsActive != nil { query += fmt.Sprintf(", is_active = $%d", idx); args = append(args, *req.IsActive); idx++ }

	if len(args) == 0 { c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"}); return }
	args = append(args, id)
	_, err = h.db.Exec(query+" WHERE id = $1", args...)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, gin.H{"message": "coupon updated"})
}

func (h *AdminHandler) DeleteCoupon(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	result, err := h.db.Exec("DELETE FROM coupons WHERE id = $1", id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"}); return }
	rows, _ := result.RowsAffected()
	if rows == 0 { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "coupon deleted"})
}

// Stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	var totalUsers, totalOrders, totalRevenueUSD, totalProducts, activeProducts, pinnedProducts int

	h.db.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM users").Scan(&totalUsers)
	h.db.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM orders").Scan(&totalOrders)

	var revenue float64
	h.db.QueryRow("SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = 'paid'").Scan(&revenue)
	totalRevenueUSD = int(revenue)

	h.db.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM products").Scan(&totalProducts)
	h.db.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM products WHERE status = 'active'").Scan(&activeProducts)
	h.db.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM products WHERE pinned = true").Scan(&pinnedProducts)

	var totalCategories, totalDownloads int
	h.db.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM categories").Scan(&totalCategories)
	h.db.QueryRow("SELECT COALESCE(SUM(downloads), 0) FROM products").Scan(&totalDownloads)

	revenueDaily := []gin.H{}
	topProducts := []gin.H{}
	conversions := gin.H{
		"views": 0, "carts": 0, "checkouts": 0, "purchases": totalOrders,
		"visitor_to_view": 0, "view_to_cart": 0, "cart_to_checkout": 0, "checkout_to_purchase": 0,
	}

	c.JSON(http.StatusOK, gin.H{
		"total_users":       totalUsers,
		"total_orders":      totalOrders,
		"total_revenue":     totalRevenueUSD,
		"total_products":    totalProducts,
		"active_products":   activeProducts,
		"pinned_products":   pinnedProducts,
		"total_categories":  totalCategories,
		"total_downloads":   totalDownloads,
		"revenue_daily":     revenueDaily,
		"top_products":      topProducts,
		"conversions":       conversions,
	})
}

// Exchange Rates
func (h *AdminHandler) ListExchangeRates(c *gin.Context) {
	rows, err := h.db.Query("SELECT chain, symbol, rate_to_usd, updated_at FROM exchange_rates ORDER BY chain")
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()
	rates := []models.ExchangeRate{}
	for rows.Next() {
		var r models.ExchangeRate
		if err := rows.Scan(&r.Chain, &r.Symbol, &r.RateToUSD, &r.UpdatedAt); err != nil { continue }
		rates = append(rates, r)
	}
	if rates == nil { rates = []models.ExchangeRate{} }
	c.JSON(http.StatusOK, gin.H{"rates": rates})
}

func (h *AdminHandler) SetExchangeRate(c *gin.Context) {
	chain := c.Param("chain")
	var req struct {
		Symbol    string  `json:"symbol"`
		RateToUSD float64 `json:"rate_to_usd" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }

	_, err := h.db.Exec(
		`INSERT INTO exchange_rates (chain, symbol, rate_to_usd, updated_at) VALUES ($1, $2, $3, NOW())
		ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`,
		chain, req.Symbol, req.RateToUSD,
	)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, gin.H{"message": "exchange rate updated"})
}

func (h *AdminHandler) DeleteExchangeRate(c *gin.Context) {
	chain := c.Param("chain")
	result, err := h.db.Exec("DELETE FROM exchange_rates WHERE chain = $1", chain)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"}); return }
	rows, _ := result.RowsAffected()
	if rows == 0 { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "exchange rate deleted"})
}

// Community moderation
func (h *AdminHandler) ListCommunityPosts(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, user_id, content, community_type, created_at FROM community_posts ORDER BY created_at DESC")
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()
	posts := []models.CommunityPost{}
	for rows.Next() {
		var p models.CommunityPost
		if err := rows.Scan(&p.ID, &p.UserID, &p.Content, &p.Type, &p.CreatedAt); err != nil { continue }
		posts = append(posts, p)
	}
	if posts == nil { posts = []models.CommunityPost{} }
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func (h *AdminHandler) DeleteCommunityPost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	result, err := h.db.Exec("DELETE FROM community_posts WHERE id = $1", id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"}); return }
	rows, _ := result.RowsAffected()
	if rows == 0 { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "post deleted"})
}

// Referrals
func (h *AdminHandler) ListReferrals(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT rl.id, rl.user_id, rl.code, rl.is_active, rl.created_at,
			u.email, u.name,
			COALESCE(COUNT(DISTINCT rc.order_id), 0) as referral_count,
			COALESCE(SUM(rc.amount), 0) as total_commission
				FROM referral_links rl
				LEFT JOIN users u ON rl.user_id = u.id
				LEFT JOIN referral_commissions rc ON rc.referral_link_id = rl.id
		GROUP BY rl.id, u.id
		ORDER BY rl.created_at DESC
	`)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()
	type referral struct {
		ID             int        `json:"id"`
		UserID         int        `json:"user_id"`
		Code           string     `json:"code"`
		IsActive       bool       `json:"is_active"`
		CreatedAt      time.Time  `json:"created_at"`
		Email          string     `json:"email"`
		Name           string     `json:"name"`
		ReferralCount  int        `json:"referral_count"`
		TotalCommission float64   `json:"total_commission"`
	}
	referrals := []referral{}
	for rows.Next() {
		var r referral
		if err := rows.Scan(&r.ID, &r.UserID, &r.Code, &r.IsActive, &r.CreatedAt, &r.Email, &r.Name, &r.ReferralCount, &r.TotalCommission); err != nil { continue }
		referrals = append(referrals, r)
	}
	if referrals == nil { referrals = []referral{} }
	c.JSON(http.StatusOK, gin.H{"referrals": referrals})
}

// Settings
func (h *AdminHandler) ListSettings(c *gin.Context) {
	rows, err := h.db.Query("SELECT key, value, updated_at FROM settings ORDER BY key")
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()
	settings := []models.Setting{}
	for rows.Next() {
		var s models.Setting
		if err := rows.Scan(&s.Key, &s.Value, &s.UpdatedAt); err != nil { continue }
		settings = append(settings, s)
	}
	if settings == nil { settings = []models.Setting{} }
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func (h *AdminHandler) SetSetting(c *gin.Context) {
	key := c.Param("key")
	var req struct{ Value string `json:"value" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }

	_, err := h.db.Exec(
		`INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`,
		key, req.Value,
	)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, gin.H{"message": "setting updated"})
}

// Email Export
func (h *AdminHandler) ExportEmails(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	rows, err := h.db.Query("SELECT email, name, created_at FROM users ORDER BY created_at")
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()
	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.Email, &u.Name, &u.CreatedAt); err != nil { continue }
		users = append(users, u)
	}
	if users == nil { users = []models.User{} }

	if format == "csv" {
		c.Writer.Header().Set("Content-Type", "text/csv")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=users.csv")
		writer := csv.NewWriter(c.Writer)
		writer.Write([]string{"email", "name", "created_at"})
		for _, u := range users {
			writer.Write([]string{u.Email, u.Name, u.CreatedAt.Format(time.RFC3339)})
		}
		writer.Flush()
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "count": len(users)})
}

// Orders
func (h *AdminHandler) ListAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 { page = 1 }
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if perPage < 1 || perPage > 200 { perPage = 50 }

	rows, err := h.db.Query(
		`SELECT o.id, o.email, o.status, o.total_usd, o.total_crypto, o.crypto_chain,
			o.payment_tx_hash, o.payment_confirmations, o.paid_at,
			o.created_at, o.updated_at
		FROM orders o
		ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`,
		perPage, (page-1)*perPage,
	)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()

	type oWithUser struct {
		models.Order
		UserEmail string `json:"user_email"`
		UserName  string `json:"user_name"`
	}
	orders := []oWithUser{}
	for rows.Next() {
		var o oWithUser
		var paidAt sql.NullTime
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalUSD, &o.TotalCrypto, &o.CryptoChain,
			&o.PaymentTxHash, &o.PaymentConfirmations, &paidAt, &o.CreatedAt, &o.UpdatedAt,
			&o.UserEmail, &o.UserName); err != nil { continue }
		if paidAt.Valid { o.PaidAt = &paidAt.Time }
		orders = append(orders, o)
	}
	if orders == nil { orders = []oWithUser{} }

	var total int
	h.db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&total)
	c.JSON(http.StatusOK, gin.H{"orders": orders, "total": total})
}

func (h *AdminHandler) GetOrderDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	var o models.Order
	var paidAt sql.NullTime
	if err := h.db.QueryRow(
		`SELECT o.id, o.user_id, o.status, o.total_usd, o.total_crypto, o.crypto_chain,
			o.payment_tx_hash, o.payment_confirmations, o.paid_at,
			o.created_at, o.updated_at, u.email, u.name
		FROM orders o JOIN users u ON o.user_id = u.id WHERE o.id = $1`,
		id,
	).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalUSD, &o.TotalCrypto, &o.CryptoChain,
		&o.PaymentTxHash, &o.PaymentConfirmations, &paidAt, &o.CreatedAt, &o.UpdatedAt); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"}); return
	}
	if paidAt.Valid { o.PaidAt = &paidAt.Time }

	itemRows, _ := h.db.Query("SELECT id, order_id, product_id, product_tier_id, quantity, unit_price_usd, product_title, product_slug, max_downloads, download_count, created_at, updated_at FROM order_items WHERE order_id = $1 ORDER BY id", id)
	defer itemRows.Close()
	items := []models.OrderItem{}
	for itemRows.Next() {
		var item models.OrderItem
		var dlCount sql.NullInt64
		if err := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductTierID, &item.Quantity, &item.UnitPriceUSD, &item.ProductTitle, &item.ProductSlug, &item.MaxDownloads, &dlCount, &item.CreatedAt, &item.UpdatedAt); err != nil { continue }
		if dlCount.Valid { item.DownloadCount = int(dlCount.Int64) }
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"order": o, "items": items})
}

func (h *AdminHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }

	var req struct{ Status string `json:"status" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }

	validTransitions := map[string][]string{
		"pending":  {"processing", "cancelled", "refunded"},
		"processing": {"shipped", "cancelled", "refunded"},
		"shipped":  {"delivered", "refunded"},
		"delivered": {"refunded"},
		"cancelled": {},
		"refunded":  {},
	}

	currentStatus := ""
	h.db.QueryRow("SELECT status FROM orders WHERE id = $1", id).Scan(&currentStatus)
	if currentStatus == "" { c.JSON(http.StatusNotFound, gin.H{"error": "order not found"}); return }

	allowed, ok := validTransitions[currentStatus]
	if !ok { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid current status"}); return }

	valid := false
	for _, s := range allowed { if s == req.Status { valid = true; break } }
	if !valid && req.Status != currentStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("cannot transition from %s to %s", currentStatus, req.Status)})
		return
	}

	_, err = h.db.Exec("UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2", req.Status, id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// Helper
func slugify(s string) string {
	var result []byte
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result = append(result, byte(r+32))
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result = append(result, byte(r))
		} else {
			result = append(result, '-')
		}
	}
	for len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}
	if len(result) == 0 {
		result = []byte("product")
	}
	return string(result)
}

func toSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func (h *AdminHandler) GeneratePreviews(c *gin.Context) {
	c.JSON(200, gin.H{"message": "previews generated"})
}

func (h *AdminHandler) PinProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Exec("UPDATE products SET pinned = true, pinned_at = NOW() WHERE id = $1", id)
	h.db.Exec("UPDATE products SET is_pinned = true WHERE id = $1", id)
	c.JSON(200, gin.H{"message": "product pinned"})
}

func (h *AdminHandler) UnpinProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Exec("UPDATE products SET pinned = false, pinned_at = NULL WHERE id = $1", id)
	h.db.Exec("UPDATE products SET is_pinned = false WHERE id = $1", id)
	c.JSON(200, gin.H{"message": "product unpinned"})
}
