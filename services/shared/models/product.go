package models

import "time"

// ProductTier represents a downloadable tier/version of a product.
type ProductTier struct {
	ID        int     `json:"id" db:"id"`
	ProductID int     `json:"product_id" db:"product_id"`
	TierName  string  `json:"tier_name" db:"tier_name"`
	FilePath  string  `json:"file_path,omitempty" db:"file_path"`
	FileHash  string  `json:"file_hash,omitempty" db:"file_hash"`
	PriceUSD  float64 `json:"price_usd" db:"price_usd"`
}

// ProductImage represents a product preview/gallery image.
type ProductImage struct {
	ID        int    `json:"id" db:"id"`
	ProductID int    `json:"product_id" db:"product_id"`
	URL       string `json:"url" db:"url"`
	ImageType string `json:"image_type" db:"image_type"`
	IsPrimary bool   `json:"is_primary" db:"is_primary"`
	Width     int    `json:"width,omitempty" db:"width"`
	Height    int    `json:"height,omitempty" db:"height"`
}

// Category organizes products hierarchically.
type Category struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description string    `json:"description" db:"description"`
	ParentID    *int      `json:"parent_id,omitempty" db:"parent_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Bundle groups multiple products with a combined price.
type Bundle struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Slug        string    `json:"slug" db:"slug"`
	Description string    `json:"description" db:"description"`
	PriceUSD    float64   `json:"price_usd" db:"price_usd"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Product represents a purchasable digital product.
type Product struct {
	ID                  int            `json:"id" db:"id"`
	Title               string         `json:"title" db:"title"`
	Slug                string         `json:"slug" db:"slug"`
	Description         string         `json:"description" db:"description"`
	ImageURL            string         `json:"image_url,omitempty" db:"image_url"`
	StockCount          int            `json:"stock_count,omitempty" db:"stock_count"`
	StockQuantity       int            `json:"stock_quantity,omitempty" db:"stock_quantity"`
	CategoryID          *int           `json:"category_id,omitempty" db:"category_id"`
	CategoryName        string         `json:"category_name,omitempty" db:"category_name"`
	PriceUSD            float64        `json:"price_usd" db:"price_usd"`
	AssetPath           string         `json:"asset_path,omitempty" db:"asset_path"`
	AssetHash           string         `json:"asset_hash,omitempty" db:"asset_hash"`
	Status              string         `json:"status" db:"status"`
	DownloadCountLimit  int            `json:"download_count_limit" db:"download_count_limit"`
	MaxDownloadsPerUser int            `json:"max_downloads_per_user" db:"max_downloads_per_user"`
	FileSizeBytes       *int64         `json:"file_size_bytes,omitempty" db:"file_size_bytes"`
	FileMimeType        string         `json:"file_mime_type,omitempty" db:"file_mime_type"`
	PWYWEnabled         bool           `json:"pwyw_enabled" db:"pwyw_enabled"`
	PWYWMinPrice        float64        `json:"pwyw_min_price" db:"pwyw_min_price"`
	Pinned              bool           `json:"pinned" db:"pinned"`
	SortOrder           int            `json:"sort_order" db:"sort_order"`
	Views               int            `json:"views" db:"views"`
	Downloads           int            `json:"downloads" db:"downloads"`
	PurchaseCount       int            `json:"purchase_count" db:"purchase_count"`
	DigitalFormats      string         `json:"digital_formats" db:"digital_formats"`
	AverageRating       float64        `json:"average_rating" db:"average_rating"`
	RatingCount         int            `json:"rating_count" db:"rating_count"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at" db:"updated_at"`
	Images              []ProductImage `json:"images,omitempty"`
	Tiers               []ProductTier  `json:"tiers,omitempty"`
}

// ProductWithImages embeds Product with its images for convenience.
type ProductWithImages struct {
	Product
	Images []ProductImage `json:"images,omitempty"`
}

// ProductFilter defines query parameters for product listing.
type ProductFilter struct {
	Page       int     `form:"page"`
	PerPage    int     `form:"per_page"`
	Query      string  `form:"query"`
	CategoryID *int    `form:"category_id"`
	PriceMin   float64 `form:"price_min"`
	PriceMax   float64 `form:"price_max"`
	FileType   string  `form:"file_type"`
	Rating     int     `form:"rating"`
	Sort       string  `form:"sort"`
}

// CreateProductRequest is the payload for creating a product.
type CreateProductRequest struct {
	Title        string  `json:"title" binding:"required,max=255"`
	Slug         string  `json:"slug" binding:"required,max=255"`
	Description  string  `json:"description" binding:"required"`
	CategoryID   *int    `json:"category_id,omitempty"`
	PriceUSD     float64 `json:"price_usd" binding:"required,gt=0"`
	AssetPath    string  `json:"asset_path,omitempty"`
	AssetHash    string  `json:"asset_hash,omitempty"`
	Status       string  `json:"status"`
	PWYWEnabled  bool    `json:"pwyw_enabled"`
	PWYWMinPrice float64 `json:"pwyw_min_price"`
	FileSizeBytes *int64  `json:"file_size_bytes,omitempty"`
	FileMimeType string  `json:"file_mime_type,omitempty"`
}

// UpdateProductRequest is the payload for updating a product.
type UpdateProductRequest struct {
	Title       *string  `json:"title,omitempty" binding:"max=255"`
	Slug        *string  `json:"slug,omitempty" binding:"max=255"`
	Description *string  `json:"description,omitempty"`
	CategoryID  *int     `json:"category_id,omitempty"`
	PriceUSD    *float64 `json:"price_usd,omitempty"`
	Status      *string  `json:"status,omitempty"`
}

// CreateBundleRequest is the payload for creating a bundle.
type CreateBundleRequest struct {
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Description string  `json:"description" binding:"required"`
	PriceUSD    float64 `json:"price_usd" binding:"required,gt=0"`
	ProductIDs  []int   `json:"product_ids" binding:"required,min=1"`
}

// AddProductTierRequest is the payload for adding a tier.
type AddProductTierRequest struct {
	ProductID int     `json:"product_id" binding:"required"`
	TierName  string  `json:"tier_name" binding:"required"`
	FilePath  string  `json:"file_path" binding:"required"`
	FileHash  string  `json:"file_hash,omitempty"`
	PriceUSD  float64 `json:"price_usd" binding:"required,gt=0"`
}

// AddProductImageRequest is the payload for adding a product image.
type AddProductImageRequest struct {
	ProductID int    `json:"product_id" binding:"required"`
	URL       string `json:"url" binding:"required"`
	IsPrimary bool   `json:"is_primary"`
}
