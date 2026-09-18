package models

import "time"

// Order represents a customer order.
type Order struct {
	ID                  int            `json:"id" db:"id"`
	UserID              int            `json:"user_id" db:"user_id"`
	Status              string         `json:"status" db:"status"`
	TotalUSD            float64        `json:"total_usd" db:"total_usd"`
	TotalCrypto         string         `json:"total_crypto" db:"total_crypto"`
	CryptoChain         string         `json:"crypto_chain" db:"crypto_chain"`
	CouponID             *int           `json:"coupon_id,omitempty" db:"coupon_id"`
	CouponDiscountUSD    float64        `json:"coupon_discount_usd" db:"coupon_discount_usd"`
	PaymentAddress      string         `json:"payment_address,omitempty" db:"payment_address"`
	PaymentTxHash       *string        `json:"payment_tx_hash,omitempty" db:"payment_tx_hash"`
	PaymentConfirmations int           `json:"payment_confirmations" db:"payment_confirmations"`
	PaymentConfirmedAt  *time.Time     `json:"payment_confirmed_at,omitempty" db:"payment_confirmed_at"`
	PaidAt              *time.Time     `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at" db:"updated_at"`
	Items               []OrderItem    `json:"items,omitempty"`
	UserEmail           string         `json:"user_email,omitempty" db:"-"`
	UserName            string         `json:"user_name,omitempty" db:"-"`
}

// OrderItem represents a single item within an order.
type OrderItem struct {
	ID             int        `json:"id" db:"id"`
	OrderID        int        `json:"order_id" db:"order_id"`
	ProductID      int        `json:"product_id" db:"product_id"`
	ProductTierID  *int       `json:"product_tier_id,omitempty" db:"product_tier_id"`
	ProductTitle   string     `json:"product_title,omitempty" db:"product_title"`
	ProductSlug    string     `json:"product_slug,omitempty" db:"product_slug"`
	Quantity       int        `json:"quantity" db:"quantity"`
	UnitPriceUSD   float64    `json:"unit_price_usd" db:"unit_price_usd"`
	DownloadCount  int        `json:"download_count" db:"download_count"`
	MaxDownloads   int        `json:"max_downloads" db:"max_downloads"`
	DownloadedAt   *time.Time `json:"downloaded_at,omitempty" db:"downloaded_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// GuestOrder represents an order from an unauthenticated user.
type GuestOrder struct {
	ID          int        `json:"id" db:"id"`
	Email       string     `json:"email" db:"email"`
	TotalUSD    float64    `json:"total_usd" db:"total_usd"`
	Status      string     `json:"status" db:"status"`
	CryptoChain string     `json:"crypto_chain" db:"crypto_chain"`
	CryptoAmount string    `json:"crypto_amount" db:"crypto_amount"`
	CryptoAddress string   `json:"crypto_address" db:"crypto_address"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// PaymentInfo provides crypto payment details for an order.
type PaymentInfo struct {
	OrderID        int     `json:"order_id"`
	AmountCrypto   string  `json:"amount_crypto"`
	CryptoChain    string  `json:"crypto_chain"`
	PaymentAddress string  `json:"payment_address"`
	Memo           string  `json:"memo,omitempty"`
	AmountUSD      float64 `json:"amount_usd"`
	Instructions   string  `json:"instructions"`
}

// OrderStatus constants for order lifecycle.
const (
	OrderStatusPending          = "pending"
	OrderStatusCreated          = "created"
	OrderStatusAwaitingPayment  = "awaiting_payment"
	OrderStatusPaid             = "paid"
	OrderStatusPreparation      = "preparation"
	OrderStatusDelivered        = "delivered"
	OrderStatusConfirmed        = "confirmed"
	OrderStatusShipped          = "shipped"
	OrderStatusCompleted        = "completed"
	OrderStatusCancelled        = "cancelled"
	OrderStatusRefunded         = "refunded"
	OrderStatusFailed           = "failed"
)

// OrderFilter defines query parameters for order listing.
type OrderFilter struct {
	UserID   int    `form:"user_id"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PerPage  int    `form:"per_page"`
	Sort     string `form:"sort"`
}

// CreateOrderRequest is the payload for creating an order.
type CreateOrderRequest struct {
	UserID     int    `json:"user_id" binding:"required"`
	CouponCode string `json:"coupon_code,omitempty"`
}

// CreateGuestOrderRequest is the payload for creating a guest order.
type CreateGuestOrderRequest struct {
	Email        string  `json:"email" binding:"required,email"`
	TotalUSD     float64 `json:"total_usd" binding:"required,gt=0"`
	CryptoChain  string  `json:"crypto_chain" binding:"required"`
	CryptoAmount string  `json:"crypto_amount" binding:"required"`
	CryptoAddress string `json:"crypto_address" binding:"required"`
}

// UpdateOrderStatusRequest is the payload for updating order status.
type UpdateOrderStatusRequest struct {
	OrderID int    `json:"order_id" binding:"required"`
	Status  string `json:"status" binding:"required"`
}
