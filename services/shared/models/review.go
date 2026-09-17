package models

import "time"

// Review represents a product review (text + rating).
type Review struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	ProductID   int       `json:"product_id" db:"product_id"`
	Title       string    `json:"title,omitempty" db:"title"`
	Content     string    `json:"content,omitempty" db:"content"`
	Rating      int       `json:"rating" db:"rating"`
	Verified    bool      `json:"verified" db:"verified"`
	IsVerifiedPurchase bool `json:"is_verified_purchase" db:"is_verified_purchase"`
	IsHelpful   bool      `json:"is_helpful" db:"is_helpful"`
	HelpfulCount int     `json:"helpful_count" db:"helpful_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	UserName    string    `json:"user_name,omitempty" db:"-"`
	UserAvatar  string    `json:"user_avatar,omitempty" db:"-"`
}

// Rating represents a numeric rating on a product.
type Rating struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	ProductID int       `json:"product_id" db:"product_id"`
	Rating    int       `json:"rating" db:"rating"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ProductRatingSummary aggregates rating data for a product.
type ProductRatingSummary struct {
	Average    float64        `json:"average"`
	Count      int            `json:"count"`
	Distribution map[int]int  `json:"distribution"` // rating (1-5) -> count
}

// CreateRatingRequest is the payload for creating a rating.
type CreateRatingRequest struct {
	UserID    int `json:"user_id" binding:"required"`
	ProductID int `json:"product_id" binding:"required"`
	Rating    int `json:"rating" binding:"required,min=1,max=5"`
}

// CreateReviewRequest is the payload for creating a review.
type CreateReviewRequest struct {
	UserID    int    `json:"user_id" binding:"required"`
	ProductID int    `json:"product_id" binding:"required"`
	Title     string `json:"title,omitempty" binding:"max=200"`
	Content   string `json:"content,omitempty" binding:"max=5000"`
	Rating    int    `json:"rating" binding:"required,min=1,max=5"`
}

// CheckVerifiedPurchaseResponse confirms whether a user bought a product.
type CheckVerifiedPurchaseResponse struct {
	Verified bool `json:"verified"`
}
