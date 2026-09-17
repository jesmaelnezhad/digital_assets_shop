// Package events defines event types and payloads for inter-service
// communication in the Pawradise microservice architecture.
package events

import "time"

// =========================================================================
// Event type constants
// =========================================================================

const (
	// Order events
	OrderCreated   = "order.created"
	OrderPaid      = "order.paid"
	OrderConfirmed = "order.confirmed"
	OrderCancelled = "order.cancelled"
	OrderRefunded  = "order.refunded"

	// Product events
	ProductCreated  = "product.created"
	ProductUpdated  = "product.updated"
	ProductDeleted  = "product.deleted"
	ProductRated    = "product.rated"

	// User events
	UserRegistered       = "user.registered"
	UserUpdated          = "user.updated"
	UserDeleted          = "user.deleted"
	UserPasswordReset    = "user.password_reset"

	// Community events
	PostCreated   = "post.created"
	PostDeleted   = "post.deleted"
	CommentAdded  = "comment.added"
	UserFollowed  = "user.followed"

	// Payment events
	PaymentDetected  = "payment.detected"
	PaymentConfirmed = "payment.confirmed"
	PaymentFailed    = "payment.failed"

	// Referral events
	ReferralCommission = "referral.commission"
	ReferralRecorded   = "referral.recorded"

	// Admin events
	StatsUpdated = "admin.stats_updated"
	SettingChanged = "admin.setting_changed"
)

// =========================================================================
// Event payloads
// =========================================================================

// OrderCreatedPayload is emitted when a new order is placed.
type OrderCreatedPayload struct {
	OrderID   int       `json:"order_id"`
	UserID    int       `json:"user_id"`
	TotalUSD  float64   `json:"total_usd"`
	Items     []OrderItemPayload `json:"items"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderItemPayload is a single item within an order event.
type OrderItemPayload struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// OrderPaidPayload is emitted when payment is confirmed.
type OrderPaidPayload struct {
	OrderID        int       `json:"order_id"`
	UserID         int       `json:"user_id"`
	PaymentTxHash  string    `json:"payment_tx_hash"`
	CryptoChain    string    `json:"crypto_chain"`
	CryptoAmount   string    `json:"crypto_amount"`
	ConfirmedAt    time.Time `json:"confirmed_at"`
}

// ProductDeletedPayload is emitted when a product is removed.
type ProductDeletedPayload struct {
	ProductID int       `json:"product_id"`
	Title     string    `json:"title"`
	DeletedAt time.Time `json:"deleted_at"`
	DeletedBy int       `json:"deleted_by"`
}

// UserDeletedPayload is emitted when a user account is deleted.
type UserDeletedPayload struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	DeletedAt time.Time `json:"deleted_at"`
}

// ReferralCommissionPayload is emitted when a referral commission is earned.
type ReferralCommissionPayload struct {
	OrderID           int       `json:"order_id"`
	ReferrerID        int       `json:"referrer_id"`
	ReferredID        int       `json:"referred_id"`
	CommissionPercent float64   `json:"commission_percent"`
	CommissionUSD     float64   `json:"commission_usd"`
	CreatedAt         time.Time `json:"created_at"`
}

// UserRegisteredPayload is emitted on new user registration.
type UserRegisteredPayload struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// PostCreatedPayload is emitted when a community post is created.
type PostCreatedPayload struct {
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// PaymentConfirmedPayload is emitted when a payment is fully confirmed.
type PaymentConfirmedPayload struct {
	OrderID       int       `json:"order_id"`
	TxHash        string    `json:"tx_hash"`
	Confirmations int       `json:"confirmations"`
	ConfirmedAt   time.Time `json:"confirmed_at"`
}

// StatsUpdatedPayload is emitted when admin stats are recalculated.
type StatsUpdatedPayload struct {
	TotalUsers     int     `json:"total_users"`
	TotalOrders    int     `json:"total_orders"`
	TotalRevenue   float64 `json:"total_revenue"`
	TotalProducts  int     `json:"total_products"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SettingChangedPayload is emitted when a setting is modified.
type SettingChangedPayload struct {
	Key       string    `json:"key"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	ChangedAt time.Time `json:"changed_at"`
	ChangedBy int       `json:"changed_by"`
}
