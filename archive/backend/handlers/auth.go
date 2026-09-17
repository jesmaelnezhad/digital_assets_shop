package handlers

import (
	"backend/database"
	"backend/models"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	if db == nil {
		db = database.DB
	}
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	var userID int
	err = h.db.QueryRow(
		"INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id",
		req.Email, string(hashedPassword), req.Name,
	).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := generateJWT(userID, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User: models.User{
			ID:    userID,
			Email: req.Email,
			Name:  req.Name,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.db.QueryRow(
		"SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := generateJWT(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	err := h.db.QueryRow(
		"SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Include profile fields (bio, wallet_address) if they exist
	var bio, walletAddress sql.NullString
	h.db.QueryRow("SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1", userID).Scan(&bio, &walletAddress)
	if bio.Valid {
		user.Bio = bio.String
	}
	if walletAddress.Valid {
		user.WalletAddress = walletAddress.String
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Name       string `json:"name"`
		Email      string `json:"email"`
		Bio        string `json:"bio"`
		WalletAddr string `json:"wallet_address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update users table (name, email)
	userUpdates := ""
	userArgs := []interface{}{}
	userArgIndex := 1

	if req.Name != "" {
		userUpdates += fmt.Sprintf("name = $%d", userArgIndex)
		userArgs = append(userArgs, req.Name)
		userArgIndex++
	}
	if req.Email != "" {
		if userUpdates != "" {
			userUpdates += ", "
		}
		userUpdates += fmt.Sprintf("email = $%d", userArgIndex)
		userArgs = append(userArgs, req.Email)
		userArgIndex++
	}

	if userUpdates != "" {
		userUpdates += fmt.Sprintf(", updated_at = CURRENT_TIMESTAMP WHERE id = $%d", userArgIndex)
		userArgs = append(userArgs, userID)

		query := "UPDATE users SET " + userUpdates
		_, err := h.db.Exec(query, userArgs...)
		if err != nil {
			if strings.Contains(err.Error(), "unique constraint") {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}
	}

	// Update user_profiles table (bio, wallet_address)
	_, err := h.db.Exec(
		`INSERT INTO user_profiles (user_id, bio, wallet_address)
		 VALUES ($1, COALESCE($2, ''), COALESCE($3, ''))
		 ON CONFLICT (user_id) DO UPDATE SET
			bio = COALESCE(EXCLUDED.bio, user_profiles.bio),
			wallet_address = COALESCE(EXCLUDED.wallet_address, user_profiles.wallet_address)`,
		userID, req.Bio, req.WalletAddr,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Return updated profile with profile fields
	var user models.User
	err = h.db.QueryRow(
		"SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated profile"})
		return
	}

	var bio, walletAddress sql.NullString
	h.db.QueryRow("SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1", userID).Scan(&bio, &walletAddress)
	if bio.Valid {
		user.Bio = bio.String
	}
	if walletAddress.Valid {
		user.WalletAddress = walletAddress.String
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Server-side invalidation: hash the bearer token into the blocklist so it
	// cannot be used again. The frontend must also drop its stored copy.
	tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no token provided"})
		return
	}
	expiresAt := time.Now().Add(time.Hour * 24) // fallback = JWT lifetime
	if token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{}); err == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				expiresAt = time.Unix(int64(exp), 0)
			}
		}
	}
	if h.db != nil {
		// opportunistic cleanup of expired entries
		h.db.Exec(`DELETE FROM invalidated_tokens WHERE expires_at < NOW()`)
		if _, err := h.db.Exec(`INSERT INTO invalidated_tokens (token_hash, expires_at) VALUES ($1, $2)
			ON CONFLICT (token_hash) DO NOTHING`, hashToken(tokenString), expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func generateJWT(userID int, email string) (string, error) {
	secret := jwtSecret()

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func jwtSecret() string {
	if value := os.Getenv("JWT_SECRET"); value != "" {
		return value
	}
	return "your-secret-key-change-in-production"
}
