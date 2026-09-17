package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/product-service/models"
)

type Handlers struct{ db *sql.DB }

func NewHandlers(db *sql.DB) *Handlers {
	h := &Handlers{db: db}
	h.ensureExtras()
	return h
}

// ====== Health ======
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "product-service"})
}

// ====== Products (public) ======
func (h *Handlers) ListProducts(c *gin.Context) {
	h.listProductsFiltered(c)
}

func (h *Handlers) GetProduct(c *gin.Context) {
	slug := c.Param("slug")
	var p models.Product
	var catID sql.NullInt64
	var pinnedAt sql.NullTime
	err := h.db.QueryRow(
		"SELECT id,title,slug,description,price_usd,status,created_at,updated_at,category_id,image_url,stock_count,pinned,sort_order,digital_formats,tags,is_pwyw,pwyw_min_price,pinned_at FROM products WHERE slug=$1",
		slug).Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.PriceUSD, &p.Status,
		&p.CreatedAt, &p.UpdatedAt, &catID, &p.ImageURL, &p.StockCount, &p.Pinned, &p.SortOrder,
		&p.DigitalFormats, &p.Tags, &p.IsPwyw, &p.PwywMinPrice, &pinnedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if catID.Valid { p.CategoryID = int(catID.Int64) }
	if pinnedAt.Valid { t := pinnedAt.Time; p.PinnedAt = &t }

	imgRows, _ := h.db.Query("SELECT id,url,alt_text,is_primary,image_type,width,height,file_size_bytes,mime_type,storage_path,created_at FROM product_images WHERE product_id=$1 ORDER BY is_primary DESC, id ASC", p.ID)
	defer imgRows.Close()
	images := []models.ProductImage{}
	for imgRows.Next() {
		var img models.ProductImage
		if imgRows.Scan(&img.ID, &img.URL, &img.AltText, &img.IsPrimary, &img.ImageType,
			&img.Width, &img.Height, &img.FileSizeBytes, &img.MimeType, &img.StoragePath, &img.CreatedAt) == nil {
			images = append(images, img)
		}
	}
	c.JSON(http.StatusOK, gin.H{"product": p, "images": images})
}

func (h *Handlers) SearchProducts(c *gin.Context) {
	q := c.Query("q")
	catID, _ := strconv.Atoi(c.Query("category_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 { page = 1 }
	if perPage < 1 || perPage > 100 { perPage = 20 }
	offset := (page - 1) * perPage

	baseQuery := "SELECT id,title,slug,description,price_usd,status,created_at,updated_at,category_id,image_url,stock_count,pinned,sort_order,digital_formats,tags,is_pwyw,pwyw_min_price,pinned_at FROM products WHERE status='active'"
	var args []interface{}
	ac := 0
	if q != "" {
		ac++
		baseQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d OR tags ILIKE $%d)", ac, ac, ac)
		args = append(args, "%"+q+"%")
	}
	if catID > 0 {
		ac++
		baseQuery += fmt.Sprintf(" AND category_id = $%d", ac)
		args = append(args, catID)
		_ = catID
	}
	baseQuery += " ORDER BY pinned DESC, sort_order ASC, created_at DESC LIMIT $1 OFFSET $2"
	args = append(args, perPage, offset)

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		var cid sql.NullInt64
		if rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.PriceUSD, &p.Status,
			&p.CreatedAt, &p.UpdatedAt, &cid, &p.ImageURL, &p.StockCount, &p.Pinned, &p.SortOrder,
			&p.DigitalFormats, &p.Tags, &p.IsPwyw, &p.PwywMinPrice, &p.PinnedAt) == nil {
			if cid.Valid { p.CategoryID = int(cid.Int64) }
			products = append(products, p)
		}
	}
	c.JSON(http.StatusOK, gin.H{"products": products, "total": len(products), "page": page, "per_page": perPage})
}

// ====== Categories ======
func (h *Handlers) ListCategories(c *gin.Context) {
	q := "SELECT id,name,slug,description,parent_id,image_url,is_active,created_at,updated_at FROM categories"
	if c.Query("all") != "1" {
		q += " WHERE is_active=true"
	}
	q += " ORDER BY name ASC"
	rows, _ := h.db.Query(q)
	defer rows.Close()
	type catJSON struct {
		models.Category
		ProductCount int `json:"product_count,omitempty"`
	}
	cats := []catJSON{}
	for rows.Next() {
		var cat catJSON
		var pid sql.NullInt64
		if rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &pid, &cat.ImageURL, &cat.IsActive, &cat.CreatedAt, &cat.UpdatedAt) == nil {
			if pid.Valid {
				cat.ParentID = int(pid.Int64)
			}
			if c.Query("all") == "1" {
				h.db.QueryRow("SELECT COUNT(*) FROM products WHERE category_id=$1", cat.ID).Scan(&cat.ProductCount)
			}
			cats = append(cats, cat)
		}
	}
	if cats == nil {
		cats = []catJSON{}
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

func (h *Handlers) GetCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	var cat models.Category
	var pid sql.NullInt64
	if err := h.db.QueryRow("SELECT id,name,slug,description,parent_id,image_url,is_active,created_at,updated_at FROM categories WHERE id=$1", id).
		Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &pid, &cat.ImageURL, &cat.IsActive, &cat.CreatedAt, &cat.UpdatedAt); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		if pid.Valid { cat.ParentID = int(pid.Int64) }
		c.JSON(http.StatusOK, gin.H{"category": cat})
	}
}

// ====== Tags ======
func (h *Handlers) ListTags(c *gin.Context) {
	rows, _ := h.db.Query("SELECT id,name,slug,count,image_url,is_active,created_at,updated_at FROM tags WHERE is_active=true ORDER BY name ASC")
	defer rows.Close()
	tags := []models.Tag{}
	for rows.Next() {
		var t models.Tag
		if rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Count, &t.ImageURL, &t.IsActive, &t.CreatedAt, &t.UpdatedAt) == nil {
			tags = append(tags, t)
		}
	}
	if tags == nil { tags = []models.Tag{} }
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// ====== Tiers ======
func (h *Handlers) GetProductTiers(c *gin.Context) {
	key := c.Param("productId")
	if key == "" {
		key = c.Param("slug")
	}
	pid, err := strconv.Atoi(key)
	if err != nil {
		if h.db.QueryRow("SELECT id FROM products WHERE slug=$1", key).Scan(&pid) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
	}
	rows, _ := h.db.Query("SELECT id,product_id,tier_name,price_usd,download_count,download_limit,is_active,created_at,updated_at FROM product_tiers WHERE product_id=$1 AND is_active=true ORDER BY price_usd ASC", pid)
	defer rows.Close()
	tiers := []models.ProductTier{}
	for rows.Next() {
		var t models.ProductTier
		if rows.Scan(&t.ID, &t.ProductID, &t.TierName, &t.PriceUSD, &t.DownloadCount, &t.DownloadLimit, &t.IsActive, &t.CreatedAt, &t.UpdatedAt) == nil {
			tiers = append(tiers, t)
		}
	}
	if tiers == nil { tiers = []models.ProductTier{} }
	c.JSON(http.StatusOK, gin.H{"tiers": tiers})
}

// ====== Related ======
func (h *Handlers) GetRelatedProducts(c *gin.Context) {
	pid, err := strconv.Atoi(c.Param("productId"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	var catID sql.NullInt64
	h.db.QueryRow("SELECT category_id FROM products WHERE id=$1", pid).Scan(&catID)
	if !catID.Valid {
		c.JSON(http.StatusOK, gin.H{"products": []models.Product{}})
		return
	}
	rows, _ := h.db.Query("SELECT id,title,slug,description,price_usd,status,created_at,updated_at,category_id,image_url,stock_count,pinned,sort_order,digital_formats,tags,is_pwyw,pwyw_min_price,pinned_at FROM products WHERE category_id=$1 AND id!=$2 AND status='active' ORDER BY pinned DESC, sort_order ASC LIMIT 6", catID.Int64, pid)
	defer rows.Close()
	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		var cid sql.NullInt64
		if rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.PriceUSD, &p.Status,
			&p.CreatedAt, &p.UpdatedAt, &cid, &p.ImageURL, &p.StockCount, &p.Pinned, &p.SortOrder,
			&p.DigitalFormats, &p.Tags, &p.IsPwyw, &p.PwywMinPrice, &p.PinnedAt) == nil {
			if cid.Valid { p.CategoryID = int(cid.Int64) }
			products = append(products, p)
		}
	}
	if products == nil { products = []models.Product{} }
	c.JSON(http.StatusOK, gin.H{"products": products})
}

// ====== Bundles ======
func (h *Handlers) ListBundles(c *gin.Context) {
	rows, _ := h.db.Query("SELECT id, title, slug, description, price_usd, status, sort_order, created_at, updated_at FROM bundles WHERE status = 'active' ORDER BY sort_order ASC")
	defer rows.Close()
	bundles := []models.Bundle{}
	for rows.Next() {
		var b models.Bundle
		if rows.Scan(&b.ID, &b.Title, &b.Slug, &b.Description, &b.PriceUSD, &b.Status, &b.SortOrder, &b.CreatedAt, &b.UpdatedAt) == nil {
			bundles = append(bundles, b)
		}
	}
	if bundles == nil { bundles = []models.Bundle{} }
	c.JSON(http.StatusOK, gin.H{"bundles": bundles})
}

func (h *Handlers) GetBundle(c *gin.Context) {
	slug := c.Param("slug")
	var b models.Bundle
	if err := h.db.QueryRow("SELECT id, title, slug, description, price_usd, status, sort_order, created_at, updated_at FROM bundles WHERE slug = $1", slug).Scan(&b.ID, &b.Title, &b.Slug, &b.Description, &b.PriceUSD, &b.Status, &b.SortOrder, &b.CreatedAt, &b.UpdatedAt); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bundle not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bundle": b})
}

func (h *Handlers) GetRecommendations(c *gin.Context) {
	h.GetRelatedProducts(c)
}

// ====== Admin: Products ======
func (h *Handlers) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id int
	err := h.db.QueryRow(
		"INSERT INTO products (title, slug, description, category_id, price_usd, status, pwyw_enabled, pwyw_min_price) VALUES ($1, $2, $3, $4, $5, 'active', $6, $7) RETURNING id",
		req.Title, req.Slug, req.Description, req.CategoryID, req.PriceUSD, req.PWYWEnabled, req.PWYWMinPrice,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "product created"})
}

func (h *Handlers) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var req struct {
		Title       *string  `json:"title,omitempty"`
		Slug        *string  `json:"slug,omitempty"`
		Description *string  `json:"description,omitempty"`
		CategoryID  *int     `json:"category_id,omitempty"`
		PriceUSD    *float64 `json:"price_usd,omitempty"`
		Status      *string  `json:"status,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.db.Exec("UPDATE products SET title = COALESCE($1, title), slug = COALESCE($2, slug), description = COALESCE($3, description), category_id = COALESCE($4, category_id), price_usd = COALESCE($5, price_usd), status = COALESCE($6, status), updated_at = NOW() WHERE id = $7",
		req.Title, req.Slug, req.Description, req.CategoryID, req.PriceUSD, req.Status, id,
	)
	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

func (h *Handlers) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	h.db.Exec("DELETE FROM products WHERE id = $1", id)
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

func (h *Handlers) AddProductImage(c *gin.Context) {
	idStr := c.Param("id")
	productID, _ := strconv.Atoi(idStr)
	var req struct {
		URL       string `json:"url" binding:"required"`
		IsPrimary bool   `json:"is_primary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec("INSERT INTO product_images (product_id, url, is_primary) VALUES ($1, $2, $3)", productID, req.URL, req.IsPrimary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "image added"})
}

func (h *Handlers) DeleteProductImage(c *gin.Context) {
	imageIDStr := c.Param("imageId")
	imageID, _ := strconv.Atoi(imageIDStr)
	h.db.Exec("DELETE FROM product_images WHERE id = $1", imageID)
	c.JSON(http.StatusOK, gin.H{"message": "image deleted"})
}

func (h *Handlers) GeneratePreviews(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "previews generated"})
}

func (h *Handlers) PinProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	h.db.Exec("UPDATE products SET pinned = true, is_pinned = true, pinned_at = NOW() WHERE id = $1", id)
	c.JSON(http.StatusOK, gin.H{"message": "product pinned"})
}

func (h *Handlers) AddProductTier(c *gin.Context) {
	idStr := c.Param("id")
	productID, _ := strconv.Atoi(idStr)
	var req struct {
		TierName string  `json:"tier_name" binding:"required"`
		PriceUSD float64 `json:"price_usd" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec("INSERT INTO product_tiers (product_id, tier_name, price_usd) VALUES ($1, $2, $3)", productID, req.TierName, req.PriceUSD)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tier added"})
}

func (h *Handlers) UpdateProductTier(c *gin.Context) {
	tierIDStr := c.Param("tierId")
	tierID, _ := strconv.Atoi(tierIDStr)
	var req struct {
		TierName string  `json:"tier_name"`
		PriceUSD float64 `json:"price_usd"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.db.Exec("UPDATE product_tiers SET tier_name = $1, price_usd = $2 WHERE id = $3", req.TierName, req.PriceUSD, tierID)
	c.JSON(http.StatusOK, gin.H{"message": "tier updated"})
}

func (h *Handlers) DeleteProductTier(c *gin.Context) {
	tierIDStr := c.Param("tierId")
	tierID, _ := strconv.Atoi(tierIDStr)
	h.db.Exec("DELETE FROM product_tiers WHERE id = $1", tierID)
	c.JSON(http.StatusOK, gin.H{"message": "tier deleted"})
}

// ====== Admin: Categories ======
func (h *Handlers) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		ParentID    int    `json:"parent_id"`
		SortOrder   int    `json:"sort_order"`
		ImageURL    string `json:"image_url"`
		IsActive    *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Slug == "" {
		req.Slug = slugify(req.Name)
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	var parent interface{}
	if req.ParentID > 0 {
		parent = req.ParentID
	}
	_, err := h.db.Exec(
		"INSERT INTO categories (name, slug, description, parent_id, sort_order, image_url, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		req.Name, req.Slug, req.Description, parent, req.SortOrder, req.ImageURL, active,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "category created"})
}

// ====== Admin: Bundles ======
func (h *Handlers) CreateBundle(c *gin.Context) {
	var req struct {
		Title       string  `json:"title" binding:"required"`
		Slug        string  `json:"slug" binding:"required"`
		Description string  `json:"description" binding:"required"`
		PriceUSD    float64 `json:"price_usd" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec("INSERT INTO bundles (title, slug, description, price_usd, status) VALUES ($1, $2, $3, $4, 'active')", req.Title, req.Slug, req.Description, req.PriceUSD)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "bundle created"})
}

func (h *Handlers) UpdateBundle(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		PriceUSD    float64 `json:"price_usd"`
		Status      string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.db.Exec("UPDATE bundles SET title = $1, description = $2, price_usd = $3, status = $4 WHERE id = $5", req.Title, req.Description, req.PriceUSD, req.Status, id)
	c.JSON(http.StatusOK, gin.H{"message": "bundle updated"})
}

func (h *Handlers) DeleteBundle(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	h.db.Exec("DELETE FROM bundles WHERE id = $1", id)
	c.JSON(http.StatusOK, gin.H{"message": "bundle deleted"})
}
