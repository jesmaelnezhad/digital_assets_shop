package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	adminmodels "github.com/pawradise/admin-service/models"
)

func couponNum(v interface{}) string {
	switch t := v.(type) {
	case float64:
		if t == float64(int(t)) {
			return strconv.Itoa(int(t))
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case string:
		return t
	default:
		return fmt.Sprint(v)
	}
}

func (h *AdminHandler) ListCoupons(c *gin.Context) {
	db := h.ordersDB()
	rows, err := db.Query(`SELECT id, code, discount_type, discount_value, expires_at,
		COALESCE(max_uses,0), COALESCE(current_uses,0), COALESCE(min_purchase,0),
		COALESCE(active,true), created_at, COALESCE(updated_at, created_at)
		FROM coupons ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	defer rows.Close()
	coupons := []adminmodels.Coupon{}
	for rows.Next() {
		var cp adminmodels.Coupon
		var expires sql.NullTime
		var discount, minPurchase float64
		if err := rows.Scan(&cp.ID, &cp.Code, &cp.DiscountType, &discount, &expires,
			&cp.UsageLimit, &cp.TimesUsed, &minPurchase, &cp.IsActive, &cp.CreatedAt, &cp.UpdatedAt); err != nil {
			continue
		}
		cp.DiscountValue = couponNum(discount)
		cp.MinPurchaseUSD = couponNum(minPurchase)
		if expires.Valid {
			cp.ExpiresAt = &expires.Time
		}
		coupons = append(coupons, cp)
	}
	if coupons == nil {
		coupons = []adminmodels.Coupon{}
	}
	c.JSON(http.StatusOK, gin.H{"coupons": coupons})
}

func (h *AdminHandler) CreateCoupon(c *gin.Context) {
	var req struct {
		Code          string  `json:"code" binding:"required"`
		DiscountType  string  `json:"discount_type"`
		DiscountValue float64 `json:"discount_value"`
		ExpiresAt     string  `json:"expires_at"`
		UsageLimit    int     `json:"usage_limit"`
		MinPurchase   float64 `json:"min_purchase_usd"`
		ProductID     *int    `json:"product_id"`
		IsActive      *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	if req.DiscountType == "" {
		req.DiscountType = "percentage"
	}
	if req.UsageLimit == 0 {
		req.UsageLimit = 100
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	var expires sql.NullTime
	if req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err == nil {
			expires = sql.NullTime{Time: t, Valid: true}
		}
	}
	var id int
	err := h.ordersDB().QueryRow(
		`INSERT INTO coupons (code, discount_type, discount_value, expires_at, max_uses, current_uses, min_purchase, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, $6, $7, NOW(), NOW()) RETURNING id`,
		req.Code, req.DiscountType, req.DiscountValue, expires, req.UsageLimit, req.MinPurchase, isActive,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "code": req.Code, "message": "coupon created"})
}

func (h *AdminHandler) UpdateCoupon(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Code          string  `json:"code"`
		DiscountType  string  `json:"discount_type"`
		DiscountValue float64 `json:"discount_value"`
		ExpiresAt     string  `json:"expires_at"`
		UsageLimit    int     `json:"usage_limit"`
		MinPurchase   float64 `json:"min_purchase_usd"`
		IsActive      *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := `UPDATE coupons SET updated_at = NOW()`
	args := []interface{}{}
	idx := 2
	if req.Code != "" {
		query += fmt.Sprintf(", code = $%d", idx)
		args = append(args, strings.ToUpper(strings.TrimSpace(req.Code)))
		idx++
	}
	if req.DiscountType != "" {
		query += fmt.Sprintf(", discount_type = $%d", idx)
		args = append(args, req.DiscountType)
		idx++
	}
	if req.DiscountValue > 0 {
		query += fmt.Sprintf(", discount_value = $%d", idx)
		args = append(args, req.DiscountValue)
		idx++
	}
	if req.ExpiresAt != "" {
		if t, err := time.Parse("2006-01-02", req.ExpiresAt); err == nil {
			query += fmt.Sprintf(", expires_at = $%d", idx)
			args = append(args, t)
			idx++
		}
	}
	if req.UsageLimit > 0 {
		query += fmt.Sprintf(", max_uses = $%d", idx)
		args = append(args, req.UsageLimit)
		idx++
	}
	if req.MinPurchase > 0 {
		query += fmt.Sprintf(", min_purchase = $%d", idx)
		args = append(args, req.MinPurchase)
		idx++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf(", active = $%d", idx)
		args = append(args, *req.IsActive)
		idx++
	}
	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	args = append(args, id)
	if _, err = h.ordersDB().Exec(query+" WHERE id = $1", args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "coupon updated"})
}

func (h *AdminHandler) DeleteCoupon(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.ordersDB().Exec("DELETE FROM coupons WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "coupon deleted"})
}
