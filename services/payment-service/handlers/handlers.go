package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/auth"
	"github.com/pawradise/shared/models"
)

type PaymentHandler struct{ db *sql.DB }

func NewPaymentHandler(db *sql.DB) *PaymentHandler { return &PaymentHandler{db} }

func (h *PaymentHandler) GetExchangeRates(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates ORDER BY chain")
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rates"}); return }
	defer rows.Close()
	rates := []models.ExchangeRate{}
	for rows.Next() {
		var r models.ExchangeRate
		if err := rows.Scan(&r.ID, &r.Chain, &r.Symbol, &r.RateToUSD, &r.UpdatedAt); err != nil { continue }
		rates = append(rates, r)
	}
	if rates == nil { rates = []models.ExchangeRate{} }
	c.JSON(http.StatusOK, gin.H{"rates": rates})
}

func (h *PaymentHandler) GetExchangeRate(c *gin.Context) {
	chain := c.Param("chain")
	if chain == "" { c.JSON(http.StatusBadRequest, gin.H{"error": "chain required"}); return }
	var r models.ExchangeRate
	err := h.db.QueryRow("SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates WHERE chain = $1", chain).Scan(&r.ID, &r.Chain, &r.Symbol, &r.RateToUSD, &r.UpdatedAt)
	if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "rate not found"}); return }
	c.JSON(http.StatusOK, gin.H{"rate": r})
}

func (h *PaymentHandler) SetExchangeRate(c *gin.Context) {
	chain := c.Param("chain")
	if chain == "" { c.JSON(http.StatusBadRequest, gin.H{"error": "chain required"}); return }
	var req struct { Symbol string `json:"symbol" binding:"required"`; RateToUSD float64 `json:"rate_to_usd" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	if req.RateToUSD <= 0 { c.JSON(http.StatusBadRequest, gin.H{"error": "rate must be positive"}); return }
	_, err := h.db.Exec("INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()", chain, req.Symbol, req.RateToUSD)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set rate"}); return }
	var rateOut models.ExchangeRate
	h.db.QueryRow("SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates WHERE chain = $1", chain).Scan(&rateOut.ID, &rateOut.Chain, &rateOut.Symbol, &rateOut.RateToUSD, &rateOut.UpdatedAt)
	c.JSON(http.StatusOK, gin.H{"message": "rate updated", "rate": rateOut})
}

func (h *PaymentHandler) DeleteExchangeRate(c *gin.Context) {
	chain := c.Param("chain")
	if chain == "" { c.JSON(http.StatusBadRequest, gin.H{"error": "chain required"}); return }
	result, err := h.db.Exec("DELETE FROM exchange_rates WHERE chain = $1", chain)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"}); return }
	rows, _ := result.RowsAffected()
	if rows == 0 { c.JSON(http.StatusNotFound, gin.H{"error": "rate not found"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "rate deleted"})
}

func (h *PaymentHandler) GetSettings(c *gin.Context) {
	rows, err := h.db.Query("SELECT key, value, updated_at FROM settings ORDER BY key")
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	defer rows.Close()
	settings := []gin.H{}
	for rows.Next() {
		var k, v, u string
		if err := rows.Scan(&k, &v, &u); err != nil { continue }
		settings = append(settings, gin.H{"key": k, "value": v, "updated_at": u})
	}
	if settings == nil { settings = []gin.H{} }
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func (h *PaymentHandler) SetSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" { c.JSON(http.StatusBadRequest, gin.H{"error": "key required"}); return }
	var req struct{ Value string `json:"value"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	_, err := h.db.Exec("INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, NOW()) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()", key, req.Value)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }
	c.JSON(http.StatusOK, gin.H{"key": key, "value": req.Value})
}

func (h *PaymentHandler) GetPaymentDetails(c *gin.Context) {
	idStr := c.Param("orderId")
	var orderID int
	if _, err := strconv.Atoi(idStr); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	orderID, _ = strconv.Atoi(idStr)

	var userID int
	if authHeader := c.GetHeader("Authorization"); authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		if claims, err := auth.ValidateJWT(strings.TrimPrefix(authHeader, "Bearer "), auth.GetJwtSecret()); err == nil { userID = claims.UserID }
	}

	var status, memo, txHash, cryptoChain sql.NullString
	var totalUSD float64
	var createdAt time.Time
	h.db.QueryRow("SELECT status, total_usd, memo, payment_tx_hash, crypto_chain, created_at FROM orders WHERE id = $1", orderID).Scan(&status, &totalUSD, &memo, &txHash, &cryptoChain, &createdAt)
	if status.String == "" { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }

	if userID > 0 {
		var uid int
		h.db.QueryRow("SELECT user_id FROM orders WHERE id = $1", orderID).Scan(&uid)
		if uid != userID { c.JSON(http.StatusForbidden, gin.H{"error": "not your order"}); return }
	}

	var rate float64
	h.db.QueryRow("SELECT rate_to_usd FROM exchange_rates WHERE chain = $1", cryptoChain.String).Scan(&rate)
	if rate == 0 { rate = 1 }
	cryptoAmount := fmt.Sprintf("%.8f", totalUSD/rate)

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID, "status": status.String, "total_usd": totalUSD,
		"crypto_chain": cryptoChain.String, "crypto_amount": cryptoAmount,
		"memo": memo.String, "payment_tx_hash": txHash.String,
		"created_at": createdAt.Format(time.RFC3339),
	})
}

func (h *PaymentHandler) CheckPaymentStatus(c *gin.Context) {
	idStr := c.Param("orderId")
	var orderID int
	if _, err := strconv.Atoi(idStr); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	orderID, _ = strconv.Atoi(idStr)
	var status string; var confs int; var txHash sql.NullString
	h.db.QueryRow("SELECT status, payment_confirmations, payment_tx_hash FROM orders WHERE id = $1", orderID).Scan(&status, &confs, &txHash)
	if status == "" { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }
	c.JSON(http.StatusOK, gin.H{"order_id": orderID, "status": status, "payment_confirmations": confs, "payment_tx_hash": txHash.String})
}

func (h *PaymentHandler) ConfirmPayment(c *gin.Context) {
	idStr := c.Param("orderId")
	var orderID int
	if _, err := strconv.Atoi(idStr); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	orderID, _ = strconv.Atoi(idStr)
	result, err := h.db.Exec("UPDATE orders SET status = 'paid', paid_at = NOW() WHERE id = $1 AND status = 'pending'", orderID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "confirm failed"}); return }
	rows, _ := result.RowsAffected()
	if rows == 0 { c.JSON(http.StatusOK, gin.H{"message": "already paid or not pending"}); return }
	h.db.Exec("UPDATE order_items SET status = 'delivered' WHERE order_id = $1", orderID)
	c.JSON(http.StatusOK, gin.H{"message": "payment confirmed", "order_id": orderID})
}
