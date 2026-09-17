package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetProductStats returns aggregated statistics for products.
func GetProductStats(c *gin.Context) {
	var stats struct {
		TotalProducts  int
		ActiveProducts int
		TotalRevenue   string
		TotalOrders    int
	}

	err := db.QueryRow(`
		SELECT
			COUNT(*) AS total_products,
			COUNT(*) FILTER (WHERE status = 'active') AS active_products,
			COALESCE(SUM(price), 0) AS total_revenue,
			(SELECT COUNT(*) FROM orders) AS total_orders
		FROM products
	`).Scan(&stats.TotalProducts, &stats.ActiveProducts, &stats.TotalRevenue, &stats.TotalOrders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_products":  stats.TotalProducts,
		"active_products": stats.ActiveProducts,
		"total_revenue":   stats.TotalRevenue,
		"total_orders":    stats.TotalOrders,
	})
}

// BulkUpdateProducts performs bulk operations on products (activate, deactivate, delete).
func BulkUpdateProducts(c *gin.Context) {
	var req struct {
		ProductIDs []int  `json:"product_ids" binding:"required"`
		Action     string `json:"action" binding:"required"`
		Status     string `json:"status,omitempty"`
		CategoryID *int   `json:"category_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(req.ProductIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_ids cannot be empty"})
		return
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(req.ProductIDs))
	args := make([]interface{}, len(req.ProductIDs))
	for i, id := range req.ProductIDs {
		placeholders[i] = "$" + itoa(i+1)
		args[i] = id
	}

	var result sql.Result
	var err error

	switch req.Action {
	case "activate":
		query := "UPDATE products SET status = 'active', updated_at = NOW() WHERE id IN (" + strings.Join(placeholders, ",") + ")"
		result, err = db.Exec(query, args...)
	case "deactivate":
		query := "UPDATE products SET status = 'inactive', updated_at = NOW() WHERE id IN (" + strings.Join(placeholders, ",") + ")"
		result, err = db.Exec(query, args...)
	case "delete":
		query := "DELETE FROM products WHERE id IN (" + strings.Join(placeholders, ",") + ")"
		result, err = db.Exec(query, args...)
	case "update_status":
		if req.Status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status is required for update_status action"})
			return
		}
		query := "UPDATE products SET status = $" + itoa(len(args)+1) + ", updated_at = NOW() WHERE id IN (" + strings.Join(placeholders, ",") + ")"
		result, err = db.Exec(query, append(args, req.Status)...)
	case "update_category":
		if req.CategoryID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category_id is required for update_category action"})
			return
		}
		query := "UPDATE products SET category_id = $" + itoa(len(args)+1) + ", updated_at = NOW() WHERE id IN (" + strings.Join(placeholders, ",") + ")"
		result, err = db.Exec(query, append(args, *req.CategoryID)...)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Must be one of: activate, deactivate, delete, update_status, update_category"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform bulk update"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"message":       "Bulk update completed",
		"action":        req.Action,
		"rows_affected": rowsAffected,
	})
}


