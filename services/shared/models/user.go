package models

import "time"

// User represents a registered user in the system.
type User struct {
	ID             int       `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	Name           string    `json:"name" db:"name"`
	Bio            string    `json:"bio" db:"bio"`
	AvatarURL      string    `json:"avatar_url,omitempty" db:"avatar_url"`
	WalletAddress  string    `json:"wallet_address,omitempty" db:"wallet_address"`
	ReferralCode   string    `json:"referral_code,omitempty" db:"referral_code"`
	ReferredBy     *int      `json:"referred_by,omitempty" db:"referred_by"`
	Role           string    `json:"role" db:"role"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfile extends User with profile-only fields.
type UserProfile struct {
	ID            int    `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Bio           string `json:"bio,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	WalletAddress string `json:"wallet_address,omitempty"`
	PostCount     int    `json:"post_count"`
	FollowerCount int    `json:"follower_count"`
	FollowingCount int   `json:"following_count"`
}

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=128"`
	Name        string `json:"name" binding:"required,min=1,max=100"`
	ReferralCode string `json:"referral_code,omitempty"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is returned after successful authentication.
type AuthResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
	User      User   `json:"user"`
}

// UpdateProfileRequest is the payload for updating a user profile.
type UpdateProfileRequest struct {
	Name          *string `json:"name,omitempty"`
	Bio           *string `json:"bio,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	WalletAddress *string `json:"wallet_address,omitempty"`
}

// LogoutRequest is the payload for logout.
type LogoutRequest struct {
	Token string `json:"token" binding:"required"`
}

// ValidateTokenRequest is the payload for token validation.
type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// ValidateTokenResponse is returned from token validation.
type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

// ResetPasswordRequest is the payload for admin password reset.
type ResetPasswordRequest struct {
	UserID int `json:"user_id" binding:"required"`
}

// ResetPasswordResponse is returned after password reset.
type ResetPasswordResponse struct {
	NewPassword string `json:"new_password"`
}

// DeleteUserRequest is the payload for user deletion.
type DeleteUserRequest struct {
	UserID int `json:"user_id" binding:"required"`
}

// ReferralStats tracks referral program metrics.
type ReferralStats struct {
	ReferralLink  string  `json:"referral_link"`
	TotalEarnings float64 `json:"total_earnings"`
	ReferredCount int     `json:"referred_count"`
}

// RecordReferralRequest is the payload for recording a referral.
type RecordReferralRequest struct {
	ReferrerID int `json:"referrer_id" binding:"required"`
	ReferredID int `json:"referred_id" binding:"required"`
}

// RecordCommissionRequest is the payload for recording a commission.
type RecordCommissionRequest struct {
	OrderID           int     `json:"order_id" binding:"required"`
	ReferrerID        int     `json:"referrer_id" binding:"required"`
	CommissionPercent float64 `json:"commission_percent" binding:"required"`
	CommissionUSD     float64 `json:"commission_usd" binding:"required"`
}
