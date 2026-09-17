package handlers

import (
	"backend/database"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// getSetting reads a key from the settings table.
// Precedence: settings table -> environment variable (uppercased key) -> fallback.
func getSetting(key, fallback string) string {
	var v string
	if database.DB != nil {
		if err := database.DB.QueryRow(`SELECT value FROM settings WHERE key = $1`, key).Scan(&v); err == nil && v != "" {
			return v
		}
	}
	if env := os.Getenv("SETTING_" + key); env != "" {
		return env
	}
	return fallback
}

// GetPaymentAddress returns the seller wallet buyers must pay to.
// Admin can change it via PUT /admin/settings/payment_address.
func GetPaymentAddress() string {
	return getSetting("payment_address", "0xYourBSCWelcomeAddressHere")
}

// GetSetting returns a single setting value by key (public).
func GetSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}
	val := getSetting(key, "")
	if val == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": key, "value": val})
}

// ListSettings returns all settings (admin only).
func ListSettings(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT key, value, updated_at FROM settings ORDER BY key`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list settings"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var k, v string
		var updated string
		if err := rows.Scan(&k, &v, &updated); err != nil {
			continue
		}
		out = append(out, gin.H{"key": k, "value": v, "updated_at": updated})
	}
	c.JSON(http.StatusOK, gin.H{"settings": out})
}

// SetSetting creates or updates a setting (admin only).
func SetSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}
	var req struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value is required"})
		return
	}
	if _, err := database.DB.Exec(`INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`, key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": key, "value": req.Value})
}
