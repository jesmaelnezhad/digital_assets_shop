package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/middleware"
)

func productIDFromRequest(c *gin.Context) int {
	if id, err := strconv.Atoi(c.Param("id")); err == nil && id > 0 {
		return id
	}
	var body struct {
		ProductID int `json:"product_id"`
	}
	_ = c.ShouldBindJSON(&body)
	return body.ProductID
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
