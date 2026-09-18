package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestGetProductReviewsRejectsBadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewReviewHandler(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "productId", Value: "abc"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/products/abc/reviews", nil)
	h.GetProductReviews(c)
	if w.Code != 400 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestGetProductReviewsListsPublicRows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ts := time.Now()
	mock.ExpectQuery("SELECT COUNT").WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT r.id, r.product_id").WithArgs(7, 20, 0).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "product_id", "user_id", "rating", "title", "content", "is_verified_purchase",
			"helpful_count", "is_helpful", "created_at", "updated_at",
			"a", "b", "c", "d",
		}).AddRow(1, 7, 3, 5, "Great", "Loved it", true, 2, false, ts, ts, "", "", "", ""),
	)
	h := NewReviewHandler(db)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "productId", Value: "7"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/products/7/reviews", nil)
	h.GetProductReviews(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"rating":5`) {
		t.Fatalf("body %s", w.Body.String())
	}
}

func TestCreateReviewRejectsOutOfRangeRating(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewReviewHandler(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/reviews", strings.NewReader(`{"product_id":1,"rating":9}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CreateReview(c)
	if w.Code != 400 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}
