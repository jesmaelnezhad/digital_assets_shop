package handlers

import (
	"backend/database"
	"database/sql"
	"backend/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

const (
	defaultCryptoChain = "bsc"
)

// paymentAddress returns the seller wallet from settings (admin-configurable).
func paymentAddress() string { return GetPaymentAddress() }

func generateMemo(orderID int) string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("PAW%d-%s", orderID, hex.EncodeToString(bytes))
}

func usdToCrypto(usdAmount float64, rateToUSD float64) string {
	if rateToUSD <= 0 {
		return "0"
	}
	return fmt.Sprintf("%.8f", usdAmount/rateToUSD)
}

func getExchangeRate(chain string) (float64, error) {
	var rateStr string
	err := database.DB.QueryRow(
		"SELECT rate_to_usd FROM exchange_rates WHERE chain = $1", chain,
	).Scan(&rateStr)
	if err != nil {
		return 0, err
	}
	rate, _ := strconv.ParseFloat(rateStr, 64)
	return rate, nil
}

// getPaymentInfo builds a PaymentInfo response for an order
func getPaymentInfo(orderID int, totalUSD float64, rate float64) models.PaymentInfo {
	totalCrypto := usdToCrypto(totalUSD, rate)
	return models.PaymentInfo{
		OrderID:        orderID,
		AmountCrypto:   totalCrypto,
		CryptoChain:    defaultCryptoChain,
		PaymentAddress: paymentAddress(),
		Memo:           generateMemo(orderID),
		Instructions:   fmt.Sprintf("Send %s %s BNB to %s on BSC network. Include memo: %s in transaction data.",
			totalCrypto, defaultCryptoChain, paymentAddress(), generateMemo(orderID)),
	}
}

// CreateOrder creates a new order for the authenticated user
func CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	uid := userID.(int)

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one item required"})
		return
	}

	qtyMap := map[int]int{}
	var productIDs []int
	for _, item := range req.Items {
		if item.ProductID < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}
		productIDs = append(productIDs, item.ProductID)
		q := item.Quantity
		if q < 1 {
			q = 1
		}
		qtyMap[item.ProductID] = q
	}

	rows, err := database.DB.Query(
		"SELECT id, price_usd, status FROM products WHERE id = ANY($1) AND status = 'active'",
		pq.Array(productIDs),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}
	defer rows.Close()

	products := map[int]float64{}
	for rows.Next() {
		var id int
		var priceUSD, status string
		if rows.Scan(&id, &priceUSD, &status) != nil {
			continue
		}
		if status != "active" {
			continue
		}
		price, _ := strconv.ParseFloat(priceUSD, 64)
		products[id] = price
	}

	if len(products) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid active products found"})
		return
	}

	var totalUSD float64
	for pid, qty := range qtyMap {
		if price, ok := products[pid]; ok {
			totalUSD += price * float64(qty)
		}
	}

	rate, err := getExchangeRate(defaultCryptoChain)
	if err != nil || rate <= 0 {
		rate = 350.0
	}

	totalCrypto := usdToCrypto(totalUSD, rate)

	var orderID int
	err = database.DB.QueryRow(`
		INSERT INTO orders (user_id, status, total_usd, total_crypto, crypto_chain, payment_address)
		VALUES ($1, 'pending', $2, $3, $4, $5)
		RETURNING id
	`, uid, fmt.Sprintf("%.2f", totalUSD), totalCrypto, defaultCryptoChain, paymentAddress(),
	).Scan(&orderID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order: " + err.Error()})
		return
	}

	for pid, qty := range qtyMap {
		if price, ok := products[pid]; ok {
			itemTotal := price * float64(qty)
			database.DB.Exec(`
				INSERT INTO order_items (order_id, product_id, quantity, price_usd, price_crypto, max_downloads)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, orderID, pid, qty,
				fmt.Sprintf("%.2f", itemTotal),
				usdToCrypto(itemTotal, rate),
				0,
			)
		}
	}

	order := getOrderWithItems(orderID, uid)
	c.JSON(http.StatusCreated, gin.H{
		"order":   order,
		"payment": getPaymentInfo(orderID, totalUSD, rate),
		"message": "order created. send BNB to the address with memo included",
	})
}

// GetUserOrders returns the authenticated user's order history
func GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	uid := userID.(int)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE user_id = $1", uid).Scan(&total)

	rows, err := database.DB.Query(`
		SELECT o.id, o.user_id, o.status, o.total_usd, o.total_crypto, o.crypto_chain,
		       o.payment_address, o.payment_tx_hash, o.payment_confirmations,
		       o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at,
		       u.email, u.name
		FROM orders o JOIN users u ON o.user_id = u.id
		WHERE o.user_id = $1 ORDER BY o.created_at DESC LIMIT $2 OFFSET $3
	`, uid, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var o models.Order
		var email, name sql.NullString
		if rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalUSD, &o.TotalCrypto,
			&o.CryptoChain, &o.PaymentAddress, &o.PaymentTxHash, &o.PaymentConfirmations,
			&o.PaymentConfirmedAt, &o.PaidAt, &o.CreatedAt, &o.UpdatedAt,
			&email, &name) != nil {
			continue
		}
		if email.Valid {
			o.UserEmail = email.String
		}
		if name.Valid {
			o.UserName = name.String
		}
		o.Items = getOrderItems(o.ID)
		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, models.OrderListResponse{Orders: orders, Total: total})
}

// GetOrder returns a single order (owner only)
func GetOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	uid := userID.(int)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
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
		WHERE o.id = $1 AND o.user_id = $2
	`, id, uid).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalUSD, &o.TotalCrypto,
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

// GetOrderPayment returns payment instructions for an order
func GetOrderPayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	uid := userID.(int)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var o models.Order
	err = database.DB.QueryRow(`
		SELECT id, user_id, status, total_usd, total_crypto, crypto_chain,
		       payment_address, payment_tx_hash, payment_confirmations
		FROM orders WHERE id = $1 AND user_id = $2
	`, id, uid).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalUSD, &o.TotalCrypto,
		&o.CryptoChain, &o.PaymentAddress, &o.PaymentTxHash, &o.PaymentConfirmations)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
		return
	}

	if o.Status == "paid" {
		c.JSON(http.StatusOK, gin.H{
			"order_id":        o.ID,
			"status":          "paid",
			"payment_tx_hash": o.PaymentTxHash,
			"message":         "payment already confirmed",
		})
		return
	}

	rate, _ := getExchangeRate(o.CryptoChain)
	if rate <= 0 {
		rate = 350.0
	}

	c.JSON(http.StatusOK, models.PaymentInfo{
		OrderID:        o.ID,
		AmountCrypto:   o.TotalCrypto,
		CryptoChain:    o.CryptoChain,
		PaymentAddress: o.PaymentAddress,
		Memo:           generateMemo(o.ID),
		Instructions:   fmt.Sprintf("Send %s %s BNB to %s on BSC network. Include memo: %s in transaction data.",
			o.TotalCrypto, o.CryptoChain, o.PaymentAddress, generateMemo(o.ID)),
	})
}

// CheckPaymentStatus returns the current payment status for an order
func CheckPaymentStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	uid := userID.(int)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var o models.Order
	err = database.DB.QueryRow(`
		SELECT id, user_id, status, total_crypto, crypto_chain,
		       payment_address, payment_tx_hash, payment_confirmations
		FROM orders WHERE id = $1 AND user_id = $2
	`, id, uid).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalCrypto, &o.CryptoChain,
		&o.PaymentAddress, &o.PaymentTxHash, &o.PaymentConfirmations)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id":             o.ID,
		"status":               o.Status,
		"payment_tx_hash":      o.PaymentTxHash,
		"payment_confirmations": o.PaymentConfirmations,
		"amount_crypto":        o.TotalCrypto,
		"crypto_chain":         o.CryptoChain,
	})
}

// GetOrderDownload returns a download link for an order item (after payment confirmed)
func GetOrderDownload(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	uid := userID.(int)

	idStr := c.Param("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	var oi struct {
		ID            int
		OrderID       int
		ProductID     int
		DownloadCount int
		MaxDownloads  int
		AssetPath     sql.NullString
		ProductTitle  string
	}
	var orderStatus string

	err = database.DB.QueryRow(`
		SELECT oi.id, oi.order_id, oi.product_id, oi.download_count, oi.max_downloads,
		       p.asset_path, p.title, o.status
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN products p ON oi.product_id = p.id
		WHERE oi.id = $1 AND o.user_id = $2
	`, itemID, uid).Scan(&oi.ID, &oi.OrderID, &oi.ProductID, &oi.DownloadCount,
		&oi.MaxDownloads, &oi.AssetPath, &oi.ProductTitle, &orderStatus)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "download not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch item"})
		return
	}

	if orderStatus != "paid" {
		c.JSON(http.StatusForbidden, gin.H{"error": "payment not confirmed yet"})
		return
	}

	if oi.MaxDownloads > 0 && oi.DownloadCount >= oi.MaxDownloads {
		c.JSON(http.StatusForbidden, gin.H{"error": "download limit reached"})
		return
	}

	database.DB.Exec("UPDATE order_items SET download_count = download_count + 1 WHERE id = $1", itemID)

	c.JSON(http.StatusOK, gin.H{
		"item_id":        oi.ID,
		"product_title":  oi.ProductTitle,
		"download_count": oi.DownloadCount + 1,
		"max_downloads":  oi.MaxDownloads,
		"asset_path":     oi.AssetPath.String,
		"message":        "download ready",
	})
}

// getOrderWithItems fetches an order with all its items
func getOrderWithItems(orderID, userID int) models.Order {
	var o models.Order
	database.DB.QueryRow(`
		SELECT o.id, o.user_id, o.status, o.total_usd, o.total_crypto, o.crypto_chain,
		       o.payment_address, o.payment_tx_hash, o.payment_confirmations,
		       o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at,
		       u.email, u.name
		FROM orders o JOIN users u ON o.user_id = u.id
		WHERE o.id = $1 AND o.user_id = $2
	`, orderID, userID).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalUSD, &o.TotalCrypto,
		&o.CryptoChain, &o.PaymentAddress, &o.PaymentTxHash, &o.PaymentConfirmations,
		&o.PaymentConfirmedAt, &o.PaidAt, &o.CreatedAt, &o.UpdatedAt,
		&o.UserEmail, &o.UserName)
	o.Items = getOrderItems(o.ID)
	return o
}

// getOrderItems fetches all items for an order
func getOrderItems(orderID int) []models.OrderItem {
	rows, err := database.DB.Query(`
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price_usd, oi.price_crypto,
		       oi.download_count, oi.max_downloads, oi.downloaded_at, oi.created_at,
		       p.title, p.slug
		FROM order_items oi JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
	`, orderID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	items := []models.OrderItem{}
	for rows.Next() {
		var item models.OrderItem
		var title, slug string
		if rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity,
			&item.PriceUSD, &item.PriceCrypto, &item.DownloadCount, &item.MaxDownloads,
			&item.DownloadedAt, &item.CreatedAt, &title, &slug) != nil {
			continue
		}
		item.Product = &models.Product{Title: title, Slug: slug}
		items = append(items, item)
	}
	return items
}
