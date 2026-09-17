package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

	mock.ExpectQuery("SELECT COUNT").WithArgs("%marble%").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT p.id").WithArgs("%marble%", 12, 0).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "title", "slug", "description", "price_usd", "status", "created_at", "updated_at",
			"category_id", "image_url", "stock_count", "pinned", "sort_order", "digital_formats",
			"tags", "is_pwyw", "pwyw_min_price", "pinned_at",
		}).AddRow(1, "Marble Kit", "marble-kit", "desc", 12.0, "active", "2026-01-01", "2026-01-01",
			1, "", 0, true, 0, "PNG", "tag", false, 0, nil),
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
