package handlers

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestAuth_Register_Valid(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`INSERT INTO users.*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	code, _, result := R("POST", "/api/v1/register", map[string]string{
		"email": "test@test.com", "password": "test1234", "name": "Test User",
	}, nil)
	assert.Equal(t, 201, code)
	assert.Contains(t, result, "token")
	expectExpectations(t, mock)
}

func TestAuth_Login_Valid(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs("test@test.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
			AddRow(1, "test@test.com", "$2a$12$/9npYy2uquhE8JKPDE8nB.acwc6VuH7N54sROUANpvWRQpoDKM7AW", "Test User", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	code, _, result := R("POST", "/api/v1/login", map[string]string{
		"email": "test@test.com", "password": "test1234",
	}, nil)
	assert.Equal(t, 200, code)
	assert.Contains(t, result, "token")
	expectExpectations(t, mock)
}

func TestAuth_Login_WrongPassword(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs("test@test.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
			AddRow(1, "test@test.com", "$2a$12$abcdefghijklmnopqrstuvwxyzABCDEFabcdef/9npYy2uquhE8JKPDE8nB.acwc6VuH7N54sROUANpvWRQpoDKM7AW", "Test User", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	code, _, _ := R("POST", "/api/v1/login", map[string]string{
		"email": "test@test.com", "password": "wrongpassword",
	}, nil)
	assert.Equal(t, 401, code)
	expectExpectations(t, mock)
}

func TestAuth_GetMe_Authorized(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(1, "test@test.com", "Test User", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/me", nil, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	assert.Equal(t, float64(1), result["id"])
	expectExpectations(t, mock)
}

func TestAuth_UpdateProfile_Valid(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(1, "test@test.com", "Test User", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"bio", "wallet_address"}))
	mock.ExpectExec(`UPDATE users SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(1, "test@test.com", "New Name", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	code, _, result := R("PUT", "/api/v1/me", map[string]string{"name": "New Name"}, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	assert.Equal(t, "New Name", result["name"])
	expectExpectations(t, mock)
}

func TestAuth_Logout_Valid(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`DELETE FROM invalidated_tokens.*`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO invalidated_tokens.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, result := R("POST", "/api/v1/logout", nil, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	assert.Contains(t, result, "message")
	expectExpectations(t, mock)
}

func TestProducts_GetCategories(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, name, slug, description, parent_id, created_at, updated_at FROM categories`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "description", "parent_id", "created_at", "updated_at"}).
			AddRow(1, "Icons", "icons", "Icon packs", nil, mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).
			AddRow(2, "Fonts", "fonts", "Programming fonts", nil, mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/categories", nil, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, result["categories"].([]interface{}), 2)
	expectExpectations(t, mock)
}

func TestProducts_GetProducts(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM products.*`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT p.id, p.title.*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).
			AddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).
			AddRow(2, "Another Product", "another-product", "Another test", nil, "", "20.00", "", "", "active", 50, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/products", nil, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, result["products"], 2)
	expectExpectations(t, mock)
}

func TestProducts_GetProductBySlug(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM products WHERE slug = \$1`).
		WithArgs("test-product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT p.id, p.title.*`).
		WithArgs("test-product").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).
			AddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	mock.ExpectQuery(`SELECT id, product_id, url.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).
			AddRow(1, 1, "http://example.com/img.png", true, mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/products/test-product", nil, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, "test-product", result["slug"])
	expectExpectations(t, mock)
}

func TestProducts_SearchProducts(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM products.*`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT p.id, p.title.*`).
		WithArgs("%test%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).
			AddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	mock.ExpectQuery(`SELECT id, product_id, url.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).
			AddRow(1, 1, "http://example.com/img.png", true, mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/products/search?q=test", nil, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, result["products"], 1)
	expectExpectations(t, mock)
}

func TestAuth_CreateProduct_Unauthorized(t *testing.T) {
	code, _, _ := R("POST", "/api/v1/products", map[string]string{"name": "Test"}, nil)
	assert.Equal(t, 401, code)
}

func TestOrders_CreateOrder_Unauthorized(t *testing.T) {
	code, _, _ := R("POST", "/api/v1/orders", nil, nil)
	assert.Equal(t, 401, code)
}

func TestOrders_GetUserOrders_Unauthorized(t *testing.T) {
	code, _, _ := R("GET", "/api/v1/orders", nil, nil)
	assert.Equal(t, 401, code)
}

func TestOrders_GetOrder_Unauthorized(t *testing.T) {
	code, _, _ := R("GET", "/api/v1/orders/1", nil, nil)
	assert.Equal(t, 401, code)
}

func TestCommunity_Profile(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, user_id, bio, avatar_url, wallet_address, created_at, updated_at FROM user_profiles WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "bio", "avatar_url", "wallet_address", "created_at", "updated_at"}).
			AddRow(1, 1, "Hello world", nil, "0x1234...", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	mock.ExpectQuery(`SELECT email, name FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"email", "name"}).AddRow("test@test.com", "Test User"))
	code, _, _ := R("GET", "/api/v1/profile", nil, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestCommunity_Feed(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content.*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "email", "name", "avatar_url", "count_likes", "count_comments", "following"}).
			AddRow(1, 1, "Content 1", mockTime(2024, time.January, 1), "test@test.com", "Test User", "", 0, 0, false).
			AddRow(2, 1, "Content 2", mockTime(2024, time.January, 1), "test@test.com", "Test User", "", 0, 0, false))
	code, _, result := R("GET", "/api/v1/community/feed", nil, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	assert.Len(t, result["posts"], 2)
	expectExpectations(t, mock)
}

func TestCommunity_CreatePost_Unauthorized(t *testing.T) {
	code, _, _ := R("POST", "/api/v1/community/posts", map[string]string{"title": "Test"}, nil)
	assert.Equal(t, 401, code)
}

func TestAdmin_BulkProductStatus(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`UPDATE products SET.*`).
		WillReturnResult(sqlmock.NewResult(2, 2))
	code, _, result := R("POST", "/api/v1/admin/bulk/product-status", map[string]string{
		"status": "active", "ids": "1,2",
	}, map[string]string{"Authorization": "Bearer " + makeAdminToken()})
	assert.Equal(t, 200, code)
	assert.Equal(t, float64(2), result["updated"])
	expectExpectations(t, mock)
}

func TestAdmin_OrderStatusPaid(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`UPDATE orders SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, result := R("PUT", "/api/v1/admin/orders/1/status/paid", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Contains(t, result, "status")
	expectExpectations(t, mock)
}

func TestAdmin_OrderStatusCompleted(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`UPDATE orders SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/orders/1/status/completed", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestAdmin_OrderStatusCancelled(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`UPDATE orders SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/orders/1/status/cancelled", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestAdmin_OrderStatusRefunded(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`UPDATE orders SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/orders/1/status/refunded", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestAdmin_BulkProductCategory(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`UPDATE products SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, result := R("POST", "/api/v1/admin/bulk/product-category", map[string]string{
		"category_id": "2", "ids": "1",
	}, map[string]string{"Authorization": "Bearer " + makeAdminToken()})
	assert.Equal(t, 200, code)
	assert.Equal(t, float64(1), result["updated"])
	expectExpectations(t, mock)
}

func TestAdmin_BulkProductToggle(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`UPDATE products SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("POST", "/api/v1/admin/bulk/product-toggle", map[string]string{
		"status": "active", "ids": "1",
	}, map[string]string{"Authorization": "Bearer " + makeAdminToken()})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestAdmin_DeleteUser(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("DELETE", "/api/v1/admin/users/1", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestAdmin_ListUsers(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users ORDER BY created_at DESC`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(1, "test@test.com", "Test User", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/admin/users", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Len(t, result["users"], 1)
	expectExpectations(t, mock)
}

func TestAdmin_DeleteCommunityPost(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`DELETE FROM community_posts WHERE id = \$1`).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("DELETE", "/api/v1/admin/community/posts/1", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestAdmin_CommunityPostList(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT cp.id, cp.user_id.*FROM community_posts cp.*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "email", "name", "avatar_url", "count_likes", "count_comments", "following"}).
			AddRow(1, 1, "Post 1", mockTime(2024, time.January, 1), "test@test.com", "Test User", "", 0, 0, false))
	code, _, result := R("GET", "/api/v1/admin/community/posts", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Len(t, result["posts"], 1)
	expectExpectations(t, mock)
}

func TestAdmin_CommunityUserList(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))
	mock.ExpectQuery(`SELECT u.id, u.email, u.name.*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "count"}).
			AddRow(1, "test@test.com", "Test User", 10).
			AddRow(2, "user2@test.com", "User 2", 5))
	code, _, result := R("GET", "/api/v1/admin/community/users", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Len(t, result["users"], 2)
	expectExpectations(t, mock)
}

func TestAdmin_GuestOrderCheck(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT o.id, o.status.*FROM orders WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "total_usd", "created_at"}).
			AddRow(1, "completed", "10.00", mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/admin/guest-orders/1", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Equal(t, "completed", result["status"])
	expectExpectations(t, mock)
}

func TestAdmin_ExchangeRateList(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).
		WillReturnRows(sqlmock.NewRows([]string{"chain", "rate", "updated_at"}).
			AddRow("BSC", "350.00", mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/admin/exchange-rates", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Len(t, result["rates"], 1)
	expectExpectations(t, mock)
}

func TestAdmin_ExchangeRateSet(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`INSERT INTO exchange_rates.*`).
		WithArgs("BSC", "350.00").WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/exchange-rates/BSC", map[string]string{
		"rate": "350.00",
	}, map[string]string{"Authorization": "Bearer " + makeAdminToken()})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestAdmin_SetSetting(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`INSERT INTO settings.*`).
		WithArgs("payment_address", "0xTestAddress").WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/settings/payment_address", map[string]string{
		"value": "0xTestAddress",
	}, map[string]string{"Authorization": "Bearer " + makeAdminToken()})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestAdmin_GetSetting(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT value FROM settings WHERE key = \$1`).
		WithArgs("payment_address").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("0xTestAddress"))
	code, _, result := R("GET", "/api/v1/admin/settings/payment_address", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Equal(t, "0xTestAddress", result["value"])
	expectExpectations(t, mock)
}

func TestAdmin_Stats(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COALESCE.*`).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow("100.00"))
	code, _, result := R("GET", "/api/v1/admin/stats", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Equal(t, float64(25), result["users"])
	expectExpectations(t, mock)
}

func TestCart_GetOrCreate(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = \$1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	code, _, result := R("GET", "/api/v1/cart", nil, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	assert.Equal(t, float64(1), result["id"])
	expectExpectations(t, mock)
}

func TestCart_AddItem(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = \$1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO cart_items.*`).
		WithArgs(1, 1, 3).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("POST", "/api/v1/cart/add", map[string]string{
		"product_id": "1", "quantity": "3",
	}, map[string]string{"Authorization": "Bearer " + makeJWT(1, "test@test.com")})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestCart_RemoveItem(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = \$1 AND product_id = \$2`).
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("DELETE", "/api/v1/cart/remove", map[string]string{
		"product_id": "1",
	}, map[string]string{"Authorization": "Bearer " + makeJWT(1, "test@test.com")})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestWishlist_Toggle(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM wishlists WHERE user_id = \$1 AND product_id = \$2\)`).
		WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO wishlists.*`).
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, result := R("POST", "/api/v1/wishlist/toggle", map[string]string{
		"product_id": "1",
	}, map[string]string{"Authorization": "Bearer " + makeJWT(1, "test@test.com")})
	assert.Equal(t, 200, code)
	assert.Contains(t, result, "added")
	expectExpectations(t, mock)
}

func TestWishlist_List(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT w.product_id, p.name, p.slug.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "name", "slug", "price_usd", "image_url"}).
			AddRow(1, "Test Product", "test-product", "10.00", "http://example.com/img.png"))
	code, _, result := R("GET", "/api/v1/wishlist", nil, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	assert.Len(t, result["items"], 1)
	expectExpectations(t, mock)
}

func TestReviews_Create(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`INSERT INTO reviews.*`).
		WithArgs(1, 1, 5, "Great product!").WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("POST", "/api/v1/reviews", map[string]string{
		"product_id": "1", "rating": "5", "comment": "Great product!",
	}, map[string]string{"Authorization": "Bearer " + makeJWT(1, "test@test.com")})
	assert.Equal(t, 201, code)
	expectExpectations(t, mock)
}

func TestReviews_List(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT r.id, r.rating, r.comment.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "rating", "comment", "created_at", "email", "name"}).
			AddRow(1, 1, 5, "Great!", mockTime(2024, time.January, 1), "test@test.com", "Test User"))
	code, _, result := R("GET", "/api/v1/reviews?product_id=1", nil, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, result["reviews"], 1)
	expectExpectations(t, mock)
}

func TestRecentlyViewed_Record(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`INSERT INTO recently_viewed.*`).
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("POST", "/api/v1/recently-viewed", map[string]string{
		"product_id": "1",
	}, map[string]string{"Authorization": "Bearer " + makeJWT(1, "test@test.com")})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestRecentlyViewed_List(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT p.id, p.name, p.slug.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "price_usd", "description"}).
			AddRow(1, "Test Product", "test-product", "10.00", "A test product"))
	code, _, result := R("GET", "/api/v1/recently-viewed", nil, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	assert.Len(t, result["products"], 1)
	expectExpectations(t, mock)
}

func TestComparison_Toggle(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM product_comparison WHERE user_id = \$1 AND product_id = \$2\)`).
		WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO product_comparison.*`).
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("POST", "/api/v1/compare/toggle", map[string]string{
		"product_id": "1",
	}, map[string]string{"Authorization": "Bearer " + makeJWT(1, "test@test.com")})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestComparison_List(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT p.id, p.name, p.slug.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "price_usd", "description"}).
			AddRow(1, "Test Product", "test-product", "10.00", "A test product"))
	code, _, result := R("GET", "/api/v1/compare", nil, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	assert.Len(t, result["products"], 1)
	expectExpectations(t, mock)
}

func TestRecommendations_Get(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT category_id FROM products WHERE id = \$1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))
	mock.ExpectQuery(`SELECT p.id, p.title.*FROM products p.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "price_usd", "description"}).
			AddRow(2, "Related Product", "related-product", "15.00", "Related to test"))
	mock.ExpectQuery(`SELECT p.id, p.name.*FROM products p.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "category_name"}).
			AddRow(1, "Test Product", "test-product", "Icons"))
	code, _, result := R("GET", "/api/v1/recommendations/1", nil, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, result["recommendations"], 1)
	expectExpectations(t, mock)
}

func TestGuestOrder_Create(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`INSERT INTO guest_orders.*`).
		WithArgs("test@example.com", "10.00", "BSC").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT id, email, status.*FROM guest_orders WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "status", "total_usd", "crypto_chain", "created_at"}).
			AddRow(1, "test@example.com", "pending", "10.00", "BSC", mockTime(2024, time.January, 1)))
	code, _, result := R("POST", "/api/v1/guest-orders", map[string]string{
		"email": "test@example.com", "product_ids": "1",
	}, nil)
	assert.Equal(t, 201, code)
	assert.Contains(t, result, "order_id")
	expectExpectations(t, mock)
}

func TestGuestOrder_Check(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, email, status.*FROM guest_orders WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "status", "total_usd", "crypto_chain", "created_at"}).
			AddRow(1, "test@example.com", "completed", "10.00", "BSC", mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/guest-orders/1", nil, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, "completed", result["status"])
	expectExpectations(t, mock)
}

func TestOrderStatusCheck(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, status, total_usd.*FROM orders WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "total_usd", "crypto_chain", "payment_tx_hash", "payment_confirmations", "paid_at"}).
			AddRow(1, "paid", "10.00", "BSC", "0xabc...", 12, mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/orders/1/status/check", nil, map[string]string{
		"Authorization": "Bearer " + makeJWT(1, "test@test.com"),
	})
	assert.Equal(t, 200, code)
	assert.Equal(t, "paid", result["status"])
	expectExpectations(t, mock)
}

func TestHealthCheck(t *testing.T) {
	code, _, result := R("GET", "/api/v1/health", nil, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, "ok", result["status"])
}

func TestProducts_CreateProduct(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM products WHERE slug = \$1`).
		WithArgs("test-product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM categories WHERE id = \$1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO products.*`).
		WithArgs("Test Product", "test-product", "A test product", 1, "10.00", "active").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO product_images.*`).
		WithArgs(1, "http://example.com/img.png", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, result := R("POST", "/api/v1/products", map[string]interface{}{
		"name": "Test Product", "slug": "test-product", "description": "A test product",
		"category_id": 1, "price": "10.00", "stock": 100, "image": "http://example.com/img.png",
	}, map[string]string{"Authorization": "Bearer " + makeAdminToken()})
	assert.Equal(t, 201, code)
	assert.Equal(t, float64(1), result["id"])
	expectExpectations(t, mock)
}

func TestSettings_GetAll(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT key, value, updated_at FROM settings ORDER BY key`).
		WillReturnRows(sqlmock.NewRows([]string{"key", "value", "updated_at"}).
			AddRow("payment_address", "0xTestAddress", mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/admin/settings", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Len(t, result["settings"], 1)
	expectExpectations(t, mock)
}

func TestNewProducts_GetActiveStat(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM products WHERE status = 'active'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM products`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM categories`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	code, _, result := R("GET", "/api/v1/admin/products/stats", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	assert.Equal(t, float64(5), result["active_products"])
	expectExpectations(t, mock)
}

func TestNewProducts_Delete(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectExec(`DELETE FROM product_images WHERE product_id = \$1`).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("DELETE", "/api/v1/admin/products/1", nil, map[string]string{
		"Authorization": "Bearer " + makeAdminToken(),
	})
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}

func TestExchangeRates_GetAll(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).
		WillReturnRows(sqlmock.NewRows([]string{"chain", "rate", "updated_at"}).
			AddRow("BSC", "350.00", mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/exchange-rates", nil, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, result["rates"], 1)
	expectExpectations(t, mock)
}

func TestExchangeRates_GetOne(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = \$1`).
		WithArgs("BSC").WillReturnRows(sqlmock.NewRows([]string{"chain", "rate", "updated_at"}).
			AddRow("BSC", "350.00", mockTime(2024, time.January, 1)))
	code, _, result := R("GET", "/api/v1/exchange-rates/BSC", nil, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, "BSC", result["chain"])
	expectExpectations(t, mock)
}

func TestDBSchema_Migration011Tables(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	for _, table := range []string{"reviews", "cart_items", "wishlists", "recently_viewed", "product_comparison"} {
		mock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = \$1`).
			WithArgs(table).WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow(table))
	}
	code, _, _ := R("GET", "/api/v1/health", nil, nil)
	assert.Equal(t, 200, code)
	expectExpectations(t, mock)
}
