package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestUpdateCartItem(t *testing.T) {
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
}
