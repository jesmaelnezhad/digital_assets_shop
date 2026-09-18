package models

import "time"

type User struct {
	ID        int        `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Role      string     `json:"role"`
	StaffTabs string     `json:"staff_tabs,omitempty"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	Profile   *UserProfile `json:"profile,omitempty"`
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserProfile struct {
	UserID            int        `json:"user_id"`
	DisplayName       string     `json:"display_name"`
	AvatarURL         string     `json:"avatar_url"`
	Bio               string     `json:"bio"`
	WalletAddress     string     `json:"wallet_address"`
	SocialLinks       string     `json:"social_links"`
	PreferredCurrency string     `json:"preferred_currency"`
	NewsletterEnabled *bool      `json:"newsletter_enabled"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
}

type ReferralLink struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Code      string    `json:"code"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type UserReferral struct {
	ID              int        `json:"id"`
	UserID          int        `json:"user_id"`
	ReferralLinkID  int        `json:"referral_link_id"`
	ReferredUserID  int        `json:"referred_user_id"`
	ReferralCreated *time.Time `json:"referral_created_at,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}

type ReferralCommission struct {
	ID             int        `json:"id"`
	UserID         int        `json:"user_id"`
	ReferralLinkID int        `json:"referral_link_id"`
	OrderID        int        `json:"order_id"`
	Amount         float64    `json:"amount"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
}
