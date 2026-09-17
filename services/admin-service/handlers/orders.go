package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListAllOrders returns a paginated list of all orders.
func ListAllOrders(c *gin.Context) {
	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed := parseInt(l); parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed := parseInt(o); parsed >= 0 {
			offset = parsed
		}
	}

	statusFilter := c.Query("status")

	query := `
		SELECT o.id, u.email, o.status, o.total_amount, o.created_at,
		       COUNT(oi.id) AS item_count
		FROM orders o
		JOIN users u ON u.id = o.user_id
		LEFT JOIN order_items oi ON oi.order_id = o.id
	`
	args := []interface{}{}
	if statusFilter != "" {
		query += " WHERE o.status = $1"
		args = append(args, statusFilter)
	}
	query += " GROUP BY o.id, u.email ORDER BY o.created_at DESC"
	query += " LIMIT $" + itoa(len(args)+1) + " OFFSET $" + itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer rows.Close()

	var orders []gin.H
	for rows.Next() {
		var (
			id        int
			email     string
			status    string
			total     string
			createdAt sql.NullTime
			itemCount int
		)
		if err := rows.Scan(&id, &email, &status, &total, &createdAt, &itemCount); err != nil {
			continue
		}
		orders = append(orders, gin.H{
			"id":         id,
			"user_email": email,
			"status":     status,
			"total_usd":  total,
			"created_at": nullTimeToString(createdAt),
			"item_count": itemCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":  orders,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetOrderDetail returns detailed information about a specific order.
func GetOrderDetail(c *gin.Context) {
	orderID := c.Param("id")

	var order struct {
		ID        int
		UserEmail string
		Status    string
		Total     string
		CreatedAt sql.NullTime
		Address   string
	}

	err := db.QueryRow(`
		SELECT o.id, u.email, o.status, o.total_amount, o.created_at, o.shipping_address
		FROM orders o
		JOIN users u ON u.id = o.user_id
		WHERE o.id = $1
	`, orderID).Scan(&order.ID, &order.UserEmail, &order.Status, &order.Total, &order.CreatedAt, &order.Address)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	// Fetch order items
	itemRows, err := db.Query(`
		SELECT oi.product_id, p.title, oi.quantity, oi.price
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = $1
	`, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
		return
	}
	defer itemRows.Close()

	var items []gin.H
	for itemRows.Next() {
		var (
			productID int
			title     string
			quantity  int
			price     string
		)
		if err := itemRows.Scan(&productID, &title, &quantity, &price); err != nil {
			continue
		}
		items = append(items, gin.H{
			"product_id": productID,
			"title":      title,
			"quantity":   quantity,
			"price":      price,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               order.ID,
		"user_email":       order.UserEmail,
		"status":           order.Status,
		"total_usd":        order.Total,
		"created_at":       nullTimeToString(order.CreatedAt),
		"shipping_address": order.Address,
		"items":            items,
	})
}

// UpdateOrderStatus updates the status of an order.
func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"shipped":    true,
		"delivered":  true,
		"cancelled":  true,
		"refunded":   true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	result, err := db.Exec(
		"UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2",
		req.Status, orderID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Order status updated",
		"order_id":   orderID,
		"new_status": req.Status,
	})
}


