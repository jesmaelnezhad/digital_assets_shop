package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestListProductsFilteredSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &Handlers{db: db}

	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT COUNT").WithArgs("%marble%").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT p.id").WithArgs("%marble%", 12, 0).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "price_usd", "status", "created_at", "updated_at",
			"category_id", "image_url", "stock_count", "pinned", "sort_order", "digital_formats",
			"tags", "is_pwyw", "pwyw_min_price", "pinned_at", "banner_sort",
		}).AddRow(1, "Marble Kit", "marble-kit", "desc", 12.0, "active", ts, ts,
			1, "", 0, true, 0, "PNG", "tag", false, 0, nil, 1),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/products?search=marble", nil)
	h.ListProducts(c)
	if w.Code != 200 {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetBundleByIDIncludesItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &Handlers{db: db}
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT id, title, slug").WithArgs(1).WillReturnRows(
		sqlmock.NewRows([]string{"id", "title", "slug", "description", "price_usd", "status", "sort_order", "created_at", "updated_at"}).
			AddRow(1, "Studio Kit", "studio-kit", "stills", 72.0, "active", 1, ts, ts),
	)
	mock.ExpectQuery("SELECT product_id FROM bundle_items").WithArgs(1).WillReturnRows(
		sqlmock.NewRows([]string{"product_id"}).AddRow(301).AddRow(303).AddRow(311),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/bundles/1", nil)
	h.GetBundle(c)
	if w.Code != 200 {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.Body.String(), `"items":"301,303,311"`) {
		t.Fatalf("missing items csv: %s", w.Body.String())
	}
}

func TestSetBannerOrdersSlides(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &Handlers{db: db}
	mock.ExpectExec("UPDATE products SET banner_sort = 0").WillReturnResult(sqlmock.NewResult(0, 12))
	mock.ExpectExec("UPDATE products SET banner_sort").WithArgs(1, 301).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE products SET banner_sort").WithArgs(2, 305).WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/products/banner", strings.NewReader(`{"product_ids":[301,305]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.SetBanner(c)
	if w.Code != 200 {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListBannerOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &Handlers{db: db}
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT p.id").WithArgs(12, 0).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "price_usd", "status", "created_at", "updated_at",
			"category_id", "image_url", "stock_count", "pinned", "sort_order", "digital_formats",
			"tags", "is_pwyw", "pwyw_min_price", "pinned_at", "banner_sort",
		}).AddRow(301, "Lunar Clay Characters", "lunar-clay-characters", "d", 48.0, "active", ts, ts,
			1, "/assets/catalog/p01.jpg", 0, true, 1, "FBX", "clay", false, 0, nil, 1),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/products?banner=1", nil)
	h.ListProducts(c)
	if w.Code != 200 {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "lunar-clay-characters") {
		t.Fatalf("missing banner product: %s", w.Body.String())
	}
}

func TestDeleteCategoryInUse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &Handlers{db: db}
	mock.ExpectQuery("SELECT COUNT").WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT COUNT").WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/categories/3", nil)
	h.DeleteCategory(c)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
	}
}

func TestGetAppearance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &Handlers{db: db}
	mock.ExpectQuery("SELECT palette").WillReturnRows(
		sqlmock.NewRows([]string{"palette", "font", "radius", "density", "icons", "contrast", "grain", "glow", "motion", "tracking"}).
			AddRow("night", "serif", "round", "roomy", "line", "standard", "light", "halo", "gentle", "normal"),
	)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/products/appearance", nil)
	h.GetAppearance(c)
	if w.Code != 200 {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"palette":"night"`) {
		t.Fatalf("body %s", w.Body.String())
	}
}

func TestSetAppearanceNormalizesUnknownPalette(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &Handlers{db: db}
	mock.ExpectExec("INSERT INTO site_appearance").WithArgs("clay", "system", "soft", "comfortable", "line", "standard", "light", "halo", "gentle", "normal").
		WillReturnResult(sqlmock.NewResult(1, 1))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/products/appearance", strings.NewReader(`{"palette":"neon","font":"comic"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.SetAppearance(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"palette":"clay"`) {
		t.Fatalf("expected fallback clay: %s", w.Body.String())
	}
}

func TestGetRecommendationsMissingProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &Handlers{db: db}
	mock.ExpectQuery("SELECT category_id FROM products").WithArgs(999999999).WillReturnError(sql.ErrNoRows)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "productId", Value: "999999999"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/999999999", nil)
	h.GetRecommendations(c)
	if w.Code != 404 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestNormalizeAppearanceAllowsConfiguredPalettes(t *testing.T) {
	got := normalizeAppearance(siteAppearance{Palette: "MARBLE", Font: "Mono", Radius: "round", Density: "roomy", Icons: "BOLD", Contrast: "PUNCHY", Grain: "HEAVY", Glow: "BLOOM", Motion: "STILL", Tracking: "WIDE"})
	if got.Palette != "marble" || got.Font != "mono" || got.Radius != "round" || got.Density != "roomy" {
		t.Fatalf("%+v", got)
	}
	if got.Icons != "bold" || got.Contrast != "punchy" || got.Grain != "heavy" || got.Glow != "bloom" || got.Motion != "still" || got.Tracking != "wide" {
		t.Fatalf("dims %+v", got)
	}
	paper := normalizeAppearance(siteAppearance{Palette: "paper"})
	if paper.Palette != "paper" {
		t.Fatalf("paper %+v", paper)
	}
}
