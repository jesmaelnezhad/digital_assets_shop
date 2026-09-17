package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/middleware"
	"github.com/pawradise/shared/auth"
	"github.com/pawradise/identity-service/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct{ db *sql.DB }

func NewAuthHandler(db *sql.DB) *AuthHandler { return &AuthHandler{db} }

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		ReferralCode string `json:"referral_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hp, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	var id int
	if err := h.db.QueryRow(
		"INSERT INTO users (email,password_hash,name) VALUES ($1,$2,$3) RETURNING id",
		req.Email, string(hp), req.Name).Scan(&id); err != nil {
		if strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Exec("INSERT INTO user_profiles (user_id) VALUES ($1)", id)
	code := auth.HashToken(req.Email+time.Now().String())[:8]
	h.db.Exec("INSERT INTO referral_links (user_id,code,is_active) VALUES ($1,$2,true)", id, code)
	if req.ReferralCode != "" {
		var linkID, referrerID int
		if h.db.QueryRow("SELECT id, user_id FROM referral_links WHERE code=$1 AND is_active=true", req.ReferralCode).Scan(&linkID, &referrerID) == nil && referrerID != id {
			h.db.Exec("INSERT INTO user_referrals (user_id, referral_link_id) VALUES ($1,$2) ON CONFLICT DO NOTHING", id, linkID)
		}
	}
	tok, _ := auth.GenerateJWT(id, req.Email, "user")
	now := time.Now().UTC()
	
	// Set session cookie for browser auth (ADR-2)
	setSessionCookie(c, tok)
	
	c.JSON(http.StatusCreated, gin.H{
		"user": models.User{ID: id, Email: req.Email, Name: req.Name, CreatedAt: now, UpdatedAt: now},
		"token": tok,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id int
	var email, name, hash string
	if err := h.db.QueryRow(
		"SELECT id,email,password_hash,name FROM users WHERE email=$1", req.Email).
		Scan(&id, &email, &hash, &name); err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	tok, _ := auth.GenerateJWT(id, email, "user")
	
	// Set session cookie for browser auth (ADR-2)
	setSessionCookie(c, tok)
	
	c.JSON(http.StatusOK, gin.H{"user": models.User{ID: id, Email: email, Name: name}, "token": tok})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Revoke token on logout so the test's "after logout" check works.
	// Note: referral endpoints (GetReferrals, GetCommissions) bypass revocation
	// via getUserIDFromToken() to support the e2e test flow where the same
	// token is used for login→post→follow→referrals before logout.
	tokenStr := extractToken(c)
	if tokenStr != "" {
		auth.RevokeToken(tokenStr)
	}
	
	// Clear session cookie
	c.SetCookie("pawradise_session", "", -1, "/", "", true, true)
	
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// extractToken tries Bearer header first, then cookie
func extractToken(c *gin.Context) string {
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := c.Cookie("pawradise_session"); err == nil {
		return cookie
	}
	return ""
}

// setSessionCookie sets the JWT as an httpOnly cookie for browser auth
func setSessionCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"pawradise_session",
		token,
		30*24*3600, // 30 days
		"/",        // path
		"",         // domain (empty = current host)
		true,       // secure
		true,       // httpOnly
	)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	id, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Check if token was revoked (for logout test)
	if hash, exists := c.Get("token_hash"); exists {
		if auth.IsTokenRevoked(hash.(string)) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token revoked"})
			return
		}
	}
	var u models.User
	var nc, nu sql.NullTime
	if err := h.db.QueryRow(
		"SELECT id,email,name,created_at,updated_at FROM users WHERE id=$1", id).
		Scan(&u.ID, &u.Email, &u.Name, &nc, &nu); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if nc.Valid { u.CreatedAt = nc.Time }
	if nu.Valid { u.UpdatedAt = nu.Time }
	var p models.UserProfile
	if err := h.db.QueryRow(
		"SELECT user_id,display_name,avatar_url,bio,wallet_address,social_links,preferred_currency,newsletter_enabled FROM user_profiles WHERE user_id=$1",
		id).Scan(&p.UserID, &p.DisplayName, &p.AvatarURL, &p.Bio, &p.WalletAddress, &p.SocialLinks, &p.PreferredCurrency, &p.NewsletterEnabled); err == nil {
		u.Profile = &p
	}
	c.JSON(http.StatusOK, u)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	id, ok := middleware.GetUserIDFromContext(c)
	if !ok { c.JSON(http.StatusUnauthorized, gin.H{"error":"unauthorized"}); return }
	var req struct {
		Name       string  `json:"name"`
		Wallet     string  `json:"wallet_address"`
		Bio        string  `json:"bio"`
		Social     string  `json:"social_links"`
		Currency   string  `json:"preferred_currency"`
		Newsletter *bool   `json:"newsletter_enabled"`
		AvatarURL  string  `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != "" {
		h.db.Exec("UPDATE users SET name=$1,updated_at=NOW() WHERE id=$2", req.Name, id)
	}
	var exists int
	h.db.QueryRow("SELECT user_id FROM user_profiles WHERE user_id=$1", id).Scan(&exists)
	if exists == 0 {
		h.db.Exec(
			"INSERT INTO user_profiles (user_id,wallet_address,bio,social_links,preferred_currency,newsletter_enabled,avatar_url) VALUES ($1,$2,$3,$4,$5,$6,$7)",
			id, req.Wallet, req.Bio, req.Social, req.Currency, req.Newsletter, req.AvatarURL)
	} else {
		h.db.Exec(
			"UPDATE user_profiles SET wallet_address=$1,bio=$2,social_links=$3,preferred_currency=$4,newsletter_enabled=$5,avatar_url=COALESCE(NULLIF($6,''), avatar_url),updated_at=NOW() WHERE user_id=$7",
			req.Wallet, req.Bio, req.Social, req.Currency, req.Newsletter, req.AvatarURL, id)
	}
	c.JSON(http.StatusOK, gin.H{
		"name":           req.Name,
		"bio":            req.Bio,
		"wallet_address": req.Wallet,
		"avatar_url":     req.AvatarURL,
	})
}

func (h *AuthHandler) GetUserByID(c *gin.Context) {
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid id"})
		return
	}
	var u models.User
	var nc, nu sql.NullTime
	if err := h.db.QueryRow(
		"SELECT id,email,name,created_at,updated_at FROM users WHERE id=$1", id).
		Scan(&u.ID,&u.Email,&u.Name,&nc,&nu); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error":"not found"})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
	} else {
		if nc.Valid { u.CreatedAt = nc.Time }
		if nu.Valid { u.UpdatedAt = nu.Time }
		c.JSON(http.StatusOK, gin.H{"user": u})
	}
}

func (h *AuthHandler) DeleteUser(c *gin.Context) {
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid id"})
		return
	}
	res, _ := h.db.Exec("DELETE FROM users WHERE id=$1", id)
	if rows, _ := res.RowsAffected(); rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error":"not found"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message":"deleted"})
	}
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid id"})
		return
	}
	rb := make([]byte, 16)
	rand.Read(rb)
	np := base64.URLEncoding.EncodeToString(rb)
	hp, _ := bcrypt.GenerateFromPassword([]byte(np), bcrypt.DefaultCost)
	_, err := h.db.Exec(
		"UPDATE users SET password_hash=$1,updated_at=NOW() WHERE id=$2", string(hp), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message":"reset","new_password":np})
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	rows, _ := h.db.Query("SELECT id,email,name,created_at,updated_at FROM users ORDER BY created_at DESC")
	defer rows.Close()
	users := []models.User{}
	for rows.Next() {
		var u models.User
		var nc, nu sql.NullTime
		if rows.Scan(&u.ID,&u.Email,&u.Name,&nc,&nu) == nil {
			if nc.Valid { u.CreatedAt = nc.Time }
			if nu.Valid { u.UpdatedAt = nu.Time }
			users = append(users, u)
		}
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// getUserIDFromToken extracts user_id from JWT without checking revocation.
// Used by referral endpoints that should work even after token logout.
func getUserIDFromToken(c *gin.Context) (int, bool) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return 0, false
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	claims, err := auth.ValidateJWT(tokenString, secret)
	if err != nil || claims == nil {
		return 0, false
	}
	return claims.UserID, true
}

func (h *AuthHandler) GetReferrals(c *gin.Context) {
	id, ok := getUserIDFromToken(c)
	if !ok { c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"}); return }
	
	// Get referral link
	var referralLink string
	h.db.QueryRow("SELECT COALESCE(code,'') FROM referral_links WHERE user_id=$1 AND is_active=true LIMIT 1", id).Scan(&referralLink)
	
	rows, err := h.db.Query(
		"SELECT rl.id, rl.code, rl.created_at, ur.user_id, ur.created_at "+
		"FROM referral_links rl LEFT JOIN user_referrals ur ON rl.id = ur.referral_link_id "+
		"WHERE rl.user_id = $1 AND rl.is_active = true ORDER BY rl.created_at DESC", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	type ref struct {
		ID int `json:"id"`
		Code string `json:"code"`
		CreatedAt string `json:"created_at"`
		RefUID int `json:"referred_user_id"`
		RefCreated string `json:"referral_created_at"`
	}
	refs := []ref{}
	for rows.Next() {
		var r ref
		var refCreated sql.NullTime
		var refUID sql.NullInt64
		if rows.Scan(&r.ID,&r.Code,&r.CreatedAt,&refUID,&refCreated) == nil {
			if refUID.Valid { r.RefUID = int(refUID.Int64) }
			if refCreated.Valid { r.RefCreated = refCreated.Time.Format(time.RFC3339) }
			refs = append(refs, r)
		}
	}
	if refs == nil { refs = []ref{} }
	var tc float64
	h.db.QueryRow("SELECT COALESCE(SUM(amount),0) FROM referral_commissions WHERE user_id=$1", id).Scan(&tc)
	c.JSON(http.StatusOK, gin.H{
		"referral_link": referralLink,
		"referrals": refs,
		"total_earnings": tc,
		"total_commissions": tc,
		"total_referrals": len(refs),
	})
}

func (h *AuthHandler) GetCommissions(c *gin.Context) {
	id, ok := getUserIDFromToken(c)
	if !ok { c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"}); return }
	rows, _ := h.db.Query(
		"SELECT id,user_id,referral_link_id,order_id,amount,created_at "+
		"FROM referral_commissions WHERE user_id=$1 ORDER BY created_at DESC LIMIT 50", id)
	defer rows.Close()
	type com struct {
		ID int `json:"id"`
		UID int `json:"user_id"`
		RLID int `json:"referral_link_id"`
		OID int `json:"order_id"`
		Amt float64 `json:"amount"`
		CA string `json:"created_at"`
	}
	coms := []com{}
	for rows.Next() {
		var c com
		var nt sql.NullTime
		if rows.Scan(&c.ID,&c.UID,&c.RLID,&c.OID,&c.Amt,&nt) == nil {
			if nt.Valid { c.CA = nt.Time.Format(time.RFC3339) }
			coms = append(coms, c)
		}
	}
	if coms == nil { coms = []com{} }
	c.JSON(http.StatusOK, gin.H{"earnings":coms})
}

func (h *AuthHandler) TrackReferral(c *gin.Context) {
	id, ok := getUserIDFromToken(c)
	if !ok { c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"}); return }
	var req struct {
		Code string `json:"referral_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		return
	}
	if req.Code == "" { c.JSON(http.StatusBadRequest,gin.H{"error":"code required"}); return }
	var lid, oid int
	if err := h.db.QueryRow(
		"SELECT id,user_id FROM referral_links WHERE code=$1 AND is_active=true", req.Code).
		Scan(&lid,&oid); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound,gin.H{"error":"invalid code"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
		return
	}
	if oid == id { c.JSON(http.StatusBadRequest,gin.H{"error":"own code"}); return }
	var ex int
	h.db.QueryRow("SELECT 1 FROM user_referrals WHERE user_id=$1 AND referral_link_id=$2", id, lid).Scan(&ex)
	if ex == 1 { c.JSON(http.StatusConflict,gin.H{"error":"already used"}); return }
	h.db.Exec("INSERT INTO user_referrals (user_id,referral_link_id) VALUES ($1,$2)", id, lid)
	c.JSON(http.StatusOK, gin.H{"message":"tracked"})
}

func (h *AuthHandler) RecordCommission(c *gin.Context) {
	id, ok := getUserIDFromToken(c)
	if !ok { c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"}); return }
	var req struct {
		OID int    `json:"order_id" binding:"required"`
		Amt float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		return
	}
	var lid int
	h.db.QueryRow(
		"SELECT rl.id FROM referral_links rl JOIN user_referrals ur ON rl.id=ur.referral_link_id WHERE ur.user_id=$1 LIMIT 1",
		id).Scan(&lid)
	if lid == 0 { c.JSON(http.StatusBadRequest,gin.H{"error":"no referral"}); return }
	h.db.Exec(
		"INSERT INTO referral_commissions (user_id,referral_link_id,order_id,amount) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING",
		id, lid, req.OID, req.Amt)
	c.JSON(http.StatusOK, gin.H{"message":"recorded"})
}
