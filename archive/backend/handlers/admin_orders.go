package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

// ListAllOrders returns every order (admin only).
func ListAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count orders"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT o.id, o.user_id, o.status, o.total_usd, o.total_crypto, o.crypto_chain,
		       o.payment_address, o.payment_tx_hash, o.payment_confirmations,
		       o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at,
		       u.email, u.name
		FROM orders o JOIN users u ON o.user_id = u.id
		ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalUSD, &o.TotalCrypto,
			&o.CryptoChain, &o.PaymentAddress, &o.PaymentTxHash, &o.PaymentConfirmations,
			&o.PaymentConfirmedAt, &o.PaidAt, &o.CreatedAt, &o.UpdatedAt,
			&o.UserEmail, &o.UserName); err != nil {
			continue
		}
		o.Items = getOrderItems(o.ID)
		orders = append(orders, o)
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders, "total": total, "page": page, "per_page": perPage})
}

// GetAnyOrder returns a single order by ID (admin only, any user).
func GetAnyOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}
	var o models.Order
	err = database.DB.QueryRow(`
		SELECT o.id, o.user_id, o.status, o.total_usd, o.total_crypto, o.crypto_chain,
		       o.payment_address, o.payment_tx_hash, o.payment_confirmations,
		       o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at,
		       u.email, u.name
		FROM orders o JOIN users u ON o.user_id = u.id
		WHERE o.id = $1
	`, id).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalUSD, &o.TotalCrypto,
		&o.CryptoChain, &o.PaymentAddress, &o.PaymentTxHash, &o.PaymentConfirmations,
		&o.PaymentConfirmedAt, &o.PaidAt, &o.CreatedAt, &o.UpdatedAt,
		&o.UserEmail, &o.UserName)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
		return
	}
	o.Items = getOrderItems(o.ID)
	c.JSON(http.StatusOK, gin.H{"order": o})
}
