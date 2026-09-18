package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

type orderStep struct {
	ID         int    `json:"id"`
	Slug       string `json:"slug"`
	Label      string `json:"label"`
	SortOrder  int    `json:"sort_order"`
	IsSystem   bool   `json:"is_system"`
	IsTerminal bool   `json:"is_terminal"`
}

func defaultOrderSteps() []orderStep {
	return []orderStep{
		{Slug: "created", Label: "Created", SortOrder: 10, IsSystem: true},
		{Slug: "awaiting_payment", Label: "Waiting for payment", SortOrder: 20, IsSystem: true},
		{Slug: "paid", Label: "Paid", SortOrder: 30, IsSystem: true},
		{Slug: "preparation", Label: "Preparation", SortOrder: 40},
		{Slug: "delivered", Label: "Delivered", SortOrder: 50},
		{Slug: "cancelled", Label: "Cancelled", SortOrder: 90, IsSystem: true, IsTerminal: true},
		{Slug: "refunded", Label: "Refunded", SortOrder: 91, IsSystem: true, IsTerminal: true},
		{Slug: "failed", Label: "Failed", SortOrder: 92, IsSystem: true, IsTerminal: true},
	}
}

func pipelineSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUnderscore = false
		} else if r == ' ' || r == '-' || r == '_' {
			if !lastUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

func (h *AdminHandler) ensureOrderSteps() {
	if h == nil || h.stepsReady {
		return
	}
	db := h.ordersDB()
	if db == nil {
		return
	}
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS order_steps (
		id SERIAL PRIMARY KEY,
		slug VARCHAR(64) UNIQUE NOT NULL,
		label VARCHAR(120) NOT NULL,
		sort_order INTEGER NOT NULL DEFAULT 0,
		is_system BOOLEAN NOT NULL DEFAULT false,
		is_terminal BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMPTZ DEFAULT NOW()
	)`)
	if err != nil {
		return
	}
	for _, s := range defaultOrderSteps() {
		_, _ = db.Exec(`INSERT INTO order_steps (slug, label, sort_order, is_system, is_terminal)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (slug) DO NOTHING`,
			s.Slug, s.Label, s.SortOrder, s.IsSystem, s.IsTerminal)
	}
	_, _ = db.Exec(`UPDATE orders SET status = 'awaiting_payment' WHERE status IN ('pending','processing','')`)
	_, _ = db.Exec(`UPDATE orders SET status = 'paid' WHERE status IN ('confirmed')`)
	_, _ = db.Exec(`UPDATE orders SET status = 'delivered' WHERE status IN ('shipped','completed')`)
	h.stepsReady = true
}

func (h *AdminHandler) loadOrderSteps() []orderStep {
	h.ensureOrderSteps()
	db := h.ordersDB()
	if db == nil {
		return defaultOrderSteps()
	}
	rows, err := db.Query(`SELECT id, slug, label, sort_order, is_system, is_terminal FROM order_steps ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return defaultOrderSteps()
	}
	defer rows.Close()
	out := []orderStep{}
	for rows.Next() {
		var s orderStep
		if rows.Scan(&s.ID, &s.Slug, &s.Label, &s.SortOrder, &s.IsSystem, &s.IsTerminal) == nil {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return defaultOrderSteps()
	}
	return out
}

func (h *AdminHandler) stepLabels() map[string]orderStep {
	out := map[string]orderStep{}
	for _, s := range h.loadOrderSteps() {
		out[s.Slug] = s
	}
	return out
}

func (h *AdminHandler) ListOrderSteps(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"steps": h.loadOrderSteps()})
}

func (h *AdminHandler) CreateOrderStep(c *gin.Context) {
	var req struct {
		Label      string `json:"label" binding:"required"`
		Slug       string `json:"slug"`
		IsTerminal bool   `json:"is_terminal"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	label := strings.TrimSpace(req.Label)
	if label == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "label required"})
		return
	}
	slug := pipelineSlug(req.Slug)
	if slug == "" {
		slug = pipelineSlug(label)
	}
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug"})
		return
	}
	if _, ok := h.stepLabels()[slug]; ok {
		c.JSON(http.StatusConflict, gin.H{"error": "step already exists"})
		return
	}
	var minTerm int
	_ = h.ordersDB().QueryRow(`SELECT COALESCE(MIN(sort_order), 80) FROM order_steps WHERE is_terminal = true`).Scan(&minTerm)
	if minTerm <= 0 {
		minTerm = 80
	}
	sort := minTerm - 1
	if sort < 31 {
		sort = 35
	}
	var id int
	err := h.ordersDB().QueryRow(
		`INSERT INTO order_steps (slug, label, sort_order, is_system, is_terminal) VALUES ($1, $2, $3, false, $4) RETURNING id`,
		slug, label, sort, req.IsTerminal,
	).Scan(&id)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "step already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create step"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "slug": slug, "label": label, "sort_order": sort, "is_system": false, "is_terminal": req.IsTerminal})
}

func (h *AdminHandler) UpdateOrderSteps(c *gin.Context) {
	var req struct {
		Steps []orderStep `json:"steps" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	current := h.loadOrderSteps()
	byID := map[int]orderStep{}
	for _, s := range current {
		byID[s.ID] = s
	}
	seenSystem := map[string]bool{}
	for i, s := range req.Steps {
		cur, ok := byID[s.ID]
		if !ok || s.ID == 0 {
			continue
		}
		label := strings.TrimSpace(s.Label)
		if label == "" {
			label = cur.Label
		}
		if _, err := h.ordersDB().Exec(`UPDATE order_steps SET label=$1, sort_order=$2 WHERE id=$3`, label, i+1, s.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save steps"})
			return
		}
		if cur.IsSystem {
			seenSystem[cur.Slug] = true
		}
	}
	for _, cur := range current {
		if cur.IsSystem && !seenSystem[cur.Slug] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove system step " + cur.Slug})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"steps": h.loadOrderSteps()})
}

func (h *AdminHandler) RenameOrderStep(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Label string `json:"label" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	label := strings.TrimSpace(req.Label)
	if label == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "label required"})
		return
	}
	res, err := h.ordersDB().Exec(`UPDATE order_steps SET label=$1 WHERE id=$2`, label, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "step not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "label": label})
}

func (h *AdminHandler) DeleteOrderStep(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var slug string
	var system bool
	if err := h.ordersDB().QueryRow(`SELECT slug, is_system FROM order_steps WHERE id=$1`, id).Scan(&slug, &system); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "step not found"})
		return
	}
	if system {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete a system step"})
		return
	}
	var n int
	_ = h.ordersDB().QueryRow(`SELECT COUNT(*) FROM orders WHERE status=$1`, slug).Scan(&n)
	if n > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%d orders still use this step", n)})
		return
	}
	if _, err = h.ordersDB().Exec(`DELETE FROM order_steps WHERE id=$1`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "step deleted"})
}

func (h *AdminHandler) countsByStep() []gin.H {
	h.ensureOrderSteps()
	db := h.ordersDB()
	out := []gin.H{}
	if db != nil {
		rows, err := db.Query(`SELECT s.slug, s.label, s.is_terminal, COUNT(o.id)
			FROM order_steps s
			LEFT JOIN orders o ON o.status = s.slug
			GROUP BY s.id, s.slug, s.label, s.is_terminal, s.sort_order
			ORDER BY s.sort_order ASC, s.id ASC`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var slug, label string
				var term bool
				var n int
				if rows.Scan(&slug, &label, &term, &n) == nil {
					out = append(out, gin.H{"slug": slug, "label": label, "count": n, "is_terminal": term})
				}
			}
		}
	}
	if len(out) == 0 {
		for _, s := range h.loadOrderSteps() {
			out = append(out, gin.H{"slug": s.Slug, "label": s.Label, "count": 0, "is_terminal": s.IsTerminal})
		}
	}
	if db != nil {
		var unknown int
		_ = db.QueryRow(`SELECT COUNT(*) FROM orders WHERE status IS NULL OR status NOT IN (SELECT slug FROM order_steps)`).Scan(&unknown)
		if unknown > 0 {
			out = append(out, gin.H{"slug": "_other", "label": "Other", "count": unknown, "is_terminal": false})
		}
	}
	return out
}
