package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/product-service/models"
)

func (h *Handlers) ensureExtras() {
	_, _ = h.db.Exec(`CREATE TABLE IF NOT EXISTS product_requests (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT DEFAULT '',
		category TEXT DEFAULT '',
		budget_usd DECIMAL(12,2) DEFAULT 0,
		email VARCHAR(255) DEFAULT '',
		status VARCHAR(50) DEFAULT 'open',
		user_id INTEGER,
		created_at TIMESTAMPTZ DEFAULT NOW()
	)`)
	_, _ = h.db.Exec(`ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS tier_name VARCHAR(255) DEFAULT ''`)
	_, _ = h.db.Exec(`ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS price_usd DECIMAL(12,2) DEFAULT 0`)
	_, _ = h.db.Exec(`ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS download_count INTEGER DEFAULT 0`)
	_, _ = h.db.Exec(`ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS download_limit INTEGER DEFAULT 0`)
	_, _ = h.db.Exec(`ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true`)
	_, _ = h.db.Exec(`ALTER TABLE product_tiers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW()`)
	_, _ = h.db.Exec(`UPDATE product_tiers SET tier_name = name WHERE (tier_name IS NULL OR tier_name = '') AND name IS NOT NULL`)
	_, _ = h.db.Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS banner_sort INTEGER NOT NULL DEFAULT 0`)
	_, _ = h.db.Exec(`UPDATE products SET banner_sort = sub.rn FROM (
		SELECT id, ROW_NUMBER() OVER (ORDER BY sort_order ASC, id ASC) AS rn
		FROM products WHERE pinned = true AND COALESCE(banner_sort,0) = 0
	) sub WHERE products.id = sub.id`)
	_, _ = h.db.Exec(`CREATE TABLE IF NOT EXISTS site_appearance (
		id INTEGER PRIMARY KEY DEFAULT 1,
		palette TEXT NOT NULL DEFAULT 'clay',
		font TEXT NOT NULL DEFAULT 'system',
		radius TEXT NOT NULL DEFAULT 'soft',
		density TEXT NOT NULL DEFAULT 'comfortable',
		icons TEXT NOT NULL DEFAULT 'line',
		contrast TEXT NOT NULL DEFAULT 'standard',
		grain TEXT NOT NULL DEFAULT 'light',
		glow TEXT NOT NULL DEFAULT 'halo',
		motion TEXT NOT NULL DEFAULT 'gentle',
		tracking TEXT NOT NULL DEFAULT 'normal',
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`)
	_, _ = h.db.Exec(`ALTER TABLE site_appearance ADD COLUMN IF NOT EXISTS icons TEXT NOT NULL DEFAULT 'line'`)
	_, _ = h.db.Exec(`ALTER TABLE site_appearance ADD COLUMN IF NOT EXISTS contrast TEXT NOT NULL DEFAULT 'standard'`)
	_, _ = h.db.Exec(`ALTER TABLE site_appearance ADD COLUMN IF NOT EXISTS grain TEXT NOT NULL DEFAULT 'light'`)
	_, _ = h.db.Exec(`ALTER TABLE site_appearance ADD COLUMN IF NOT EXISTS glow TEXT NOT NULL DEFAULT 'halo'`)
	_, _ = h.db.Exec(`ALTER TABLE site_appearance ADD COLUMN IF NOT EXISTS motion TEXT NOT NULL DEFAULT 'gentle'`)
	_, _ = h.db.Exec(`ALTER TABLE site_appearance ADD COLUMN IF NOT EXISTS tracking TEXT NOT NULL DEFAULT 'normal'`)
	_, _ = h.db.Exec(`INSERT INTO site_appearance (id) VALUES (1) ON CONFLICT (id) DO NOTHING`)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "category"
	}
	return out
}

func (h *Handlers) bundleItemCSV(bundleID int) string {
	rows, err := h.db.Query("SELECT product_id FROM bundle_items WHERE bundle_id=$1 ORDER BY sort_order ASC, id ASC", bundleID)
	if err != nil {
		return ""
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id int
		if rows.Scan(&id) == nil && id > 0 {
			ids = append(ids, strconv.Itoa(id))
		}
	}
	return strings.Join(ids, ",")
}

func (h *Handlers) listProductsFiltered(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "12"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 12
	}
	offset := (page - 1) * perPage

	where := "p.status = 'active'"
	args := []interface{}{}
	n := 0

	search := c.Query("search")
	if search == "" {
		search = c.Query("q")
	}
	if search != "" {
		n++
		where += fmt.Sprintf(" AND (p.title ILIKE $%d OR COALESCE(p.description,'') ILIKE $%d OR COALESCE(p.tags,'') ILIKE $%d)", n, n, n)
		args = append(args, "%"+search+"%")
	}

	cat := c.Query("category")
	if cat == "" {
		cat = c.Query("category_id")
	}
	if cat != "" {
		if id, err := strconv.Atoi(cat); err == nil && id > 0 {
			n++
			where += fmt.Sprintf(" AND p.category_id = $%d", n)
			args = append(args, id)
		} else {
			n++
			where += fmt.Sprintf(" AND p.category_id = (SELECT id FROM categories WHERE slug = $%d LIMIT 1)", n)
			args = append(args, cat)
		}
	}
	if ft := c.Query("file_type"); ft != "" {
		n++
		where += fmt.Sprintf(" AND COALESCE(p.digital_formats,'') ILIKE $%d", n)
		args = append(args, "%"+ft+"%")
	}
	if mn := c.Query("price_min"); mn != "" {
		if v, err := strconv.ParseFloat(mn, 64); err == nil {
			n++
			where += fmt.Sprintf(" AND p.price_usd >= $%d", n)
			args = append(args, v)
		}
	}
	if mx := c.Query("price_max"); mx != "" {
		if v, err := strconv.ParseFloat(mx, 64); err == nil {
			n++
			where += fmt.Sprintf(" AND p.price_usd <= $%d", n)
			args = append(args, v)
		}
	}
	if rt := c.Query("rating"); rt != "" {
		if v, err := strconv.ParseFloat(rt, 64); err == nil && v > 0 {
			n++
			where += fmt.Sprintf(" AND COALESCE(p.average_rating, 0) >= $%d", n)
			args = append(args, v)
		}
	}
	bannerOnly := c.Query("banner") == "1" || c.Query("banner") == "true"
	if bannerOnly {
		where += " AND COALESCE(p.banner_sort,0) > 0"
	}

	order := "p.pinned DESC, p.sort_order ASC, p.created_at DESC"
	if bannerOnly {
		order = "p.banner_sort ASC, p.id ASC"
	} else {
		switch c.Query("sort") {
		case "popular":
			order = "p.pinned DESC, COALESCE(p.purchase_count,0) DESC, COALESCE(p.views,0) DESC, p.created_at DESC"
		case "price_asc":
			order = "p.pinned DESC, p.price_usd ASC"
		case "price_desc":
			order = "p.pinned DESC, p.price_usd DESC"
		}
	}

	var total int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM products p WHERE "+where, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	n++
	limP := n
	n++
	offP := n
	q := `SELECT p.id,p.title,p.slug,p.description,p.price_usd,p.status,p.created_at,p.updated_at,
		p.category_id,p.image_url,p.stock_count,p.pinned,p.sort_order,p.digital_formats,p.tags,
		p.is_pwyw,p.pwyw_min_price,p.pinned_at,COALESCE(p.banner_sort,0)
		FROM products p WHERE ` + where + ` ORDER BY ` + order + fmt.Sprintf(" LIMIT $%d OFFSET $%d", limP, offP)
	qargs := append(append([]interface{}{}, args...), perPage, offset)
	rows, err := h.db.Query(q, qargs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		var cid sql.NullInt64
		var pinnedAt sql.NullTime
		if rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.PriceUSD, &p.Status,
			&p.CreatedAt, &p.UpdatedAt, &cid, &p.ImageURL, &p.StockCount, &p.Pinned, &p.SortOrder,
			&p.DigitalFormats, &p.Tags, &p.IsPwyw, &p.PwywMinPrice, &pinnedAt, &p.BannerSort) == nil {
			if cid.Valid {
				p.CategoryID = int(cid.Int64)
			}
			if pinnedAt.Valid {
				t := pinnedAt.Time
				p.PinnedAt = &t
			}
			p.IsPinned = p.Pinned
			products = append(products, p)
		}
	}
	if products == nil {
		products = []models.Product{}
	}
	c.JSON(http.StatusOK, gin.H{"products": products, "total": total, "page": page, "per_page": perPage})
}

func (h *Handlers) UnpinProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Exec("UPDATE products SET pinned = false, is_pinned = false, pinned_at = NULL, banner_sort = 0 WHERE id = $1", id)
	c.JSON(http.StatusOK, gin.H{"message": "product unpinned"})
}

func (h *Handlers) SetBanner(c *gin.Context) {
	var req struct {
		ProductIDs []int `json:"product_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, _ = h.db.Exec("UPDATE products SET banner_sort = 0")
	for i, id := range req.ProductIDs {
		if id <= 0 {
			continue
		}
		_, _ = h.db.Exec(`UPDATE products SET banner_sort = $1, pinned = true, is_pinned = true, pinned_at = NOW() WHERE id = $2`, i+1, id)
	}
	c.JSON(http.StatusOK, gin.H{"product_ids": req.ProductIDs, "message": "banner updated"})
}

func (h *Handlers) UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Name        string `json:"name"`
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
	if req.Slug == "" && req.Name != "" {
		req.Slug = slugify(req.Name)
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	} else {
		_ = h.db.QueryRow("SELECT is_active FROM categories WHERE id=$1", id).Scan(&active)
	}
	var parent interface{}
	if req.ParentID > 0 {
		parent = req.ParentID
	}
	_, err = h.db.Exec(`UPDATE categories SET name=COALESCE(NULLIF($1,''), name), slug=COALESCE(NULLIF($2,''), slug),
		description=$3, parent_id=$4, sort_order=$5, image_url=$6, is_active=$7, updated_at=NOW() WHERE id=$8`,
		req.Name, req.Slug, req.Description, parent, req.SortOrder, req.ImageURL, active, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category updated"})
}

func (h *Handlers) DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var used, children int
	h.db.QueryRow("SELECT COUNT(*) FROM products WHERE category_id=$1", id).Scan(&used)
	h.db.QueryRow("SELECT COUNT(*) FROM categories WHERE parent_id=$1", id).Scan(&children)
	if used > 0 || children > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category still in use"})
		return
	}
	res, err := h.db.Exec("DELETE FROM categories WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}

func (h *Handlers) CreateProductRequest(c *gin.Context) {
	var req struct {
		Title       string  `json:"title" binding:"required"`
		Description string  `json:"description"`
		Category    string  `json:"category"`
		BudgetUSD   float64 `json:"budget_usd"`
		Email       string  `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id int
	err := h.db.QueryRow(
		`INSERT INTO product_requests (title, description, category, budget_usd, email, status)
		 VALUES ($1,$2,$3,$4,$5,'open') RETURNING id`,
		req.Title, req.Description, req.Category, req.BudgetUSD, req.Email,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit request"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "request submitted"})
}

func (h *Handlers) ListProductRequests(c *gin.Context) {
	email := c.Query("email")
	q := `SELECT id, title, description, category, budget_usd, email, status, created_at FROM product_requests`
	args := []interface{}{}
	if email != "" {
		q += " WHERE email = $1"
		args = append(args, email)
	}
	q += " ORDER BY created_at DESC"
	rows, err := h.db.Query(q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	defer rows.Close()
	type row struct {
		ID          int     `json:"id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Category    string  `json:"category"`
		BudgetUSD   float64 `json:"budget_usd"`
		Email       string  `json:"email"`
		Status      string  `json:"status"`
		Created     string  `json:"created"`
	}
	out := []row{}
	for rows.Next() {
		var r row
		var t sql.NullTime
		if rows.Scan(&r.ID, &r.Title, &r.Description, &r.Category, &r.BudgetUSD, &r.Email, &r.Status, &t) == nil {
			if t.Valid {
				r.Created = t.Time.Format("2006-01-02T15:04:05Z07:00")
			}
			out = append(out, r)
		}
	}
	c.JSON(http.StatusOK, gin.H{"requests": out})
}

func (h *Handlers) UpdateProductRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = h.db.Exec("UPDATE product_requests SET status=$1 WHERE id=$2", req.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "request updated"})
}

type siteAppearance struct {
	Palette  string `json:"palette"`
	Font     string `json:"font"`
	Radius   string `json:"radius"`
	Density  string `json:"density"`
	Icons    string `json:"icons"`
	Contrast string `json:"contrast"`
	Grain    string `json:"grain"`
	Glow     string `json:"glow"`
	Motion   string `json:"motion"`
	Tracking string `json:"tracking"`
}

var appearancePalettes = []string{
	"clay", "marble", "night", "moss", "ink", "ember", "dune", "frost",
	"paper", "chalk", "linen", "mist", "petal", "foam", "porcelain", "sage",
	"snow", "honey", "bone", "cloud", "wine", "violet", "ocean", "slate",
}

func defaultAppearance() siteAppearance {
	return siteAppearance{
		Palette: "clay", Font: "system", Radius: "soft", Density: "comfortable",
		Icons: "line", Contrast: "standard", Grain: "light", Glow: "halo",
		Motion: "gentle", Tracking: "normal",
	}
}

func pickAllowed(v string, allowed []string, fallback string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	for _, a := range allowed {
		if v == a {
			return a
		}
	}
	return fallback
}

func normalizeAppearance(t siteAppearance) siteAppearance {
	t.Palette = pickAllowed(t.Palette, appearancePalettes, "clay")
	t.Font = pickAllowed(t.Font, []string{"system", "humanist", "serif", "mono", "display"}, "system")
	t.Radius = pickAllowed(t.Radius, []string{"sharp", "soft", "round"}, "soft")
	t.Density = pickAllowed(t.Density, []string{"compact", "comfortable", "roomy"}, "comfortable")
	t.Icons = pickAllowed(t.Icons, []string{"line", "bold", "filled", "glyph"}, "line")
	t.Contrast = pickAllowed(t.Contrast, []string{"soft", "standard", "punchy"}, "standard")
	t.Grain = pickAllowed(t.Grain, []string{"off", "light", "heavy"}, "light")
	t.Glow = pickAllowed(t.Glow, []string{"none", "halo", "bloom"}, "halo")
	t.Motion = pickAllowed(t.Motion, []string{"still", "gentle"}, "gentle")
	t.Tracking = pickAllowed(t.Tracking, []string{"tight", "normal", "wide"}, "normal")
	return t
}

func (h *Handlers) GetAppearance(c *gin.Context) {
	t := defaultAppearance()
	err := h.db.QueryRow(`SELECT palette, font, radius, density,
		COALESCE(icons,'line'), COALESCE(contrast,'standard'), COALESCE(grain,'light'),
		COALESCE(glow,'halo'), COALESCE(motion,'gentle'), COALESCE(tracking,'normal')
		FROM site_appearance WHERE id=1`).
		Scan(&t.Palette, &t.Font, &t.Radius, &t.Density, &t.Icons, &t.Contrast, &t.Grain, &t.Glow, &t.Motion, &t.Tracking)
	if err != nil {
		_ = h.db.QueryRow(`SELECT palette, font, radius, density FROM site_appearance WHERE id=1`).
			Scan(&t.Palette, &t.Font, &t.Radius, &t.Density)
	}
	c.JSON(http.StatusOK, normalizeAppearance(t))
}

func (h *Handlers) SetAppearance(c *gin.Context) {
	var t siteAppearance
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t = normalizeAppearance(t)
	_, err := h.db.Exec(`INSERT INTO site_appearance (id, palette, font, radius, density, icons, contrast, grain, glow, motion, tracking, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (id) DO UPDATE SET palette=$1, font=$2, radius=$3, density=$4, icons=$5, contrast=$6, grain=$7, glow=$8, motion=$9, tracking=$10, updated_at=NOW()`,
		t.Palette, t.Font, t.Radius, t.Density, t.Icons, t.Contrast, t.Grain, t.Glow, t.Motion, t.Tracking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save appearance"})
		return
	}
	c.JSON(http.StatusOK, t)
}
