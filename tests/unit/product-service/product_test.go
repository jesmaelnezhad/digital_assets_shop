package product_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// setupTestDB creates a mock database connection for testing.
// Returns the *sql.DB handle and the sqlmock.Sqlmock for setting expectations.
func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	return db, mock
}

// =============================================================================
// Product CRUD Tests
// =============================================================================

func TestProductService_GetProducts(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_list_with_pagination", func(t *testing.T) {
		productRows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
			"asset_path", "asset_hash", "status", "download_count_limit",
			"max_downloads_per_user", "file_size_bytes", "file_mime_type",
			"created_at", "updated_at",
		}).AddRow(
			1, "Cool Asset Pack", "cool-asset-pack", "A cool asset pack", 1, "3D Models", "29.99",
			"/assets/cool.zip", "hash123", "active", 5,
			3, int64(1048576), "application/zip",
			"2024-01-01", "2024-01-01",
		).AddRow(
			2, "Fantasy Pack", "fantasy-pack", "Fantasy themed assets", 2, "Textures", "19.99",
			"/assets/fantasy.zip", "hash456", "active", 5,
			3, int64(5242880), "application/zip",
			"2024-01-02", "2024-01-02",
		)

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(2),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(productRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_with_search_filter", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(1),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_with_category_filter", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(0),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_with_sort_order", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(0),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

func TestProductService_GetProductBySlug(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_get_product_by_slug", func(t *testing.T) {
		slug := "cool-asset-pack"

		productRows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
			"asset_path", "asset_hash", "status", "download_count_limit",
			"max_downloads_per_user", "file_size_bytes", "file_mime_type",
			"created_at", "updated_at",
		}).AddRow(
			1, "Cool Asset Pack", "cool-asset-pack", "A cool asset pack", 1, "3D Models", "29.99",
			"/assets/cool.zip", "hash123", "active", 5,
			3, int64(1048576), "application/zip",
			"2024-01-01", "2024-01-01",
		)

		imageRows := sqlmock.NewRows([]string{
			"id", "product_id", "url", "is_primary", "created_at",
		}).AddRow(
			1, 1, "/images/cool-main.jpg", true, "2024-01-01",
		).AddRow(
			2, 1, "/images/cool-alt.jpg", false, "2024-01-01",
		)

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(slug).WillReturnRows(productRows)
		mock.ExpectQuery("SELECT id, product_id, url, is_primary").WithArgs(1).WillReturnRows(imageRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_product_not_found", func(t *testing.T) {
		slug := "nonexistent-product"

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(slug).WillReturnError(sql.ErrNoRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_product_without_images", func(t *testing.T) {
		slug := "no-images-product"

		productRows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
			"asset_path", "asset_hash", "status", "download_count_limit",
			"max_downloads_per_user", "file_size_bytes", "file_mime_type",
			"created_at", "updated_at",
		}).AddRow(
			2, "No Images", "no-images-product", "Product with no images", nil, "", "9.99",
			"/assets/noimages.zip", "hash789", "active", 5,
			3, int64(204800), "application/zip",
			"2024-01-03", "2024-01-03",
		)

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(slug).WillReturnRows(productRows)
		mock.ExpectQuery("SELECT id, product_id, url, is_primary").WithArgs(2).WillReturnRows(
			sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

func TestProductService_CreateProduct(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_create_product", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products WHERE slug").WithArgs("new-product").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(0),
		)
		mock.ExpectExec("INSERT INTO products").WillReturnResult(
			sqlmock.NewResult(1, 1),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(1).WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				1, "New Product", "new-product", "A new product", 1, "3D Models", "39.99",
				"", "", "active", 5,
				3, nil, "",
				"2024-01-01", "2024-01-01",
			),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_slug_already_exists", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products WHERE slug").WithArgs("existing-slug").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(1),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_invalid_category", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products WHERE slug").WithArgs("bad-category-product").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(0),
		)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM categories WHERE id").WithArgs(999).WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(0),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_with_draft_status", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products WHERE slug").WithArgs("draft-product").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(0),
		)
		mock.ExpectExec("INSERT INTO products").WillReturnResult(
			sqlmock.NewResult(2, 1),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(2).WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				2, "Draft Product", "draft-product", "A draft product", nil, "", "0.00",
				"", "", "draft", 5,
				3, nil, "",
				"2024-01-01", "2024-01-01",
			),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_update_title_and_price", func(t *testing.T) {
		productID := 1

		mock.ExpectQuery("SELECT id, title, slug, description, category_id, price_usd, status FROM products WHERE id").
			WithArgs(productID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "price_usd", "status",
			}).AddRow(
				1, "Old Title", "old-title", "Old description", 1, "19.99", "active",
			))

		mock.ExpectExec("UPDATE products SET").WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(productID).WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				1, "New Title", "old-title", "Old description", 1, "3D Models", "29.99",
				"", "", "active", 5,
				3, nil, "",
				"2024-01-01", "2024-01-02",
			),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_product_not_found", func(t *testing.T) {
		productID := 999

		mock.ExpectQuery("SELECT id, title, slug, description, category_id, price_usd, status FROM products WHERE id").
			WithArgs(productID).
			WillReturnError(sql.ErrNoRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_slug_already_exists", func(t *testing.T) {
		productID := 1

		mock.ExpectQuery("SELECT id, title, slug, description, category_id, price_usd, status FROM products WHERE id").
			WithArgs(productID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "price_usd", "status",
			}).AddRow(
				1, "My Product", "my-product", "Description", 1, "19.99", "active",
			))

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products WHERE slug").WithArgs("taken-slug", productID).WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(1),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_no_fields_to_update", func(t *testing.T) {
		productID := 1

		mock.ExpectQuery("SELECT id, title, slug, description, category_id, price_usd, status FROM products WHERE id").
			WithArgs(productID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "price_usd", "status",
			}).AddRow(
				1, "My Product", "my-product", "Description", 1, "19.99", "active",
			))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_update_status_to_archived", func(t *testing.T) {
		productID := 1

		mock.ExpectQuery("SELECT id, title, slug, description, category_id, price_usd, status FROM products WHERE id").
			WithArgs(productID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "price_usd", "status",
			}).AddRow(
				1, "My Product", "my-product", "Description", 1, "19.99", "active",
			))

		mock.ExpectExec("UPDATE products SET").WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(productID).WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				1, "My Product", "my-product", "Description", 1, "3D Models", "19.99",
				"", "", "archived", 5,
				3, nil, "",
				"2024-01-01", "2024-01-02",
			),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_archive_product", func(t *testing.T) {
		productID := 1

		mock.ExpectExec("UPDATE products SET status").WithArgs(productID).WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_product_not_found", func(t *testing.T) {
		productID := 999

		mock.ExpectExec("UPDATE products SET status").WithArgs(productID).WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

// =============================================================================
// Category Tests
// =============================================================================

func TestProductService_GetCategories(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_list_categories", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "name", "slug", "description", "parent_id", "created_at", "updated_at",
		}).AddRow(
			1, "3D Models", "3d-models", "3D model assets", nil, "2024-01-01", "2024-01-01",
		).AddRow(
			2, "Textures", "textures", "Texture assets", nil, "2024-01-01", "2024-01-01",
		).AddRow(
			3, "PBR Textures", "pbr-textures", "PBR texture assets", 2, "2024-01-01", "2024-01-01",
		)

		mock.ExpectQuery("SELECT id, name, slug, description, parent_id, created_at, updated_at FROM categories").WillReturnRows(rows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_empty_categories", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, slug, description, parent_id, created_at, updated_at FROM categories").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "name", "slug", "description", "parent_id", "created_at", "updated_at",
			}),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

// =============================================================================
// Product Tier Tests
// =============================================================================

func TestProductService_AddProductTier(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_add_tier", func(t *testing.T) {
		productID := 1

		mock.ExpectExec("INSERT INTO product_tiers").WithArgs(
			productID, "Premium", "/assets/premium.zip", "hash-premium", "49.99", 1,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_add_multiple_tiers", func(t *testing.T) {
		productID := 1

		mock.ExpectExec("INSERT INTO product_tiers").WithArgs(
			productID, "Basic", "/assets/basic.zip", "hash-basic", "19.99", 1,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO product_tiers").WithArgs(
			productID, "Pro", "/assets/pro.zip", "hash-pro", "39.99", 2,
		).WillReturnResult(sqlmock.NewResult(2, 1))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_product_not_found", func(t *testing.T) {
		productID := 999

		mock.ExpectExec("INSERT INTO product_tiers").WithArgs(
			productID, "Premium", "/assets/premium.zip", "hash-premium", "49.99", 1,
		).WillReturnError(sql.ErrNoRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

func TestProductService_UpdateProductTier(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_update_tier_name_and_price", func(t *testing.T) {
		tierID := 1

		mock.ExpectExec("UPDATE product_tiers SET").WithArgs(
			"Premium Plus", "59.99", tierID,
		).WillReturnResult(sqlmock.NewResult(0, 1))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_update_tier_file", func(t *testing.T) {
		tierID := 1

		mock.ExpectExec("UPDATE product_tiers SET").WithArgs(
			"/assets/premium-v2.zip", "hash-v2", tierID,
		).WillReturnResult(sqlmock.NewResult(0, 1))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_tier_not_found", func(t *testing.T) {
		tierID := 999

		mock.ExpectExec("UPDATE product_tiers SET").WithArgs(
			"Updated Name", "29.99", tierID,
		).WillReturnResult(sqlmock.NewResult(0, 0))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

func TestProductService_DeleteProductTier(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_delete_tier", func(t *testing.T) {
		tierID := 1

		mock.ExpectExec("DELETE FROM product_tiers").WithArgs(tierID).WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_tier_not_found", func(t *testing.T) {
		tierID := 999

		mock.ExpectExec("DELETE FROM product_tiers").WithArgs(tierID).WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

// =============================================================================
// Product Image Tests
// =============================================================================

func TestProductService_AddProductImage(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_add_primary_image", func(t *testing.T) {
		productID := 1

		mock.ExpectExec("INSERT INTO product_images").WithArgs(
			productID, "/images/product-main.jpg", "full", true, 1920, 1080,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_add_preview_image", func(t *testing.T) {
		productID := 1

		mock.ExpectExec("INSERT INTO product_images").WithArgs(
			productID, "/images/product-preview.jpg", "preview", false, 800, 600,
		).WillReturnResult(sqlmock.NewResult(2, 1))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_add_thumbnail_image", func(t *testing.T) {
		productID := 1

		mock.ExpectExec("INSERT INTO product_images").WithArgs(
			productID, "/images/product-thumb.jpg", "thumbnail", false, 200, 200,
		).WillReturnResult(sqlmock.NewResult(3, 1))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_product_not_found", func(t *testing.T) {
		productID := 999

		mock.ExpectExec("INSERT INTO product_images").WithArgs(
			productID, "/images/product.jpg", "full", true, 1920, 1080,
		).WillReturnError(sql.ErrNoRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

func TestProductService_DeleteProductImage(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_delete_image", func(t *testing.T) {
		imageID := 1

		mock.ExpectExec("DELETE FROM product_images").WithArgs(imageID).WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_image_not_found", func(t *testing.T) {
		imageID := 999

		mock.ExpectExec("DELETE FROM product_images").WithArgs(imageID).WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

// =============================================================================
// Preview Generation Tests
// =============================================================================

func TestProductService_GeneratePreviews(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_trigger_preview_generation", func(t *testing.T) {
		productID := 1

		productRows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
			"asset_path", "asset_hash", "status", "download_count_limit",
			"max_downloads_per_user", "file_size_bytes", "file_mime_type",
			"created_at", "updated_at",
		}).AddRow(
			1, "Cool Asset Pack", "cool-asset-pack", "A cool asset pack", 1, "3D Models", "29.99",
			"/assets/cool.zip", "hash123", "active", 5,
			3, int64(1048576), "application/zip",
			"2024-01-01", "2024-01-01",
		)

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(productID).WillReturnRows(productRows)
		mock.ExpectExec("INSERT INTO product_images").WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_product_not_found", func(t *testing.T) {
		productID := 999

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(productID).WillReturnError(sql.ErrNoRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_regenerate_all_previews", func(t *testing.T) {
		productID := 1

		productRows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
			"asset_path", "asset_hash", "status", "download_count_limit",
			"max_downloads_per_user", "file_size_bytes", "file_mime_type",
			"created_at", "updated_at",
		}).AddRow(
			1, "Cool Asset Pack", "cool-asset-pack", "A cool asset pack", 1, "3D Models", "29.99",
			"/assets/cool.zip", "hash123", "active", 5,
			3, int64(1048576), "application/zip",
			"2024-01-01", "2024-01-01",
		)

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(productID).WillReturnRows(productRows)
		mock.ExpectExec("DELETE FROM product_images").WithArgs(productID).WillReturnResult(
			sqlmock.NewResult(0, 3),
		)
		mock.ExpectExec("INSERT INTO product_images").WillReturnResult(
			sqlmock.NewResult(4, 1),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

// =============================================================================
// Pin Product Tests
// =============================================================================

func TestProductService_PinProduct(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_pin_product", func(t *testing.T) {
		productID := 1

		mock.ExpectExec("UPDATE products SET pinned").WithArgs(true, productID).WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_unpin_product", func(t *testing.T) {
		productID := 1

		mock.ExpectExec("UPDATE products SET pinned").WithArgs(false, productID).WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_product_not_found", func(t *testing.T) {
		productID := 999

		mock.ExpectExec("UPDATE products SET pinned").WithArgs(true, productID).WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

// =============================================================================
// Bundle Tests
// =============================================================================

func TestProductService_GetBundles(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_list_bundles", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "price_usd", "status", "created_at", "updated_at",
		}).AddRow(
			1, "Starter Bundle", "starter-bundle", "Great starter pack", "49.99", "active", "2024-01-01", "2024-01-01",
		).AddRow(
			2, "Pro Bundle", "pro-bundle", "Professional asset pack", "99.99", "active", "2024-01-02", "2024-01-02",
		)

		mock.ExpectQuery("SELECT id, title, slug, description, price_usd, status, created_at, updated_at FROM bundles").WillReturnRows(rows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_empty_bundles", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, title, slug, description, price_usd, status, created_at, updated_at FROM bundles").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "price_usd", "status", "created_at", "updated_at",
			}),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

func TestProductService_GetBundle(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_get_bundle_detail", func(t *testing.T) {
		bundleID := 1

		bundleRows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "price_usd", "status", "created_at", "updated_at",
		}).AddRow(
			1, "Starter Bundle", "starter-bundle", "Great starter pack", "49.99", "active", "2024-01-01", "2024-01-01",
		)

		itemRows := sqlmock.NewRows([]string{
			"id", "bundle_id", "product_id", "quantity", "created_at",
		}).AddRow(
			1, 1, 1, 1, "2024-01-01",
		).AddRow(
			2, 1, 2, 1, "2024-01-01",
		)

		mock.ExpectQuery("SELECT id, title, slug, description, price_usd, status, created_at, updated_at FROM bundles WHERE id").
			WithArgs(bundleID).
			WillReturnRows(bundleRows)
		mock.ExpectQuery("SELECT id, bundle_id, product_id, quantity, created_at FROM bundle_items WHERE bundle_id").
			WithArgs(bundleID).
			WillReturnRows(itemRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_bundle_not_found", func(t *testing.T) {
		bundleID := 999

		mock.ExpectQuery("SELECT id, title, slug, description, price_usd, status, created_at, updated_at FROM bundles WHERE id").
			WithArgs(bundleID).
			WillReturnError(sql.ErrNoRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_bundle_without_items", func(t *testing.T) {
		bundleID := 2

		bundleRows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "price_usd", "status", "created_at", "updated_at",
		}).AddRow(
			2, "Empty Bundle", "empty-bundle", "No items yet", "0.00", "draft", "2024-01-01", "2024-01-01",
		)

		mock.ExpectQuery("SELECT id, title, slug, description, price_usd, status, created_at, updated_at FROM bundles WHERE id").
			WithArgs(bundleID).
			WillReturnRows(bundleRows)
		mock.ExpectQuery("SELECT id, bundle_id, product_id, quantity, created_at FROM bundle_items WHERE bundle_id").
			WithArgs(bundleID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "bundle_id", "product_id", "quantity", "created_at",
			}))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

func TestProductService_CreateBundle(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_create_bundle", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO bundles").WithArgs(
			"New Bundle", "new-bundle", "A new bundle", "79.99", "active",
		).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO bundle_items").WithArgs(
			1, 1, 1,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO bundle_items").WithArgs(
			1, 2, 1,
		).WillReturnResult(sqlmock.NewResult(2, 1))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("failure_duplicate_slug", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO bundles").WithArgs(
			"Duplicate Bundle", "duplicate-bundle", "A bundle", "49.99", "active",
		).WillReturnError(sql.ErrNoRows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_create_draft_bundle", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO bundles").WithArgs(
			"Draft Bundle", "draft-bundle", "A draft bundle", "0.00", "draft",
		).WillReturnResult(sqlmock.NewResult(2, 1))

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

// =============================================================================
// Recommendation Tests
// =============================================================================

func TestProductService_GetRecommendations(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("success_get_related_products", func(t *testing.T) {
		productID := 1

		rows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
			"asset_path", "asset_hash", "status", "download_count_limit",
			"max_downloads_per_user", "file_size_bytes", "file_mime_type",
			"created_at", "updated_at",
		}).AddRow(
			2, "Related Pack", "related-pack", "Related product", 1, "3D Models", "24.99",
			"/assets/related.zip", "hash-rel", "active", 5,
			3, int64(524288), "application/zip",
			"2024-01-01", "2024-01-01",
		).AddRow(
			3, "Similar Pack", "similar-pack", "Similar product", 1, "3D Models", "34.99",
			"/assets/similar.zip", "hash-sim", "active", 5,
			3, int64(7340032), "application/zip",
			"2024-01-02", "2024-01-02",
		)

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(productID).WillReturnRows(rows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_no_recommendations", func(t *testing.T) {
		productID := 99

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(productID).WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}),
		)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})

	t.Run("success_recommendations_by_tags", func(t *testing.T) {
		productID := 1

		rows := sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
			"asset_path", "asset_hash", "status", "download_count_limit",
			"max_downloads_per_user", "file_size_bytes", "file_mime_type",
			"created_at", "updated_at",
		}).AddRow(
			4, "Tagged Pack", "tagged-pack", "Tagged product", 2, "Textures", "14.99",
			"/assets/tagged.zip", "hash-tag", "active", 5,
			3, int64(3145728), "application/zip",
			"2024-01-03", "2024-01-03",
		)

		mock.ExpectQuery("SELECT p\\.id, p\\.title").WithArgs(productID).WillReturnRows(rows)

		// TODO: Call ServiceMethod()
		_ = mock
		_ = db
	})
}

// =============================================================================
// Search and Filtering Tests
// =============================================================================

func TestSearchAndFiltering(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	t.Run("search_by_text", func(t *testing.T) {
		searchQuery := "fantasy"

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(2),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				1, "Fantasy Pack", "fantasy-pack", "Fantasy assets", 1, "3D Models", "29.99",
				"/assets/fantasy.zip", "hash-fan", "active", 5,
				3, int64(1048576), "application/zip",
				"2024-01-01", "2024-01-01",
			).AddRow(
				2, "Fantasy Landscape", "fantasy-landscape", "Fantasy landscape", 2, "Textures", "19.99",
				"/assets/landscape.zip", "hash-land", "active", 5,
				3, int64(5242880), "application/zip",
				"2024-01-02", "2024-01-02",
			),
		)

		// TODO: Call ServiceMethod()
		_ = searchQuery
		_ = mock
		_ = db
	})

	t.Run("filter_by_category", func(t *testing.T) {
		categorySlug := "3d-models"

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(3),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				1, "Character Pack", "character-pack", "3D characters", 1, "3D Models", "49.99",
				"/assets/char.zip", "hash-char", "active", 5,
				3, int64(2097152), "application/zip",
				"2024-01-01", "2024-01-01",
			),
		)

		// TODO: Call ServiceMethod()
		_ = categorySlug
		_ = mock
		_ = db
	})

	t.Run("filter_by_price_range", func(t *testing.T) {
		minPrice := 10.00
		maxPrice := 50.00

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(5),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				1, "Mid Range Pack", "mid-range", "Mid range product", 1, "3D Models", "29.99",
				"/assets/mid.zip", "hash-mid", "active", 5,
				3, int64(1048576), "application/zip",
				"2024-01-01", "2024-01-01",
			),
		)

		// TODO: Call ServiceMethod()
		_ = minPrice
		_ = maxPrice
		_ = mock
		_ = db
	})

	t.Run("filter_by_file_type", func(t *testing.T) {
		fileType := "application/zip"

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(4),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				1, "ZIP Pack", "zip-pack", "ZIP product", 1, "3D Models", "19.99",
				"/assets/zip.zip", "hash-zip", "active", 5,
				3, int64(1048576), "application/zip",
				"2024-01-01", "2024-01-01",
			),
		)

		// TODO: Call ServiceMethod()
		_ = fileType
		_ = mock
		_ = db
	})

	t.Run("filter_by_rating", func(t *testing.T) {
		minRating := 4.0

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(2),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				1, "Top Rated Pack", "top-rated", "Highly rated", 1, "3D Models", "39.99",
				"/assets/top.zip", "hash-top", "active", 5,
				3, int64(1048576), "application/zip",
				"2024-01-01", "2024-01-01",
			),
		)

		// TODO: Call ServiceMethod()
		_ = minRating
		_ = mock
		_ = db
	})

	t.Run("combined_search_and_filter", func(t *testing.T) {
		searchQuery := "fantasy"
		categorySlug := "3d-models"
		minPrice := 10.00
		maxPrice := 100.00
		fileType := "application/zip"
		minRating := 3.0

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(1),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}).AddRow(
				1, "Fantasy Character", "fantasy-char", "Fantasy character pack", 1, "3D Models", "49.99",
				"/assets/fantasy-char.zip", "hash-fc", "active", 5,
				3, int64(1048576), "application/zip",
				"2024-01-01", "2024-01-01",
			),
		)

		// TODO: Call ServiceMethod()
		_ = searchQuery
		_ = categorySlug
		_ = minPrice
		_ = maxPrice
		_ = fileType
		_ = minRating
		_ = mock
		_ = db
	})

	t.Run("search_no_results", func(t *testing.T) {
		searchQuery := "nonexistent-query-xyz"

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(0),
		)
		mock.ExpectQuery("SELECT p\\.id, p\\.title").WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "title", "slug", "description", "category_id", "category_name", "price_usd",
				"asset_path", "asset_hash", "status", "download_count_limit",
				"max_downloads_per_user", "file_size_bytes", "file_mime_type",
				"created_at", "updated_at",
			}),
		)

		// TODO: Call ServiceMethod()
		_ = searchQuery
		_ = mock
		_ = db
	})
}
