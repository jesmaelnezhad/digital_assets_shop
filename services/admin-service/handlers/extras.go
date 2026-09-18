package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/auth"
	"github.com/pawradise/shared/middleware"
)

func (h *AdminHandler) ensureExtras() {
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
}

func (h *AdminHandler) UseCommerceDB(db *sql.DB) {
	if db != nil {
		h.commerceDB = db
		h.stepsReady = false
		h.ensureOrderSteps()
	}
}

func (h *AdminHandler) ordersDB() *sql.DB {
	if h.commerceDB != nil {
		return h.commerceDB
	}
	return h.db
}

func (h *AdminHandler) CreateProductRequest(c *gin.Context) {
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

func (h *AdminHandler) ListProductRequests(c *gin.Context) {
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
				r.Created = t.Time.Format(time.RFC3339)
			}
			out = append(out, r)
		}
	}
	c.JSON(http.StatusOK, gin.H{"requests": out})
}

func (h *AdminHandler) UpdateProductRequest(c *gin.Context) {
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

func (h *AdminHandler) ListGuestOrders(c *gin.Context) {
	db := h.ordersDB()
	rows, err := db.Query(`SELECT id, COALESCE(email,''), total_usd, status, COALESCE(crypto_chain,''), created_at
		FROM orders WHERE COALESCE(email,'') <> '' AND (user_id IS NULL OR user_id = 0)
		ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"orders": []gin.H{}, "total": 0})
		return
	}
	defer rows.Close()
	type row struct {
		ID          int     `json:"id"`
		Email       string  `json:"email"`
		TotalUSD    float64 `json:"total_usd"`
		Status      string  `json:"status"`
		CryptoChain string  `json:"crypto_chain"`
		CreatedAt   string  `json:"created_at"`
	}
	orders := []row{}
	for rows.Next() {
		var r row
		var t time.Time
		if rows.Scan(&r.ID, &r.Email, &r.TotalUSD, &r.Status, &r.CryptoChain, &t) == nil {
			r.CreatedAt = t.Format(time.RFC3339)
			orders = append(orders, r)
		}
	}
	if orders == nil {
		orders = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders, "total": len(orders)})
}

func (h *AdminHandler) GetGuestOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	db := h.ordersDB()
	var email, status string
	var total float64
	var created time.Time
	if err := db.QueryRow(`SELECT COALESCE(email,''), total_usd, status, created_at FROM orders WHERE id=$1`, id).
		Scan(&email, &total, &status, &created); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	items := []gin.H{}
	rows, err := db.Query(`SELECT id, product_id, quantity, price_usd FROM order_items WHERE order_id=$1`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var iid, pid, qty int
			var price float64
			if rows.Scan(&iid, &pid, &qty, &price) == nil {
				items = append(items, gin.H{"id": iid, "product_id": pid, "quantity": qty, "unit_price_usd": price})
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"order": gin.H{"id": id, "total_usd": total, "status": status, "created_at": created.Format(time.RFC3339)},
		"email": email,
		"items": items,
	})
}

func (h *AdminHandler) ListAdminCategories(c *gin.Context) {
	rows, err := h.db.Query(`SELECT c.id, c.name, c.slug, COALESCE(c.description,''), COALESCE(c.parent_id,0),
		COALESCE(c.image_url,''), COALESCE(c.is_active,true), c.created_at, c.updated_at,
		(SELECT COUNT(*) FROM products p WHERE p.category_id = c.id) AS product_count
		FROM categories c ORDER BY c.name ASC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	defer rows.Close()
	cats := []gin.H{}
	for rows.Next() {
		var id, parent, count int
		var name, slug, desc, img string
		var active bool
		var created, updated time.Time
		if rows.Scan(&id, &name, &slug, &desc, &parent, &img, &active, &created, &updated, &count) == nil {
			cats = append(cats, gin.H{
				"id": id, "name": name, "slug": slug, "description": desc, "parent_id": parent,
				"image_url": img, "is_active": active, "created_at": created, "updated_at": updated,
				"product_count": count,
			})
		}
	}
	if cats == nil {
		cats = []gin.H{}
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

func (h *AdminHandler) SetUserAccess(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Role      string          `json:"role"`
		StaffTabs json.RawMessage `json:"staff_tabs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := auth.NormalizeRole(req.Role)
	tabs := ""
	if role == "staff" {
		parsed := []string{}
		if len(req.StaffTabs) > 0 && string(req.StaffTabs) != "null" {
			if req.StaffTabs[0] == '[' {
				_ = json.Unmarshal(req.StaffTabs, &parsed)
			} else {
				var s string
				if json.Unmarshal(req.StaffTabs, &s) == nil {
					parsed = strings.Split(s, ",")
				}
			}
		}
		tabs = middleware.JoinTabs(parsed)
	}

	var current string
	if err := h.identityDB.QueryRow(`SELECT COALESCE(role,'customer') FROM users WHERE id=$1`, id).Scan(&current); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	current = auth.NormalizeRole(current)
	if current == "admin" && role != "admin" {
		var n int
		_ = h.identityDB.QueryRow(`SELECT COUNT(*) FROM users WHERE role='admin'`).Scan(&n)
		if n <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot demote the last admin"})
			return
		}
	}

	_, err = h.identityDB.Exec(`UPDATE users SET role=$1, staff_tabs=$2, updated_at=NOW() WHERE id=$3`, role, tabs, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update access"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "role": role, "staff_tabs": tabs, "message": "access updated"})
}
