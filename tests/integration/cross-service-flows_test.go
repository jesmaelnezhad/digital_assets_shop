package integration_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func stagingBase() string {
	if u := os.Getenv("STAGING_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir"
}

func httpClient() *http.Client {
	return &http.Client{
		Timeout: 25 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func doJSON(t *testing.T, method, path string, body any, headers map[string]string) (int, map[string]any) {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, stagingBase()+path, rdr)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := httpClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return res.StatusCode, out
}

func arr(v any) []any {
	if a, ok := v.([]any); ok {
		return a
	}
	return nil
}

func TestProductToOrderFlow(t *testing.T) {
	t.Run("browse_product_add_to_cart_checkout_order", func(t *testing.T) {
		status, data := doJSON(t, "GET", "/api/v1/products?search=marble&page=1&per_page=5", nil, nil)
		if status != 200 {
			t.Fatalf("products search status %d", status)
		}
		products := arr(data["products"])
		if len(products) == 0 {
			t.Fatal("expected marble search hits")
		}
		status, cats := doJSON(t, "GET", "/api/v1/categories", nil, nil)
		if status != 200 || len(arr(cats["categories"])) < 8 {
			t.Fatalf("categories volume status=%d count=%d", status, len(arr(cats["categories"])))
		}
		status, page2 := doJSON(t, "GET", "/api/v1/products?page=2&per_page=12", nil, nil)
		if status != 200 || len(arr(page2["products"])) == 0 {
			t.Fatal("page 2 of catalog should have products")
		}
		email := "e2e-int-" + time.Now().Format("150405") + "@example.com"
		status, auth := doJSON(t, "POST", "/api/v1/register", map[string]any{
			"email": email, "password": "TestPass123", "name": "Int User",
		}, nil)
		if status != 201 && status != 200 {
			t.Fatalf("register %d %#v", status, auth)
		}
		token, _ := auth["token"].(string)
		if token == "" {
			_, login := doJSON(t, "POST", "/api/v1/login", map[string]any{"email": email, "password": "TestPass123"}, nil)
			token, _ = login["token"].(string)
		}
		if token == "" {
			t.Fatal("no auth token")
		}
		hdr := map[string]string{"Authorization": "Bearer " + token}
		prod := products[0].(map[string]any)
		pid := prod["id"]
		status, _ = doJSON(t, "POST", "/api/v1/cart/items", map[string]any{"product_id": pid, "quantity": 1}, hdr)
		if status != 200 && status != 201 {
			t.Fatalf("add cart %d", status)
		}
		status, cart := doJSON(t, "GET", "/api/v1/cart", nil, hdr)
		if status != 200 || len(arr(cart["items"])) == 0 && len(arr(cart["cart_items"])) == 0 {
			t.Fatalf("cart empty after add: %#v", cart)
		}
	})
}

func TestCouponDiscountFlow(t *testing.T) {
	t.Run("validate_seeded_coupons", func(t *testing.T) {
		for _, code := range []string{"SAVE12", "WELCOME", "MARBLE"} {
			status, data := doJSON(t, "POST", "/api/v1/coupons/validate", map[string]any{
				"code": code, "cart_total": 50,
			}, nil)
			if status != 200 {
				t.Fatalf("coupon %s status %d %#v", code, status, data)
			}
		}
	})
}

func TestCommunityFollowLikeFlow(t *testing.T) {
	t.Run("feed_people_and_auth_post", func(t *testing.T) {
		status, feed := doJSON(t, "GET", "/api/v1/community/posts?page=1&per_page=10", nil, nil)
		if status != 200 {
			t.Fatalf("feed %d", status)
		}
		total, _ := feed["total"].(float64)
		if total <= 10 {
			t.Fatalf("feed volume total=%v", feed["total"])
		}
		status, people := doJSON(t, "GET", "/api/v1/community/users?page=1&per_page=12", nil, nil)
		if status != 200 {
			t.Fatalf("people %d", status)
		}
		ptotal, _ := people["total"].(float64)
		if ptotal <= 12 {
			t.Fatalf("people volume total=%v", people["total"])
		}
		status, login := doJSON(t, "POST", "/api/v1/login", map[string]any{
			"email": "nia@example.com", "password": "nia",
		}, nil)
		if status != 200 {
			t.Fatalf("nia login %d %#v", status, login)
		}
		token, _ := login["token"].(string)
		hdr := map[string]string{"Authorization": "Bearer " + token}
		status, created := doJSON(t, "POST", "/api/v1/community/posts", map[string]any{
			"content": "Integration note " + time.Now().Format(time.RFC3339),
		}, hdr)
		if status != 201 && status != 200 {
			t.Fatalf("create post %d %#v", status, created)
		}
	})
}

func TestBannerSliderContract(t *testing.T) {
	status, data := doJSON(t, "GET", "/api/v1/products?banner=1&per_page=24", nil, nil)
	if status != 200 {
		t.Fatalf("banner list %d", status)
	}
	slides := arr(data["products"])
	if len(slides) < 2 {
		t.Fatalf("banner needs 2+ slides, got %d", len(slides))
	}
}

func TestBundleCatalogVolume(t *testing.T) {
	status, data := doJSON(t, "GET", "/api/v1/bundles", nil, nil)
	if status != 200 {
		t.Fatalf("bundles %d", status)
	}
	if len(arr(data["bundles"])) < 3 {
		t.Fatalf("expected extra bundles, got %d", len(arr(data["bundles"])))
	}
}

func TestGuestCheckoutSurface(t *testing.T) {
	status, _ := doJSON(t, "GET", "/api/v1/products?per_page=1", nil, nil)
	if status != 200 {
		t.Fatalf("products %d", status)
	}
	status, data := doJSON(t, "GET", "/api/v1/cart", nil, nil)
	if status != 401 && status != 200 {
		t.Fatalf("guest cart expected 401 or empty 200, got %d %#v", status, data)
	}
}

func loginDemo(t *testing.T, email, password string) map[string]any {
	t.Helper()
	status, data := doJSON(t, "POST", "/api/v1/login", map[string]any{"email": email, "password": password}, nil)
	if status != 200 {
		t.Fatalf("login %s %d %#v", email, status, data)
	}
	return data
}

func tokenOf(auth map[string]any) string {
	tok, _ := auth["token"].(string)
	return tok
}

func TestAppearancePublicAndProtected(t *testing.T) {
	status, data := doJSON(t, "GET", "/api/v1/products/appearance", nil, nil)
	if status != 200 {
		t.Fatalf("public appearance %d", status)
	}
	for _, k := range []string{"palette", "font", "radius", "density"} {
		if data[k] == nil || data[k] == "" {
			t.Fatalf("missing %s %#v", k, data)
		}
	}
	maya := loginDemo(t, "maya@example.com", "maya")
	status, _ = doJSON(t, "PUT", "/api/v1/products/appearance", map[string]any{
		"palette": "night", "font": "system", "radius": "soft", "density": "comfortable",
	}, map[string]string{"Authorization": "Bearer " + tokenOf(maya)})
	if status != 403 {
		t.Fatalf("customer appearance write %d", status)
	}
	nia := loginDemo(t, "nia@example.com", "nia")
	cur := map[string]any{}
	for k, v := range data {
		cur[k] = v
	}
	status, saved := doJSON(t, "PUT", "/api/v1/products/appearance", cur, map[string]string{
		"Authorization": "Bearer " + tokenOf(nia),
	})
	if status != 200 {
		t.Fatalf("admin appearance write %d %#v", status, saved)
	}
}

func TestRBACLoginRoles(t *testing.T) {
	nia := loginDemo(t, "nia@example.com", "nia")
	user, _ := nia["user"].(map[string]any)
	if user["role"] != "admin" {
		t.Fatalf("nia role %#v", user)
	}
	leo := loginDemo(t, "leo@example.com", "leo")
	luser, _ := leo["user"].(map[string]any)
	if luser["role"] != "staff" {
		t.Fatalf("leo role %#v", luser)
	}
	status, _ := doJSON(t, "GET", "/api/v1/admin/stats", nil, map[string]string{
		"Authorization": "Bearer " + tokenOf(leo),
	})
	if status != 403 {
		t.Fatalf("leo stats %d", status)
	}
	status, _ = doJSON(t, "GET", "/api/v1/admin/products", nil, map[string]string{
		"Authorization": "Bearer " + tokenOf(leo),
	})
	if status != 200 {
		t.Fatalf("leo products %d", status)
	}
	maya := loginDemo(t, "maya@example.com", "maya")
	status, _ = doJSON(t, "GET", "/api/v1/admin/stats", nil, map[string]string{
		"Authorization": "Bearer " + tokenOf(maya),
	})
	if status != 403 {
		t.Fatalf("maya stats %d", status)
	}
}

func TestAccessRequiresOperatorToken(t *testing.T) {
	nia := loginDemo(t, "nia@example.com", "nia")
	hdr := map[string]string{"Authorization": "Bearer " + tokenOf(nia)}
	status, users := doJSON(t, "GET", "/api/v1/admin/users", nil, hdr)
	if status != 200 {
		t.Fatalf("users %d", status)
	}
	var leoID any
	for _, row := range arr(users["users"]) {
		u := row.(map[string]any)
		if u["email"] == "leo@example.com" {
			leoID = u["id"]
			if u["role"] != "staff" {
				t.Fatalf("leo role in list %#v", u)
			}
		}
	}
	if leoID == nil {
		t.Fatal("leo missing")
	}
	path := "/api/v1/admin/users/" + jsonNumber(leoID) + "/access"
	status, _ = doJSON(t, "PUT", path, map[string]any{
		"role": "staff", "staff_tabs": []string{"Products", "Banner", "Orders", "Community"},
	}, hdr)
	if status != 403 {
		t.Fatalf("access without operator %d", status)
	}
	op := os.Getenv("ADMIN_TOKEN")
	if op == "" {
		op = "admin_secret_staging_2026"
	}
	hdr["X-Admin-Token"] = op
	status, out := doJSON(t, "PUT", path, map[string]any{
		"role": "staff", "staff_tabs": []string{"Products", "Banner", "Orders", "Community"},
	}, hdr)
	if status != 200 {
		t.Fatalf("access with operator %d %#v", status, out)
	}
}

func TestRecommendationsMissingProduct(t *testing.T) {
	status, data := doJSON(t, "GET", "/api/v1/recommendations/999999999", nil, nil)
	if status != 404 {
		t.Fatalf("missing product recs %d %#v", status, data)
	}
}

func TestOrderPipelineOpsDesk(t *testing.T) {
	op := os.Getenv("ADMIN_TOKEN")
	if op == "" {
		op = "admin_secret_staging_2026"
	}
	hdr := map[string]string{"Authorization": "Bearer " + op}
	status, steps := doJSON(t, "GET", "/api/v1/admin/order-steps", nil, hdr)
	if status != 200 {
		t.Fatalf("steps %d %#v", status, steps)
	}
	slugs := map[string]bool{}
	for _, row := range arr(steps["steps"]) {
		s := row.(map[string]any)
		slug, _ := s["slug"].(string)
		slugs[slug] = true
	}
	for _, want := range []string{"created", "awaiting_payment", "paid", "preparation", "delivered"} {
		if !slugs[want] {
			t.Fatalf("missing step %s %#v", want, steps["steps"])
		}
	}
	status, listed := doJSON(t, "GET", "/api/v1/admin/orders?status=paid", nil, hdr)
	if status != 200 {
		t.Fatalf("orders filter %d %#v", status, listed)
	}
	if arr(listed["by_step"]) == nil {
		t.Fatalf("missing by_step %#v", listed)
	}
	for _, row := range arr(listed["orders"]) {
		o := row.(map[string]any)
		if o["status"] != "paid" {
			t.Fatalf("filter leaked %#v", o)
		}
	}
	status, stats := doJSON(t, "GET", "/api/v1/admin/stats", nil, hdr)
	if status != 200 {
		t.Fatalf("stats %d %#v", status, stats)
	}
	if arr(stats["order_by_step"]) == nil {
		t.Fatalf("stats missing order_by_step %#v", stats)
	}
	leo := loginDemo(t, "leo@example.com", "leo")
	status, _ = doJSON(t, "GET", "/api/v1/admin/order-steps", nil, map[string]string{
		"Authorization": "Bearer " + tokenOf(leo),
	})
	if status != 200 {
		t.Fatalf("leo GET steps %d", status)
	}
	status, _ = doJSON(t, "POST", "/api/v1/admin/order-steps", map[string]any{"label": "NoStaffWrite"}, map[string]string{
		"Authorization": "Bearer " + tokenOf(leo),
	})
	if status != 403 {
		t.Fatalf("leo POST steps %d", status)
	}
}

func jsonNumber(v any) string {
	switch n := v.(type) {
	case float64:
		return strconv.Itoa(int(n))
	case int:
		return strconv.Itoa(n)
	case json.Number:
		return n.String()
	default:
		return fmt.Sprint(v)
	}
}
