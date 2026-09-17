package models

import "time"

// Stats holds aggregate admin dashboard statistics.
type Stats struct {
	TotalUsers      int     `json:"total_users"`
	TotalOrders     int     `json:"total_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalProducts   int     `json:"total_products"`
	TotalCategories int     `json:"total_categories"`
	TotalDownloads  int     `json:"total_downloads"`
}

// ReferralInfo represents a referral record for admin views.
type ReferralInfo struct {
	ID            int       `json:"id" db:"id"`
	ReferrerID    int       `json:"referrer_id" db:"referrer_id"`
	ReferrerName  string    `json:"referrer_name,omitempty" db:"referrer_name"`
	ReferredID    int       `json:"referred_id" db:"referred_id"`
	ReferredName  string    `json:"referred_name,omitempty" db:"referred_name"`
	CommissionUSD float64   `json:"commission_usd" db:"commission_usd"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Setting represents a key-value application setting.
type Setting struct {
	ID        int       `json:"id" db:"id"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	Category  string    `json:"category" db:"category"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// BulkUpdateProductsRequest is the payload for bulk product updates.
type BulkUpdateProductsRequest struct {
	ProductIDs []int  `json:"product_ids" binding:"required,min=1"`
	Status     string `json:"status,omitempty"`
	CategoryID *int   `json:"category_id,omitempty"`
}

// SetSettingRequest is the payload for setting a configuration value.
type SetSettingRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// ExportEmailsRequest is the payload for exporting user emails.
type ExportEmailsRequest struct {
	Format   string `form:"format" binding:"required,oneof=csv json"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}
