package models

import "time"

type Product struct {
	ID                 int       `json:"id"`
	Title              string    `json:"title"`
	Slug               string    `json:"slug"`
	Description        string    `json:"description"`
	ImageURL           string    `json:"image_url,omitempty"`
	StockCount         int       `json:"stock_count"`
	CategoryID         int       `json:"category_id"`
	PriceUSD           float64   `json:"price_usd"`
	Status             string    `json:"status"`
	StockQuantity      int       `json:"stock_quantity"`
	MaxDownloadsPerUser int      `json:"max_downloads_per_user"`
	PreviewImages      int       `json:"preview_images"`
	DownloadWindowHours int      `json:"download_window_hours"`
	FreeDownload       bool      `json:"free_download"`
	SortOrder           int       `json:"sort_order"`
	DigitalFormats      string    `json:"digital_formats"`
	IsPinned            bool      `json:"is_pinned"`
	Pinned             bool      `json:"pinned"`
	Views              int       `json:"views"`
	Downloads          int       `json:"downloads"`
	PurchaseCount      int       `json:"purchase_count"`
	Tags               string    `json:"tags"`
	IsPwyw             bool      `json:"is_pwyw"`
	PwywMinPrice       float64   `json:"pwyw_min_price"`
	PinnedAt           *time.Time `json:"pinned_at,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	CategoryName       string    `json:"category_name,omitempty"`
	CategorySlug       string    `json:"category_slug,omitempty"`
}

type ProductImage struct {
	ID             int       `json:"id"`
	ProductID      int       `json:"product_id"`
	URL            string    `json:"url"`
	MimeType       string    `json:"mime_type"`
	AltText        string    `json:"alt_text"`
	IsPrimary      bool      `json:"is_primary"`
	ImageType      string    `json:"image_type"`
	Width          int       `json:"width"`
	Height         int       `json:"height"`
	FileSizeBytes  int64     `json:"file_size_bytes"`
	StoragePath    string    `json:"storage_path"`
	SortOrder      int       `json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
}

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	ParentID    int       `json:"parent_id"`
	ImageURL    string    `json:"image_url"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Bundle struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	PriceUSD    float64   `json:"price_usd"`
	Status      string    `json:"status"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Items       string    `json:"items"`
}

type Tag struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Count     int       `json:"count"`
	ImageURL  string    `json:"image_url"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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
}

type ProductTier struct {
	ID              int     `json:"id"`
	ProductID       int     `json:"product_id"`
	TierName        string  `json:"tier_name"`
	PriceUSD        float64 `json:"price_usd"`
	DownloadCount   int     `json:"download_count"`
	DownloadLimit   int     `json:"download_limit"`
	IsActive        bool    `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdateProductRequest struct {
	Title       *string  `json:"title,omitempty" binding:"max=255"`
	Slug        *string  `json:"slug,omitempty" binding:"max=255"`
	Description *string  `json:"description,omitempty"`
	CategoryID  *int     `json:"category_id,omitempty"`
	PriceUSD    *float64 `json:"price_usd,omitempty"`
	Status      *string  `json:"status,omitempty"`
}
