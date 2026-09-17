package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"backend/database"
	"backend/models"
)

// GetCategories returns all categories
func GetCategories(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, name, slug, description, parent_id, created_at, updated_at
		FROM categories ORDER BY name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &cat.ParentID, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			continue
		}
		categories = append(categories, cat)
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// GetProducts returns paginated products with optional search and category filter
func GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "12"))
	if perPage < 1 || perPage > 50 {
		perPage = 12
	}

	search := c.Query("search")
	categorySlug := c.Query("category")
	status := c.DefaultQuery("status", "active")

	var whereClause string
	var args []interface{}
	argCount := 0

	whereClause = "WHERE p.status = $1"
	argCount++
	args = append(args, status)

	if search != "" {
		whereClause += fmt.Sprintf(" AND p.search_vector @@ plainto_tsquery('english', $%d)", argCount+1)
		args = append(args, search)
		argCount++
	}

	if categorySlug != "" {
		whereClause += fmt.Sprintf(" AND c.slug = $%d", argCount+1)
		args = append(args, categorySlug)
		argCount++
	}

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id %s", whereClause)
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count products"})
		return
	}

	// Fetch products
	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd,
		       p.asset_path, p.asset_hash, p.status, p.download_count_limit,
		       p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type,
		       p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount+1, argCount+2)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.CategoryID,
			&p.CategoryName, &p.PriceUSD, &p.AssetPath, &p.AssetHash, &p.Status,
			&p.DownloadCountLimit, &p.MaxDownloadsPerUser, &p.FileSizeBytes, &p.FileMimeType,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, models.ProductListResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	})
}

// GetProduct returns a single product by slug
func GetProduct(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		slug = c.Param("id")
	}

	var p models.Product
	err := database.DB.QueryRow(`
		SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd,
		       p.asset_path, p.asset_hash, p.status, p.download_count_limit,
		       p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type,
		       p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.slug = $1 AND p.status = 'active'
	`, slug).Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.CategoryID,
		&p.CategoryName, &p.PriceUSD, &p.AssetPath, &p.AssetHash, &p.Status,
		&p.DownloadCountLimit, &p.MaxDownloadsPerUser, &p.FileSizeBytes, &p.FileMimeType,
		&p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch product"})
		return
	}

	// Fetch images
	imageRows, err := database.DB.Query(`
		SELECT id, product_id, url, is_primary, created_at
		FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id
	`, p.ID)
	if err == nil {
		defer imageRows.Close()
		images := []models.ProductImage{}
		for imageRows.Next() {
			var img models.ProductImage
			if imageRows.Scan(&img.ID, &img.ProductID, &img.URL, &img.IsPrimary, &img.CreatedAt) == nil {
				images = append(images, img)
			}
		}
		if len(images) > 0 {
			c.JSON(http.StatusOK, gin.H{"product": p, "images": images})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"product": p})
}

// CreateProduct creates a new product (admin only)
func CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Validate price
	if req.PriceUSD <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be greater than 0"})
		return
	}

	// If slug is not provided, generate from title
	if req.Slug == "" {
		req.Slug = strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
		// Clean slug
		req.Slug = strings.ReplaceAll(req.Slug, "--", "-")
	}

	// Check slug uniqueness
	var existing int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM products WHERE slug = $1", req.Slug).Scan(&existing); err == nil && existing > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug already exists"})
		return
	}

	// Validate category
	if req.CategoryID != nil && *req.CategoryID > 0 {
		var catCount int
		if err := database.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE id = $1", *req.CategoryID).Scan(&catCount); err != nil || catCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
			return
		}
	}

	status := req.Status
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "draft" {
		status = "active"
	}

	priceStr := fmt.Sprintf("%.2f", req.PriceUSD)

	var productID int
	err := database.DB.QueryRow(`
		INSERT INTO products (title, slug, description, category_id, price_usd, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.Title, req.Slug, req.Description, req.CategoryID, priceStr, status).Scan(&productID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product: " + err.Error()})
		return
	}

	// Fetch created product
	var p models.Product
	database.DB.QueryRow(`
		SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd,
		       p.asset_path, p.asset_hash, p.status, p.download_count_limit,
		       p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type,
		       p.created_at, p.updated_at
		FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id = $1
	`, productID).Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.CategoryID,
		&p.CategoryName, &p.PriceUSD, &p.AssetPath, &p.AssetHash, &p.Status,
		&p.DownloadCountLimit, &p.MaxDownloadsPerUser, &p.FileSizeBytes, &p.FileMimeType,
		&p.CreatedAt, &p.UpdatedAt)

	c.JSON(http.StatusCreated, gin.H{"product": p})
}

// UpdateProduct updates an existing product (admin only)
func UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Check product exists
	var existing models.Product
	err = database.DB.QueryRow(`
		SELECT id, title, slug, description, category_id, price_usd, status
		FROM products WHERE id = $1
	`, id).Scan(&existing.ID, &existing.Title, &existing.Slug, &existing.Description,
		&existing.CategoryID, &existing.PriceUSD, &existing.Status)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch product"})
		return
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argNum := 1

	if req.Title != "" {
		updates = append(updates, fmt.Sprintf("title = $%d", argNum))
		args = append(args, req.Title)
		argNum++
	}
	if req.Slug != "" {
		// Check uniqueness
		var slugCount int
		database.DB.QueryRow("SELECT COUNT(*) FROM products WHERE slug = $1 AND id != $2", req.Slug, id).Scan(&slugCount)
		if slugCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "slug already exists"})
			return
		}
		updates = append(updates, fmt.Sprintf("slug = $%d", argNum))
		args = append(args, req.Slug)
		argNum++
	}
	if req.Description != "" {
		updates = append(updates, fmt.Sprintf("description = $%d", argNum))
		args = append(args, req.Description)
		argNum++
	}
	if req.CategoryID != nil {
		if *req.CategoryID > 0 {
			var catCount int
			database.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE id = $1", *req.CategoryID).Scan(&catCount)
			if catCount == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
				return
			}
		}
		updates = append(updates, fmt.Sprintf("category_id = $%d", argNum))
		args = append(args, req.CategoryID)
		argNum++
	}
	if req.PriceUSD > 0 {
		updates = append(updates, fmt.Sprintf("price_usd = $%d", argNum))
		args = append(args, fmt.Sprintf("%.2f", req.PriceUSD))
		argNum++
	}
	if req.Status != "" {
		if req.Status == "active" || req.Status == "draft" || req.Status == "archived" {
			updates = append(updates, fmt.Sprintf("status = $%d", argNum))
			args = append(args, req.Status)
			argNum++
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d", strings.Join(updates, ", "), argNum)

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product: " + err.Error()})
		return
	}

	// Fetch updated product
	var p models.Product
	database.DB.QueryRow(`
		SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd,
		       p.asset_path, p.asset_hash, p.status, p.download_count_limit,
		       p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type,
		       p.created_at, p.updated_at
		FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id = $1
	`, id).Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.CategoryID,
		&p.CategoryName, &p.PriceUSD, &p.AssetPath, &p.AssetHash, &p.Status,
		&p.DownloadCountLimit, &p.MaxDownloadsPerUser, &p.FileSizeBytes, &p.FileMimeType,
		&p.CreatedAt, &p.UpdatedAt)

	c.JSON(http.StatusOK, gin.H{"product": p})
}

// DeleteProduct soft-deletes a product (admin only) — sets status to archived
func DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	result, err := database.DB.Exec("UPDATE products SET status = 'archived' WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product archived successfully"})
}

// GetProductStats returns product statistics (admin)
func GetProductStats(c *gin.Context) {
	var totalProducts, activeProducts, totalCategories int
	database.DB.QueryRow("SELECT COUNT(*) FROM products WHERE status = 'active'").Scan(&activeProducts)
	database.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
	database.DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&totalCategories)

	c.JSON(http.StatusOK, gin.H{
		"total_products":    totalProducts,
		"active_products":   activeProducts,
		"total_categories":  totalCategories,
	})
}

// RegisterProductRoutes registers product-related routes
func RegisterProductRoutes(r *gin.RouterGroup) {
	// Public product routes
	r.GET("/products", GetProducts)
	r.GET("/products/:slug", GetProduct)
	r.GET("/categories", GetCategories)
}
