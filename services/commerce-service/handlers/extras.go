package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/store4bots/shared/middleware"
)

func applyCouponDiscount(discountType, discountValue string, sub float64) float64 {
	v, err := strconv.ParseFloat(discountValue, 64)
	if err != nil || v <= 0 || sub <= 0 {
		return 0
	}
	if strings.EqualFold(discountType, "percentage") {
		return math.Round(sub*v) / 100
	}
	if v > sub {
		return sub
	}
	return v
}

func productIDFromRequest(c *gin.Context) int {
	if id, err := strconv.Atoi(c.Param("id")); err == nil && id > 0 {
		return id
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return 0
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	var body map[string]any
	if json.Unmarshal(raw, &body) != nil {
		return 0
	}
	return parsePositiveInt(body["product_id"])
}

func parsePositiveInt(v any) int {
	switch t := v.(type) {
	case nil:
		return 0
	case int:
		if t > 0 {
			return t
		}
	case int64:
		if t > 0 {
			return int(t)
		}
	case float64:
		n := int(t)
		if n > 0 && float64(n) == t {
			return n
		}
		if t >= 1 {
			return int(t)
		}
	case json.Number:
		return parsePositiveInt(string(t))
	case string:
		s := strings.TrimSpace(strings.Trim(t, `"'`))
		n, err := strconv.Atoi(s)
		if err == nil && n > 0 {
			return n
		}
		f, err := strconv.ParseFloat(s, 64)
		if err == nil && f >= 1 {
			return int(f)
		}
	}
	return 0
}

func listIDPairs(db *sql.DB, query string, args ...any) ([]gin.H, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, pid int
		if rows.Scan(&id, &pid) != nil {
			continue
		}
		out = append(out, gin.H{"id": id, "product_id": pid})
	}
	return out, nil
}

func toggleOwnedProduct(db *sql.DB, table string, userID, productID int) (bool, error) {
	var existing int
	err := db.QueryRow("SELECT 1 FROM "+table+" WHERE user_id = $1 AND product_id = $2", userID, productID).Scan(&existing)
	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO "+table+" (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING", userID, productID)
		return true, err
	}
	if err != nil {
		return false, err
	}
	_, err = db.Exec("DELETE FROM "+table+" WHERE user_id = $1 AND product_id = $2", userID, productID)
	return false, err
}

func (h *CommerceHandler) UpdateCartItem(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be at least 1"})
		return
	}
	h.db.Exec("UPDATE cart_items SET quantity = $1 WHERE id = $2 AND user_id = $3", req.Quantity, id, userID)
	c.JSON(http.StatusOK, gin.H{"message": "quantity updated"})
}

func (h *CommerceHandler) orderItems(orderID int) []gin.H {
	rows, err := h.db.Query(
		`SELECT id, order_id, product_id, quantity, price_usd FROM order_items WHERE order_id = $1 ORDER BY id`,
		orderID,
	)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()
	items := []gin.H{}
	for rows.Next() {
		var id, oid, pid, qty int
		var price float64
		if rows.Scan(&id, &oid, &pid, &qty, &price) != nil {
			continue
		}
		items = append(items, gin.H{
			"id":              id,
			"order_id":        oid,
			"product_id":      pid,
			"product_title":   "",
			"product_slug":    "",
			"quantity":        qty,
			"unit_price_usd":  price,
			"download_count":  0,
		})
	}
	if items == nil {
		items = []gin.H{}
	}
	return items
}
