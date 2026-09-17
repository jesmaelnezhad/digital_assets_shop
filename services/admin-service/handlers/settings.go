package handlers

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetSettings returns all configuration settings.
func GetSettings(c *gin.Context) {
	rows, err := db.Query("SELECT key, value, updated_at FROM settings ORDER BY key")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}
	defer rows.Close()

	var settings []gin.H
	for rows.Next() {
		var key, value string
		var updatedAt sql.NullTime
		if err := rows.Scan(&key, &value, &updatedAt); err != nil {
			continue
		}
		settings = append(settings, gin.H{
			"key":        key,
			"value":      value,
			"updated_at": nullTimeToString(updatedAt),
		})
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// SetSetting creates or updates a configuration setting.
func SetSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	var req struct {
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value is required"})
		return
	}

	_, err := db.Exec(`
		INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, key, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key, "value": req.Value})
}

// ExportEmails exports all user email addresses.
func ExportEmails(c *gin.Context) {
	rows, err := db.Query("SELECT email FROM users WHERE email IS NOT NULL AND email != '' ORDER BY email")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export emails"})
		return
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			continue
		}
		emails = append(emails, email)
	}

	c.JSON(http.StatusOK, gin.H{
		"emails": emails,
		"total":  len(emails),
	})
}

// GetStats returns dashboard statistics.
func GetStats(c *gin.Context) {
	var stats struct {
		TotalUsers      int
		TotalProducts   int
		TotalOrders     int
		TotalRevenue    string
		PendingOrders   int
		ActiveProducts  int
		NewUsersToday   int
		NewOrdersToday  int
		RevenueToday    string
	}

	err := db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(*) FROM products) AS total_products,
			(SELECT COUNT(*) FROM orders) AS total_orders,
			(SELECT COALESCE(SUM(total_amount), 0) FROM orders) AS total_revenue,
			(SELECT COUNT(*) FROM orders WHERE status = 'pending') AS pending_orders,
			(SELECT COUNT(*) FROM products WHERE status = 'active') AS active_products,
			(SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE) AS new_users_today,
			(SELECT COUNT(*) FROM orders WHERE created_at >= CURRENT_DATE) AS new_orders_today,
			(SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE created_at >= CURRENT_DATE) AS revenue_today
	`).Scan(
		&stats.TotalUsers, &stats.TotalProducts, &stats.TotalOrders,
		&stats.TotalRevenue, &stats.PendingOrders, &stats.ActiveProducts,
		&stats.NewUsersToday, &stats.NewOrdersToday, &stats.RevenueToday,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_users":       stats.TotalUsers,
		"total_products":    stats.TotalProducts,
		"total_orders":      stats.TotalOrders,
		"total_revenue_usd": stats.TotalRevenue,
		"pending_orders":    stats.PendingOrders,
		"active_products":   stats.ActiveProducts,
		"new_users_today":   stats.NewUsersToday,
		"new_orders_today":  stats.NewOrdersToday,
		"revenue_today_usd": stats.RevenueToday,
	})
}

// GetReferrals returns referral statistics.
func GetReferrals(c *gin.Context) {
	rows, err := db.Query(`
		SELECT u.email, COUNT(r.id) AS referral_count,
		       COALESCE(SUM(r.reward_amount), 0) AS total_rewards
		FROM users u
		LEFT JOIN referrals r ON r.referrer_id = u.id
		GROUP BY u.email
		ORDER BY referral_count DESC
		LIMIT 100
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch referrals"})
		return
	}
	defer rows.Close()

	var referrals []gin.H
	for rows.Next() {
		var email string
		var count int
		var rewards string
		if err := rows.Scan(&email, &count, &rewards); err != nil {
			continue
		}
		referrals = append(referrals, gin.H{
			"email":          email,
			"referral_count": count,
			"total_rewards":  rewards,
		})
	}

	c.JSON(http.StatusOK, gin.H{"referrals": referrals})
}

// getSetting reads a key from the settings table with env var fallback.
func getSetting(key, fallback string) string {
	var v string
	if db != nil {
		if err := db.QueryRow("SELECT value FROM settings WHERE key = $1", key).Scan(&v); err == nil && v != "" {
			return v
		}
	}
	if env := os.Getenv("SETTING_" + key); env != "" {
		return env
	}
	return fallback
}
