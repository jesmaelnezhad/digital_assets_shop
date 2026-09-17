package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/middleware"
	"github.com/pawradise/commerce-service/models"
)

type CommerceHandler struct{ db *sql.DB }

func NewCommerceHandler(db *sql.DB) *CommerceHandler { return &CommerceHandler{db} }

func (h *CommerceHandler) CreateGuestOrder(c *gin.Context) {
	var req struct {
		Email       string  `json:"email" binding:"required,email"`
		TotalUSD    float64 `json:"total_usd"`
		CryptoChain string  `json:"crypto_chain"`
		CryptoAmount string `json:"crypto_amount"`
		CryptoAddress string `json:"crypto_address"`
		Status      string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "" { req.Status = "pending" }

	var orderID int
	err := h.db.QueryRow(
		`INSERT INTO orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		req.Email, req.TotalUSD, req.CryptoChain, req.CryptoAmount, req.CryptoAddress, req.Status,
	).Scan(&orderID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"}); return }

	c.JSON(http.StatusCreated, gin.H{"id": orderID, "message": "guest order created"})
}

func (h *CommerceHandler) GetGuestOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	email := c.Query("email")
	
	var orderID int
	var orderEmail string
	var totalUSD float64
	var status string
	var createdAt time.Time
	err = h.db.QueryRow("SELECT id, email, total_usd, status, created_at FROM orders WHERE id = $1 AND email = $2", id, email).Scan(&orderID, &orderEmail, &totalUSD, &status, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": orderID, "email": orderEmail, "total_usd": totalUSD, "status": status, "created_at": createdAt})
}

// GetGuestOrders lists all guest orders (public, for e2e tests)
func (h *CommerceHandler) GetGuestOrders(c *gin.Context) {
	email := c.Query("email")
	rows, err := h.db.Query("SELECT id, email, total_usd, status, created_at FROM orders WHERE email = $1 ORDER BY created_at DESC", email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	defer rows.Close()
	type guestOrder struct {
		ID        int     `json:"id"`
		Email     string  `json:"email"`
		TotalUSD  float64 `json:"total_usd"`
		Status    string  `json:"status"`
		CreatedAt string  `json:"created_at"`
	}
	orders := []guestOrder{}
	for rows.Next() {
		var o guestOrder
		var t time.Time
		if rows.Scan(&o.ID, &o.Email, &o.TotalUSD, &o.Status, &t) == nil {
			o.CreatedAt = t.Format(time.RFC3339)
			orders = append(orders, o)
		}
	}
	if orders == nil { orders = []guestOrder{} }
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *CommerceHandler) ValidateCoupon(c *gin.Context) {
	var req models.ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var coupon models.Coupon
	if err := h.db.QueryRow(
		"SELECT id, code, discount_type, discount_value, min_purchase, max_uses, current_uses, expires_at, active FROM coupons WHERE code = $1",
		req.Code,
	).Scan(&coupon.ID, &coupon.Code, &coupon.DiscountType, &coupon.DiscountValue, &coupon.MinPurchase, &coupon.MaxUses, &coupon.CurrentUses, &coupon.ExpiresAt, &coupon.Active); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "coupon not found"})
		return
	}
	if !coupon.Active {
		c.JSON(http.StatusBadRequest, gin.H{"error": "coupon is inactive"})
		return
	}
	if coupon.ExpiresAt != nil && coupon.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "coupon expired"})
		return
	}
	if coupon.MinPurchase != "" {
		min, err := strconv.ParseFloat(coupon.MinPurchase, 64)
		if err == nil && req.CartTotal > 0 && req.CartTotal < min {
			c.JSON(http.StatusBadRequest, gin.H{"error": "minimum purchase not met"})
			return
		}
	}
	if coupon.MaxUses > 0 && coupon.CurrentUses >= coupon.MaxUses {
		c.JSON(http.StatusBadRequest, gin.H{"error": "coupon usage limit reached"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"valid":          true,
		"discount_type":  coupon.DiscountType,
		"discount_value": coupon.DiscountValue,
	})
}

func (h *CommerceHandler) CreateOrder(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req models.CreateOrderRequest
	// Parse body first
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// If no items in request, pull from cart
	if len(req.Items) == 0 {
		rows, _ := h.db.Query("SELECT product_id, quantity FROM cart_items WHERE user_id = $1", userID)
		defer rows.Close()
		for rows.Next() {
			var productID, quantity int
			if rows.Scan(&productID, &quantity) == nil {
				req.Items = append(req.Items, models.OrderItemRequest{ProductID: productID, Quantity: quantity})
			}
		}
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no items in cart or request"})
		return
	}
	tx, _ := h.db.Begin()
	var orderID int
	err := tx.QueryRow(
		"INSERT INTO orders (user_id, total_usd, status) VALUES ($1, 0, 'pending') RETURNING id",
		userID,
	).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	var totalUSD float64
	for _, item := range req.Items {
		var price float64
		tx.QueryRow("SELECT price_usd FROM products WHERE id = $1", item.ProductID).Scan(&price)
		tx.Exec(
			"INSERT INTO order_items (order_id, product_id, quantity, price_usd) VALUES ($1, $2, $3, $4)",
			orderID, item.ProductID, item.Quantity, price,
		)
		totalUSD += price * float64(item.Quantity)
	}
	tx.Exec("UPDATE orders SET total_usd = $1 WHERE id = $2", totalUSD, orderID)
	tx.Commit()

	if req.CouponCode != "" {
		var coupon models.Coupon
		if err := h.db.QueryRow("SELECT id, discount_type, discount_value FROM coupons WHERE code = $1 AND active = true", req.CouponCode).Scan(&coupon.ID, &coupon.DiscountType, &coupon.DiscountValue); err == nil {
			h.db.Exec("INSERT INTO coupon_usages (coupon_id, order_id, user_id) VALUES ($1, $2, $3)", coupon.ID, orderID, userID)
		}
	}

	// Store payment info
	walletAddress := "0xPAWRADISE_WALLET_BSC"
	h.db.Exec("UPDATE orders SET payment_address = $1, memo = $2 WHERE id = $3", walletAddress, orderID, orderID)

	c.JSON(http.StatusCreated, gin.H{
		"order":          gin.H{"id": orderID},
		"total_usd":      totalUSD,
		"payment_address": walletAddress,
		"total_crypto":   0,
		"crypto_chain":   "BSC",
		"memo":           orderID,
	})
}

func (h *CommerceHandler) GetUserOrders(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, _ := h.db.Query("SELECT id, total_usd, status, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC", userID)
	defer rows.Close()
	type order struct {
		ID        int     `json:"id"`
		TotalUSD  float64 `json:"total_usd"`
		Status    string  `json:"status"`
		CreatedAt string  `json:"created_at"`
	}
	orders := []order{}
	for rows.Next() {
		var o order
		var t time.Time
		if rows.Scan(&o.ID, &o.TotalUSD, &o.Status, &t) == nil {
			o.CreatedAt = t.Format(time.RFC3339)
			orders = append(orders, o)
		}
	}
	if orders == nil { orders = []order{} }
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *CommerceHandler) GetOrder(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var orderUserID int
	var o models.Order
	var paidAt sql.NullTime
	if err := h.db.QueryRow("SELECT id, user_id, total_usd, status, memo, payment_tx_hash, created_at, paid_at FROM orders WHERE id = $1", id).Scan(&o.ID, &orderUserID, &o.TotalUSD, &o.Status, &o.Memo, &o.PaymentTxHash, &o.CreatedAt, &paidAt); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if orderUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your order"})
		return
	}
	if paidAt.Valid { o.PaidAt = &paidAt.Time }
	items := h.orderItems(id)
	c.JSON(http.StatusOK, gin.H{"order": o, "items": items})
}

func (h *CommerceHandler) GetOrderPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var status string
	var totalUSD float64
	if err := h.db.QueryRow("SELECT status, total_usd FROM orders WHERE id = $1", id).Scan(&status, &totalUSD); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order_id": id, "status": status, "total_usd": totalUSD})
}

func (h *CommerceHandler) CheckPaymentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var status string
	var txHash sql.NullString
	var confirmations int
	if err := h.db.QueryRow("SELECT status, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1", id).Scan(&status, &txHash, &confirmations); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order_id": id, "status": status, "payment_tx_hash": txHash.String, "confirmations": confirmations})
}

func (h *CommerceHandler) GetOrderDownload(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "download endpoint"})
}

func (h *CommerceHandler) ConfirmPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	h.db.Exec("UPDATE orders SET status = 'paid', paid_at = NOW() WHERE id = $1", id)
	c.JSON(http.StatusOK, gin.H{"message": "payment confirmed"})
}

func (h *CommerceHandler) GetCart(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, _ := h.db.Query("SELECT id, product_id, quantity FROM cart_items WHERE user_id = $1", userID)
	defer rows.Close()
	type cartItem struct {
		ID        int `json:"id"`
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}
	items := []cartItem{}
	for rows.Next() {
		var ci cartItem
		if rows.Scan(&ci.ID, &ci.ProductID, &ci.Quantity) == nil {
			items = append(items, ci)
		}
	}
	if items == nil { items = []cartItem{} }
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *CommerceHandler) AddCartItem(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(
		"INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (user_id, product_id) DO UPDATE SET quantity = cart_items.quantity + $3",
		userID, req.ProductID, req.Quantity,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "added to cart"})
}

func (h *CommerceHandler) RemoveCartItem(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	h.db.Exec("DELETE FROM cart_items WHERE id = $1 AND user_id = $2", id, userID)
	c.JSON(http.StatusOK, gin.H{"message": "removed from cart"})
}

func (h *CommerceHandler) GetWishlist(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, _ := h.db.Query("SELECT id, product_id FROM wishlist_items WHERE user_id = $1", userID)
	defer rows.Close()
	type wishItem struct {
		ID        int `json:"id"`
		ProductID int `json:"product_id"`
	}
	items := []wishItem{}
	for rows.Next() {
		var wi wishItem
		if rows.Scan(&wi.ID, &wi.ProductID) == nil {
			items = append(items, wi)
		}
	}
	if items == nil { items = []wishItem{} }
	c.JSON(http.StatusOK, gin.H{"products": items})
}

func (h *CommerceHandler) ToggleWishlist(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	productID := productIDFromRequest(c)
	if productID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id required"})
		return
	}
	var existing int
	h.db.QueryRow("SELECT 1 FROM wishlist_items WHERE user_id = $1 AND product_id = $2", userID, productID).Scan(&existing)
	if existing == 1 {
		h.db.Exec("DELETE FROM wishlist_items WHERE user_id = $1 AND product_id = $2", userID, productID)
		c.JSON(http.StatusOK, gin.H{"message": "removed from wishlist", "added": false})
	} else {
		h.db.Exec("INSERT INTO wishlist_items (user_id, product_id) VALUES ($1, $2)", userID, productID)
		c.JSON(http.StatusOK, gin.H{"message": "added to wishlist", "added": true})
	}
}

func (h *CommerceHandler) RecordView(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	productID := productIDFromRequest(c)
	if productID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id required"})
		return
	}
	h.db.Exec("INSERT INTO recently_viewed (user_id, product_id) VALUES ($1, $2)", userID, productID)
	c.JSON(http.StatusOK, gin.H{"message": "view recorded"})
}

func (h *CommerceHandler) GetRecentlyViewed(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, _ := h.db.Query("SELECT id, product_id, viewed_at FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 20", userID)
	defer rows.Close()
	type rv struct {
		ID        int    `json:"id"`
		ProductID int    `json:"product_id"`
		ViewedAt  string `json:"viewed_at"`
	}
	items := []rv{}
	for rows.Next() {
		var r rv
		var t time.Time
		if rows.Scan(&r.ID, &r.ProductID, &t) == nil {
			r.ViewedAt = t.Format(time.RFC3339)
			items = append(items, r)
		}
	}
	if items == nil { items = []rv{} }
	c.JSON(http.StatusOK, gin.H{"products": items})
}

func (h *CommerceHandler) ToggleCompare(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	productID := productIDFromRequest(c)
	if productID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id required"})
		return
	}
	var existing int
	h.db.QueryRow("SELECT 1 FROM product_comparisons WHERE user_id = $1 AND product_id = $2", userID, productID).Scan(&existing)
	if existing == 1 {
		h.db.Exec("DELETE FROM product_comparisons WHERE user_id = $1 AND product_id = $2", userID, productID)
		c.JSON(http.StatusOK, gin.H{"message": "removed from compare"})
	} else {
		h.db.Exec("INSERT INTO product_comparisons (user_id, product_id) VALUES ($1, $2)", userID, productID)
		c.JSON(http.StatusOK, gin.H{"message": "added to compare"})
	}
}

func (h *CommerceHandler) GetCompare(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, _ := h.db.Query("SELECT id, product_id FROM product_comparisons WHERE user_id = $1", userID)
	defer rows.Close()
	type pc struct {
		ID        int `json:"id"`
		ProductID int `json:"product_id"`
	}
	items := []pc{}
	for rows.Next() {
		var p pc
		if rows.Scan(&p.ID, &p.ProductID) == nil {
			items = append(items, p)
		}
	}
	if items == nil { items = []pc{} }
	c.JSON(http.StatusOK, gin.H{"products": items})
}

func (h *CommerceHandler) CreateCoupon(c *gin.Context) {
	var req struct {
		Code          string  `json:"code" binding:"required"`
		DiscountType  string  `json:"discount_type" binding:"required"`
		DiscountValue float64 `json:"discount_value" binding:"required"`
		MaxUses       int     `json:"max_uses"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(
		"INSERT INTO coupons (code, discount_type, discount_value, max_uses, active) VALUES ($1, $2, $3, $4, true)",
		req.Code, req.DiscountType, req.DiscountValue, req.MaxUses,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "coupon created"})
}

func (h *CommerceHandler) UpdateCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var req struct {
		DiscountValue float64 `json:"discount_value"`
		MaxUses       int     `json:"max_uses"`
		Active        bool    `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.db.Exec("UPDATE coupons SET discount_value = $1, max_uses = $2, active = $3 WHERE id = $4", req.DiscountValue, req.MaxUses, req.Active, id)
	c.JSON(http.StatusOK, gin.H{"message": "coupon updated"})
}

func (h *CommerceHandler) DeleteCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	h.db.Exec("DELETE FROM coupons WHERE id = $1", id)
	c.JSON(http.StatusOK, gin.H{"message": "coupon deleted"})
}
