package handlers

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestApplyCouponDiscount(t *testing.T) {
	if applyCouponDiscount("percentage", "100", 40) != 40 {
		t.Fatal("100 percent")
	}
	if applyCouponDiscount("percentage", "12", 100) != 12 {
		t.Fatal("12 percent")
	}
	if applyCouponDiscount("fixed", "15", 10) != 10 {
		t.Fatal("fixed capped")
	}
}

func TestZeroDueIsPaid(t *testing.T) {
	if applyCouponDiscount("percentage", "100", 48) < 48 {
		t.Fatal("full off")
	}
}

func TestUpdateCartItemQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewCommerceHandler(db)
	mock.ExpectExec("UPDATE cart_items").WithArgs(3, 9, 1).WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/cart/items/9", bytes.NewBufferString(`{"quantity":3}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.UpdateCartItem(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestToggleWishlistBodyProductID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewCommerceHandler(db)
	mock.ExpectQuery("SELECT 1 FROM wishlist_items").WithArgs(1, 42).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	mock.ExpectExec("INSERT INTO wishlist_items").WithArgs(1, 42).WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/wishlist/toggle", bytes.NewBufferString(`{"product_id":42}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.ToggleWishlist(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"added":true`)) {
		t.Fatalf("want added true: %s", w.Body.String())
	}
}

func TestToggleWishlistRemovesExisting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewCommerceHandler(db)
	mock.ExpectQuery("SELECT 1 FROM wishlist_items").WithArgs(1, 42).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("DELETE FROM wishlist_items").WithArgs(1, 42).WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/wishlist/toggle", bytes.NewBufferString(`{"product_id":42}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.ToggleWishlist(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"added":false`)) {
		t.Fatalf("want added false: %s", w.Body.String())
	}
}

func TestGetWishlistReturnsProductID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewCommerceHandler(db)
	mock.ExpectQuery("SELECT id, product_id FROM wishlist_items").WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_id"}).AddRow(521, 42))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/wishlist", nil)
	h.GetWishlist(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"product_id":42`)) {
		t.Fatalf("want product_id 42: %s", w.Body.String())
	}
}

func TestGetWishlistQueryError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewCommerceHandler(db)
	mock.ExpectQuery("SELECT id, product_id FROM wishlist_items").WithArgs(1).WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/wishlist", nil)
	h.GetWishlist(c)
	if w.Code != 500 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestToggleCompareCapsAtFour(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewCommerceHandler(db)
	mock.ExpectQuery("SELECT 1 FROM product_comparisons").WithArgs(1, 99).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	mock.ExpectQuery("SELECT COUNT").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/compare/toggle", bytes.NewBufferString(`{"product_id":99}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.ToggleCompare(c)
	if w.Code != 400 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}
