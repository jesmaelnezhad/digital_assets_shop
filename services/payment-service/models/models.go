package models

import "time"

// ExchangeRate represents a crypto-to-USD exchange rate.
type ExchangeRate struct {
	ID        int       `json:"id" db:"id"`
	Chain     string    `json:"chain" db:"chain"`
	Symbol    string    `json:"symbol" db:"symbol"`
	RateToUSD string    `json:"rate_to_usd" db:"rate_to_usd"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PaymentDetail contains the information needed to complete a payment.
type PaymentDetail struct {
	OrderID        int    `json:"order_id"`
	AmountCrypto   string `json:"amount_crypto"`
	CryptoChain    string `json:"crypto_chain"`
	PaymentAddress string `json:"payment_address"`
	Memo           string `json:"memo,omitempty"`
	ExpiresAt      string `json:"expires_at"`
	Instructions   string `json:"instructions"`
}

// PaymentStatus tracks the current state of a payment.
type PaymentStatus struct {
	OrderID              int    `json:"order_id"`
	Status               string `json:"status"`
	AmountCrypto         string `json:"amount_crypto"`
	CryptoChain          string `json:"crypto_chain"`
	PaymentAddress       string `json:"payment_address"`
	Confirmations        int    `json:"confirmations"`
	RequiredConfirmations int   `json:"required_confirmations"`
}

// ConfirmPaymentRequest is the payload for confirming a payment.
type ConfirmPaymentRequest struct {
	TxHash string `json:"tx_hash" binding:"required"`
}

// SetExchangeRateRequest is the payload for setting an exchange rate.
type SetExchangeRateRequest struct {
	Chain     string `json:"chain" binding:"required"`
	Symbol    string `json:"symbol" binding:"required"`
	RateToUSD string `json:"rate_to_usd" binding:"required"`
}

// ExchangeRateListResponse wraps a list of exchange rates.
type ExchangeRateListResponse struct {
	Rates []ExchangeRate `json:"rates"`
}
