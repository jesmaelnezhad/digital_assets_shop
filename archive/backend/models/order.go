package models

import "time"

type OrderItemRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity"`
}

type Order struct {
	ID                  int        `json:"id" db:"id"`
	UserID              int        `json:"user_id" db:"user_id"`
	Status              string     `json:"status" db:"status"`
	TotalUSD            string     `json:"total_usd" db:"total_usd"`
	TotalCrypto         string     `json:"total_crypto" db:"total_crypto"`
	CryptoChain         string     `json:"crypto_chain" db:"crypto_chain"`
	PaymentAddress      string     `json:"payment_address" db:"payment_address"`
	PaymentTxHash       *string    `json:"payment_tx_hash,omitempty" db:"payment_tx_hash"`
	PaymentConfirmations int       `json:"payment_confirmations" db:"payment_confirmations"`
	PaymentConfirmedAt  *time.Time `json:"payment_confirmed_at,omitempty" db:"payment_confirmed_at"`
	PaidAt              *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	Items               []OrderItem `json:"items,omitempty"`
	UserEmail           string     `json:"user_email,omitempty" db:"-"`
	UserName            string     `json:"user_name,omitempty" db:"-"`
}

type OrderItem struct {
	ID           int        `json:"id" db:"id"`
	OrderID      int        `json:"order_id" db:"order_id"`
	ProductID    int        `json:"product_id" db:"product_id"`
	Quantity     int        `json:"quantity" db:"quantity"`
	PriceUSD     string     `json:"price_usd" db:"price_usd"`
	PriceCrypto  string     `json:"price_crypto" db:"price_crypto"`
	DownloadCount int       `json:"download_count" db:"download_count"`
	MaxDownloads int        `json:"max_downloads" db:"max_downloads"`
	Product      *Product   `json:"product,omitempty"`
	DownloadedAt *time.Time `json:"downloaded_at,omitempty" db:"downloaded_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type OrderListResponse struct {
	Orders []Order `json:"orders"`
	Total  int     `json:"total"`
}

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" binding:"required,min=1"`
}

type PaymentInfo struct {
	OrderID        int    `json:"order_id"`
	AmountCrypto   string `json:"amount_crypto"`
	CryptoChain    string `json:"crypto_chain"`
	PaymentAddress string `json:"payment_address"`
	Memo           string `json:"memo"`
	Instructions   string `json:"instructions"`
}

type ExchangeRate struct {
	ID        int       `json:"id" db:"id"`
	Chain     string    `json:"chain" db:"chain"`
	Symbol    string    `json:"symbol" db:"symbol"`
	RateToUSD string    `json:"rate_to_usd" db:"rate_to_usd"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
