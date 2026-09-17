package handlers

import (
	"backend/database"
	"backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListExchangeRates returns all exchange rates
func ListExchangeRates(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates ORDER BY chain")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rates"})
		return
	}
	defer rows.Close()

	rates := []models.ExchangeRate{}
	for rows.Next() {
		var r models.ExchangeRate
		if rows.Scan(&r.ID, &r.Chain, &r.Symbol, &r.RateToUSD, &r.UpdatedAt) != nil {
			continue
		}
		rates = append(rates, r)
	}
	c.JSON(http.StatusOK, gin.H{"rates": rates})
}

// GetExchangeRate returns a single exchange rate
func GetExchangeRate(c *gin.Context) {
	chain := c.Param("chain")
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chain is required"})
		return
	}

	var r models.ExchangeRate
	err := database.DB.QueryRow(
		"SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates WHERE chain = $1",
		chain,
	).Scan(&r.ID, &r.Chain, &r.Symbol, &r.RateToUSD, &r.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rate not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rate": r})
}

// SetExchangeRate creates or updates an exchange rate
func SetExchangeRate(c *gin.Context) {
	chain := c.Param("chain")
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chain is required"})
		return
	}

	var req struct {
		Symbol    string `json:"symbol" binding:"required"`
		RateToUSD string `json:"rate_to_usd" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	rate, err := strconv.ParseFloat(req.RateToUSD, 64)
	if err != nil || rate <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rate_to_usd must be a positive number"})
		return
	}

	_, err = database.DB.Exec(
		"INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()",
		chain, req.Symbol, req.RateToUSD,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set rate: " + err.Error()})
		return
	}

	var rateOut models.ExchangeRate
	database.DB.QueryRow(
		"SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates WHERE chain = $1",
		chain,
	).Scan(&rateOut.ID, &rateOut.Chain, &rateOut.Symbol, &rateOut.RateToUSD, &rateOut.UpdatedAt)

	c.JSON(http.StatusOK, gin.H{"message": "rate updated", "rate": rateOut})
}

// DeleteExchangeRate removes an exchange rate
func DeleteExchangeRate(c *gin.Context) {
	chain := c.Param("chain")
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chain is required"})
		return
	}

	result, err := database.DB.Exec("DELETE FROM exchange_rates WHERE chain = $1", chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete rate: " + err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "rate not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rate deleted"})
}

// RegisterExchangeRateRoutes registers exchange rate admin routes under /admin/exchange-rates
// Must be called inside an admin group (with admin auth middleware)
func RegisterExchangeRateRoutes(r *gin.RouterGroup) {
	r.GET("", ListExchangeRates)
	r.GET("/:chain", GetExchangeRate)
	r.PUT("/:chain", SetExchangeRate)
	r.DELETE("/:chain", DeleteExchangeRate)
}
