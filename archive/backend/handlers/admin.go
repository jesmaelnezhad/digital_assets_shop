package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"backend/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminToken := os.Getenv("ADMIN_TOKEN")
		if adminToken == "" {
			adminToken = "admin-secret-token-change-in-production"
		}
		
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin authorization required"})
			c.Abort()
			return
		}
		
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != adminToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid admin token"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, email, name, created_at, updated_at FROM users ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()
	
	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt); err != nil {
			continue
		}
		users = append(users, user)
	}
	
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	
	result, err := h.db.Exec("DELETE FROM users WHERE id = $1", userID)
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

func (h *AuthHandler) ResetPassword(c *gin.Context) {
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
	
	_, err = h.db.Exec("UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", string(hashedPassword), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
		"new_password": newPassword,
	})
}
