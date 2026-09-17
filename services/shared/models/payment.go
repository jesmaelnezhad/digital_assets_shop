package models

import "time"

// ExchangeRate holds the current USD exchange rate for a crypto chain.
type ExchangeRate struct {
	ID        int       `json:"id" db:"id"`
	Chain     string    `json:"chain" db:"chain"`
	Symbol    string    `json:"symbol" db:"symbol"`
	RateToUSD float64   `json:"rate_to_usd" db:"rate_to_usd"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PaymentDetails provides full payment information for an order.
type PaymentDetails struct {
	OrderID        int     `json:"order_id"`
	CryptoAddress  string  `json:"crypto_address"`
	CryptoAmount   string  `json:"crypto_amount"`
	CryptoChain    string  `json:"crypto_chain"`
	CryptoMemo     string  `json:"crypto_memo,omitempty"`
	AmountUSD      float64 `json:"amount_usd"`
	ExchangeRate   float64 `json:"exchange_rate"`
}

// PaymentStatus tracks the on-chain payment verification state.
type PaymentStatus struct {
	OrderID        int    `json:"order_id"`
	Status         string `json:"status"`
	Confirmations  int    `json:"confirmations"`
	TxHash         string `json:"tx_hash,omitempty"`
}

// CalculateCryptoRequest is the payload for crypto amount calculation.
type CalculateCryptoRequest struct {
	AmountUSD float64 `json:"amount_usd" binding:"required,gt=0"`
	Chain     string  `json:"chain" binding:"required"`
}

// CalculateCryptoResponse holds the calculated crypto amount.
type CalculateCryptoResponse struct {
	CryptoAmount string  `json:"crypto_amount"`
	Chain        string  `json:"chain"`
	RateUsed     float64 `json:"rate_used"`
}

// SetExchangeRateRequest is the payload for setting an exchange rate.
type SetExchangeRateRequest struct {
	Chain     string  `json:"chain" binding:"required"`
	Symbol    string  `json:"symbol" binding:"required"`
	RateToUSD float64 `json:"rate_to_usd" binding:"required,gt=0"`
}

// PaymentStatus constants for payment lifecycle.
const (
	PaymentStatusPending     = "pending"
	PaymentStatusDetecting   = "detecting"
	PaymentStatusConfirming  = "confirming"
	PaymentStatusConfirmed   = "confirmed"
	PaymentStatusFailed      = "failed"
)
