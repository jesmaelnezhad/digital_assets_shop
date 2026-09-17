package models

import "time"

type Category struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description string    `json:"description" db:"description"`
	ParentID    *int      `json:"parent_id" db:"parent_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Product struct {
	ID                   int        `json:"id" db:"id"`
	Title                string     `json:"title" db:"title"`
	Slug                 string     `json:"slug" db:"slug"`
	Description          string     `json:"description" db:"description"`
	CategoryID           *int       `json:"category_id" db:"category_id"`
	CategoryName         string     `json:"category_name,omitempty" db:"-"`
	PriceUSD             string     `json:"price_usd" db:"price_usd"` // string for precise decimal
	AssetPath            string     `json:"asset_path" db:"asset_path"`
	AssetHash            string     `json:"asset_hash" db:"asset_hash"`
	Status               string     `json:"status" db:"status"`
	DownloadCountLimit   int        `json:"download_count_limit" db:"download_count_limit"`
	MaxDownloadsPerUser  int        `json:"max_downloads_per_user" db:"max_downloads_per_user"`
	FileSizeBytes        *int64     `json:"file_size_bytes,omitempty" db:"file_size_bytes"`
	FileMimeType         string     `json:"file_mime_type,omitempty" db:"file_mime_type"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

type ProductImage struct {
	ID        int       `json:"id" db:"id"`
	ProductID int       `json:"product_id" db:"product_id"`
	URL       string    `json:"url" db:"url"`
	IsPrimary bool      `json:"is_primary" db:"is_primary"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ProductWithImages struct {
	Product
	Images []ProductImage `json:"images,omitempty"`
}

type ProductListResponse struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PerPage  int       `json:"per_page"`
}

type CreateProductRequest struct {
	Title       string  `json:"title" binding:"required,max=255"`
	Slug        string  `json:"slug" binding:"required,max=255"`
	Description string  `json:"description" binding:"required"`
	CategoryID  *int    `json:"category_id"`
	PriceUSD    float64 `json:"price_usd" binding:"required,gt=0"`
	Status      string  `json:"status"`
}

type UpdateProductRequest struct {
	Title       string  `json:"title,omitempty" binding:"max=255"`
	Slug        string  `json:"slug,omitempty" binding:"max=255"`
	Description string  `json:"description,omitempty"`
	CategoryID  *int    `json:"category_id,omitempty"`
	PriceUSD    float64 `json:"price_usd,omitempty"`
	Status      string  `json:"status,omitempty"`
}
