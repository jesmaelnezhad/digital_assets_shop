package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ListUsers returns a list of all users for admin review.
func ListUsers(c *gin.Context) {
	rows, err := db.Query(`
		SELECT u.id, u.email, u.name, u.created_at, u.updated_at,
		       COUNT(DISTINCT o.id) AS order_count,
		       COALESCE(SUM(o.total_amount), 0) AS total_spent
		FROM users u
		LEFT JOIN orders o ON o.user_id = u.id
		GROUP BY u.id
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	type userRow struct {
		ID         int
		Email      string
		Name       string
		CreatedAt  sql.NullTime
		UpdatedAt  sql.NullTime
		OrderCount int
		TotalSpent string
	}

	var users []gin.H
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt, &u.OrderCount, &u.TotalSpent); err != nil {
			continue
		}
		users = append(users, gin.H{
			"id":              u.ID,
			"email":           u.Email,
			"name":            u.Name,
			"created_at":      nullTimeToString(u.CreatedAt),
			"updated_at":      nullTimeToString(u.UpdatedAt),
			"order_count":     u.OrderCount,
			"total_spent_usd": u.TotalSpent,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// DeleteUser removes a user account.
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	result, err := db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ResetPassword generates a new random password for a user and returns it.
func ResetPassword(c *gin.Context) {
	userID := c.Param("id")

	// Generate random password
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate password"})
		return
	}
	newPassword := base64.URLEncoding.EncodeToString(randomBytes)

	// Hash and update
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	result, err := db.Exec(
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2",
		string(hashedPassword), userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Password reset successfully",
		"new_password": newPassword,
	})
}
