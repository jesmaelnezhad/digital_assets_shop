package models

import "time"

type Coupon struct {
	ID            int        `json:"id"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue string     `json:"discount_value"`
	MinPurchaseUSD string    `json:"min_purchase_usd"`
	UsageLimit    int        `json:"usage_limit"`
	TimesUsed     int        `json:"times_used"`
	ProductID     *int       `json:"product_id,omitempty"`
	IsActive      bool       `json:"is_active"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ExchangeRate struct {
	Chain     string    `json:"chain"`
	Symbol    string    `json:"symbol"`
	RateToUSD float64   `json:"rate_to_usd"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID                     int        `json:"id"`
	Title                  string     `json:"title"`
	Slug                   string     `json:"slug"`
	Description            string     `json:"description"`
	CategoryID             *int       `json:"category_id,omitempty"`
	PriceUSD               float64    `json:"price_usd"`
	Status                 string     `json:"status"`
	StockQuantity          int        `json:"stock_quantity"`
	MaxDownloadsPerUser    int        `json:"max_downloads_per_user"`
	Pinned                 bool       `json:"pinned"`
	Views                  int        `json:"views"`
	Downloads              int        `json:"downloads"`
	PurchaseCount          int        `json:"purchase_count"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type Bundle struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	PriceUSD    float64   `json:"price_usd"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CommunityPost struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Type         string    `json:"type"`
	Content      string    `json:"content"`
	IsPublic     bool      `json:"is_public"`
	IsPinned     bool      `json:"is_pinned"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Setting struct {
	ID        int       `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
