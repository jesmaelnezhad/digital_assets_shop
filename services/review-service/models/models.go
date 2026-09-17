package models

import "time"

// Review represents a product rating/review by a user.
type Review struct {
	ID          int       `json:"id" db:"id"`
	ProductID   int       `json:"product_id" db:"product_id"`
	UserID      int       `json:"user_id" db:"user_id"`
	Rating      int       `json:"rating" db:"rating"`
	Title       string    `json:"title,omitempty" db:"title"`
	Content     string    `json:"content,omitempty" db:"content"`
	Verified    bool      `json:"verified" db:"verified"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RatingAggregate holds aggregated rating data for a product.
type RatingAggregate struct {
	ProductID    int     `json:"product_id"`
	Average      float64 `json:"average"`
	Count        int     `json:"count"`
	Distribution map[int]int `json:"distribution"`
}

// CreateReviewRequest is the payload for creating a review.
type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Title   string `json:"title,omitempty" binding:"max=255"`
	Content string `json:"content,omitempty"`
}

// ReviewListResponse wraps a paginated list of reviews.
type ReviewListResponse struct {
	Reviews  []Review `json:"reviews"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PerPage  int      `json:"per_page"`
}
