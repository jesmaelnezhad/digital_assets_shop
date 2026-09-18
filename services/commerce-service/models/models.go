package models

import "time"

// User represents an authenticated user.
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Profile   *UserProfile `json:"profile,omitempty"`
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

// Order represents a purchase transaction.
type Order struct {
	ID                   int        `json:"id"`
	Status               string     `json:"status"`
	TotalUSD             string     `json:"total_usd"`
	TotalCrypto          string     `json:"total_crypto,omitempty"`
	CryptoChain          string     `json:"crypto_chain,omitempty"`
	PaymentAddress       string     `json:"payment_address,omitempty"`
	PaymentTxHash        *string    `json:"payment_tx_hash,omitempty"`
	Memo                 string     `json:"memo,omitempty"`
	PaymentConfirmations int        `json:"payment_confirmations"`
	PaymentConfirmedAt   *time.Time `json:"payment_confirmed_at,omitempty"`
	PaidAt               *time.Time `json:"paid_at,omitempty"`
	CouponID             *int       `json:"coupon_id,omitempty"`
	DiscountUSD          string     `json:"discount_usd,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	Items                []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID            int        `json:"id"`
	OrderID       int        `json:"order_id"`
	ProductID     int        `json:"product_id"`
	Quantity      int        `json:"quantity"`
	PriceUSD      string     `json:"price_usd"`
	PriceCrypto   string     `json:"price_crypto,omitempty"`
	DownloadCount int        `json:"download_count"`
	MaxDownloads  int        `json:"max_downloads"`
	DownloadedAt  *time.Time `json:"downloaded_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type GuestOrder struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	OrderID   int       `json:"order_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CartItem struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WishlistItem struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Coupon struct {
	ID            int        `json:"id"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue string     `json:"discount_value"`
	MinPurchase   string     `json:"min_purchase,omitempty"`
	MaxUses       int        `json:"max_uses"`
	CurrentUses   int        `json:"current_uses"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Active        bool       `json:"active"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CouponUsage struct {
	ID       int       `json:"id"`
	CouponID int       `json:"coupon_id"`
	OrderID  int       `json:"order_id"`
	UserID   *int      `json:"user_id,omitempty"`
	UsedAt   time.Time `json:"used_at"`
}

type RecentlyViewed struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	ViewedAt  time.Time `json:"viewed_at"`
}

type ProductComparison struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderItemRequest struct {
	ProductID int     `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity"`
	PriceUSD  float64 `json:"price_usd,omitempty"`
}

type CreateOrderRequest struct {
	Items    []OrderItemRequest `json:"items" binding:"min=0"`
	CouponCode string            `json:"coupon_code,omitempty"`
}

type CreateGuestOrderRequest struct {
	Email  string              `json:"email" binding:"required,email"`
	Items  []OrderItemRequest  `json:"items" binding:"required,min=1"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type OrderListResponse struct {
	Orders []Order `json:"orders"`
	Total  int     `json:"total"`
}

type AddToCartRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"min=1"`
}

type ValidateCouponRequest struct {
	Code       string  `json:"code" binding:"required"`
	CartTotal  float64 `json:"cart_total"`
}

type ValidateCouponResponse struct {
	Valid         bool   `json:"valid"`
	DiscountType  string `json:"discount_type,omitempty"`
	DiscountValue string `json:"discount_value,omitempty"`
}

// Admin models
type Product struct {
	ID                     int       `json:"id"`
	Title                  string    `json:"title"`
	Slug                   string    `json:"slug"`
	Description            string    `json:"description"`
	CategoryID             *int      `json:"category_id,omitempty"`
	PriceUSD               float64   `json:"price_usd"`
	Status                 string    `json:"status"`
	StockQuantity          int       `json:"stock_quantity"`
	MaxDownloadsPerUser    int       `json:"max_downloads_per_user"`
	Pinned                 bool      `json:"pinned"`
	Views                  int       `json:"views"`
	Downloads              int       `json:"downloads"`
	PurchaseCount          int       `json:"purchase_count"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
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

type CouponAdmin struct {
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

type CommunityPost struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	IsPublic  bool      `json:"is_public"`
	IsPinned  bool      `json:"is_pinned"`
	LikeCount int       `json:"like_count"`
	CommentCount int    `json:"comment_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
