# Pawradise — TDD Unit & Integration Test Suite (Microservices)

> TDD test suite for the Pawradise microservice/microfrontend architecture.
> Tests are designed to FAIL until the corresponding feature is implemented.
> Based on MICROSERVICE-ARCHITECTURE.md v1.0 and PRODUCT-SPEC.md v1.1.

---

## File Inventory

```
tests/
├── e2e/
│   └── e2e-suite.js              ← 104 E2E tests (existing)
├── unit/
│   ├── identity-service/
│   │   └── auth_test.go          ← identity: register, login, logout, profiles, JWT, password
│   ├── product-service/
│   │   └── product_test.go       ← products: CRUD, search, filter, sort, tiers, images, bundles, previews, pinning, recommendations
│   ├── commerce-service/
│   │   └── commerce_test.go      ← commerce: orders, cart, wishlist, coupons, recently-viewed, comparison, guest orders, discount calc
│   ├── community-service/
│   │   └── community_test.go     ← community: posts, comments, likes, follows, public profiles
│   ├── review-service/
│   │   └── review_test.go        ← reviews: rating creation, aggregation, verified purchase check
│   ├── payment-service/
│   │   └── payment_test.go       ← payments: payment details, status, confirm, exchange rates, crypto calc
│   ├── admin-service/
│   │   └── admin_test.go         ← admin: stats, users, products bulk, orders, community moderation, referrals, settings, email export
│   └── media-service/
│       └── media_test.go         ← media: upload, download, preview gen, thumbnail, access verify, file validation
├── integration/
│   └── cross-service-flows_test.go ← cross-service: product→order, coupon, referral commission, payment→download, user deletion cascade, guest checkout, bundle, PWYW, review after purchase
└── README.md                     ← THIS FILE
```

---

## Unit Tests (Go + sqlmock)

All unit tests use `sqlmock` for database mocking. No real DB connections.
Each test is independent, idempotent, and isolated.

### Identity Service: `tests/unit/identity-service/auth_test.go`

| Test | Status | Doc |
|------|--------|-----|
| `TestAuthService_Register` | FAIL (TDD) | [auth_test.go:24] |
| `TestAuthService_Login` | FAIL (TDD) | [auth_test.go:51] |
| `TestAuthService_ValidateToken` | FAIL (TDD) | [auth_test.go:78] |
| `TestAuthService_Logout` | FAIL (TDD) | [auth_test.go:91] |
| `TestAuthService_ResetPassword` | FAIL (TDD) | [auth_test.go:103] |
| `TestAuthService_DeleteUser` | FAIL (TDD) | [auth_test.go:117] |
| `TestAuthService_GetProfile` | FAIL (TDD) | [auth_test.go:133] |
| `TestAuthService_UpdateProfile` | FAIL (TDD) | [auth_test.go:147] |
| `TestJWT_Generation` | FAIL (TDD) | [auth_test.go:162] |
| `TestJWT_Validation` | FAIL (TDD) | [auth_test.go:176] |
| `TestPassword_Hashing` | FAIL (TDD) | [auth_test.go:190] |
| `TestPassword_Verification` | FAIL (TDD) | [auth_test.go:200] |

DB interactions mocked:
- SELECT user by email → check for duplicate
- INSERT user → create account
- INSERT user_profiles → create profile
- SELECT user by ID → get profile
- SELECT profile by user_id → get extended data
- UPDATE user → update profile fields
- UPDATE user_profiles → update bio, wallet
- INSERT invalidated_tokens → logout
- SELECT count from invalidated_tokens → validate token

### Product Service: `tests/unit/product-service/product_test.go` (PENDING — awaiting subagent completion)

| Test | Status | Doc |
|------|--------|-----|
| `TestProductService_GetProducts` | FAIL (TDD) | List with search, filter, sort, pagination |
| `TestProductService_GetProductBySlug` | FAIL (TDD) | Get single product by slug |
| `TestProductService_CreateProduct` | FAIL (TDD) | Create product |
| `TestProductService_UpdateProduct` | FAIL (TDD) | Update product |
| `TestProductService_DeleteProduct` | FAIL (TDD) | Delete product |
| `TestProductService_GetCategories` | FAIL (TDD) | List categories |
| `TestProductService_AddProductTier` | FAIL (TDD) | Add tier to product |
| `TestProductService_UpdateProductTier` | FAIL (TDD) | Update tier |
| `TestProductService_DeleteProductTier` | FAIL (TDD) | Delete tier |
| `TestProductService_AddProductImage` | FAIL (TDD) | Add image to gallery |
| `TestProductService_DeleteProductImage` | FAIL (TDD) | Delete image |
| `TestProductService_GeneratePreviews` | FAIL (TDD) | Trigger preview generation |
| `TestProductService_PinProduct` | FAIL (TDD) | Pin to top |
| `TestProductService_GetBundles` | FAIL (TDD) | List bundles |
| `TestProductService_GetBundle` | FAIL (TDD) | Get bundle detail |
| `TestProductService_CreateBundle` | FAIL (TDD) | Create bundle |
| `TestProductService_GetRecommendations` | FAIL (TDD) | Related products by category |
| `TestSearchAndFiltering` | FAIL (TDD) | Search by text, filter by category/price/file_type/rating |

### Commerce Service: `tests/unit/commerce-service/commerce_test.go` (COMPLETE — written by subagent)

| Test | Status | Doc |
|------|--------|-----|
| `TestOrderService_CreateOrder` | FAIL (TDD) | Create from cart |
| `TestOrderService_GetOrders` | FAIL (TDD) | List user orders |
| `TestOrderService_GetOrder` | FAIL (TDD) | Get order detail |
| `TestOrderService_UpdateOrderStatus` | FAIL (TDD) | Status workflow (6 transitions) |
| `TestOrderService_CreateGuestOrder` | FAIL (TDD) | Guest checkout |
| `TestOrderService_GetGuestOrder` | FAIL (TDD) | Check guest order |
| `TestCartService_GetCart` | FAIL (TDD) | Get/create cart |
| `TestCartService_AddItem` | FAIL (TDD) | Add to cart |
| `TestCartService_RemoveItem` | FAIL (TDD) | Remove from cart |
| `TestWishlistService_GetWishlist` | FAIL (TDD) | List wishlist |
| `TestWishlistService_ToggleWishlist` | FAIL (TDD) | Add/remove from wishlist |
| `TestCouponService_ValidateCoupon` | FAIL (TDD) | Validate code |
| `TestRecentlyViewedService_RecordView` | FAIL (TDD) | Record product view |
| `TestRecentlyViewedService_GetRecentlyViewed` | FAIL (TDD) | List recent |
| `TestCompareService_ToggleCompare` | FAIL (TDD) | Add/remove comparison |
| `TestCompareService_GetCompare` | FAIL (TDD) | List compared |
| `TestDiscountCalculation` | FAIL (TDD) | Local discount calc validation |

### Community Service: `tests/unit/community-service/community_test.go` (PENDING — awaiting subagent completion)

| Test | Status | Doc |
|------|--------|-----|
| `TestCommunityService_GetPosts` | FAIL (TDD) | Feed with pagination |
| `TestCommunityService_CreatePost` | FAIL (TDD) | Create post |
| `TestCommunityService_GetPost` | FAIL (TDD) | Single post with comments |
| `TestCommunityService_DeletePost` | FAIL (TDD) | Delete post |
| `TestCommunityService_LikePost` | FAIL (TDD) | Like toggle |
| `TestCommunityService_AddComment` | FAIL (TDD) | Add comment |
| `TestCommunityService_DeleteComment` | FAIL (TDD) | Delete comment |
| `TestCommunityService_FollowUser` | FAIL (TDD) | Follow user |
| `TestCommunityService_UnfollowUser` | FAIL (TDD) | Unfollow user |
| `TestCommunityService_GetPublicProfile` | FAIL (TDD) | Get public profile |

### Review Service: `tests/unit/review-service/review_test.go` (PENDING — awaiting subagent completion)

| Test | Status | Doc |
|------|--------|-----|
| `TestReviewService_CreateRating` | FAIL (TDD) | Rate product (verified purchase) |
| `TestReviewService_GetProductRatings` | FAIL (TDD) | Get ratings for product |
| `TestReviewService_GetAverageRating` | FAIL (TDD) | Calculate average |
| `TestReviewService_VerifiedPurchaseCheck` | FAIL (TDD) | Only purchasers can rate |

### Payment Service: `tests/unit/payment-service/payment_test.go` (PENDING — awaiting subagent completion)

| Test | Status | Doc |
|------|--------|-----|
| `TestPaymentService_GetPaymentDetails` | FAIL (TDD) | Get crypto payment info |
| `TestPaymentService_CheckPaymentStatus` | FAIL (TDD) | Check payment status |
| `TestPaymentService_ConfirmPayment` | FAIL (TDD) | Confirm payment |
| `TestPaymentService_GetExchangeRates` | FAIL (TDD) | List all rates |
| `TestPaymentService_GetExchangeRate` | FAIL (TDD) | Get single rate |
| `TestPaymentService_SetExchangeRate` | FAIL (TDD) | Set rate |
| `TestPaymentService_DeleteExchangeRate` | FAIL (TDD) | Delete rate |
| `TestPaymentService_CalculateCryptoAmount` | FAIL (TDD) | USD to crypto conversion |

### Admin Service: `tests/unit/admin-service/admin_test.go` (PENDING — awaiting subagent completion)

| Test | Status | Doc |
|------|--------|-----|
| `TestAdminService_GetStats` | FAIL (TDD) | Dashboard statistics |
| `TestAdminService_ListUsers` | FAIL (TDD) | List all users |
| `TestAdminService_DeleteUser` | FAIL (TDD) | Delete user |
| `TestAdminService_ResetPassword` | FAIL (TDD) | Reset user password |
| `TestAdminService_BulkUpdateProducts` | FAIL (TDD) | Bulk status/category update |
| `TestAdminService_ListAllOrders` | FAIL (TDD) | List all orders |
| `TestAdminService_ListCommunityPosts` | FAIL (TDD) | Moderation list |
| `TestAdminService_DeleteCommunityPost` | FAIL (TDD) | Delete post |
| `TestAdminService_ListReferrals` | FAIL (TDD) | All referrals |
| `TestAdminService_ExportEmails` | FAIL (TDD) | Export buyer emails |
| `TestAdminService_GetSettings` | FAIL (TDD) | List settings |
| `TestAdminService_SetSetting` | FAIL (TDD) | Update setting |

### Media Service: `tests/unit/media-service/media_test.go` (PENDING — awaiting subagent completion)

| Test | Status | Doc |
|------|--------|-----|
| `TestMediaService_Upload` | FAIL (TDD) | Upload file |
| `TestMediaService_Download` | FAIL (TDD) | Serve file (auth verified) |
| `TestMediaService_GeneratePreview` | FAIL (TDD) | Create watermarked preview |
| `TestMediaService_CreateThumbnail` | FAIL (TDD) | Create thumbnail |
| `TestMediaService_DeleteFile` | FAIL (TDD) | Delete file |
| `TestMediaService_VerifyAccess` | FAIL (TDD) | Verify download access |
| `TestMediaService_GetFileMetadata` | FAIL (TDD) | Get file info |
| `TestMediaService_ValidateFileType` | FAIL (TDD) | Validate file type |

---

## Integration Tests (Go + testcontainers / SQLite in-memory DB)

Integration tests verify cross-service workflows. They use a real database
(or SQLite in-memory) to test data flow across service boundaries.

### Cross-Service Flows: `tests/integration/cross-service-flows_test.go` (PENDING — awaiting subagent completion)

| Test | Status | Doc |
|------|--------|-----|
| `TestProductToOrderFlow` | FAIL (TDD) | Browse → cart → checkout → order |
| `TestCouponDiscountFlow` | FAIL (TDD) | Create coupon → apply → verify discount |
| `TestReferralCommissionFlow` | FAIL (TDD) | Refer → purchase → commission earned |
| `TestPaymentFlow` | FAIL (TDD) | Order → details → confirm → paid → download |
| `TestUserDeletionCascade` | FAIL (TDD) | Delete user → anonymize posts → remove cart/wishlist |
| `TestGuestCheckoutFlow` | FAIL (TDD) | Guest email → order → check status |
| `TestBundlePurchaseFlow` | FAIL (TDD) | Buy bundle → all component products accessible |
| `TestPWYWFlow` | FAIL (TDD) | PWYW enabled → user enters price → purchase at custom price |
| `TestReviewAfterPurchase` | FAIL (TDD) | Purchase → rate → rating appears |
| `TestAdminModerationFlow` | FAIL (TDD) | Admin deletes post → removed from feed |
| `TestExchangeRateUpdate` | FAIL (TDD) | Admin sets rate → used in payment calc |
| `TestCartToOrderWithMultipleItems` | FAIL (TDD) | Multiple items → single order |
| `TestWishlistToCartFlow` | FAIL (TDD) | Wishlist → cart → checkout |
| `TestProductSearchFlow` | FAIL (TDD) | Search → filter → sort → expected results |

---

## Running the Tests

### Run all unit tests

```bash
cd /root/project/backend
go test ./tests/unit/... -v -count=1
```

### Run a single service's tests

```bash
cd /root/project/backend
go test ./tests/unit/identity-service/... -v
go test ./tests/unit/product-service/... -v
go test ./tests/unit/commerce-service/... -v
go test ./tests/unit/community-service/... -v
go test ./tests/unit/review-service/... -v
go test ./tests/unit/payment-service/... -v
go test ./tests/unit/admin-service/... -v
go test ./tests/unit/media-service/... -v
```

### Run integration tests

```bash
cd /root/project/backend
go test ./tests/integration/... -v -count=1
```

### Measure coverage

```bash
go test ./tests/... -coverprofile=/tmp/cover.out
go tool cover -func=/tmp/cover.out | grep total
```

Target: ≥70% line coverage across all services.

---

## TDD Workflow

1. **Run tests** → all FAIL (expected)
2. **Implement feature** in the corresponding microservice
3. **Run tests again** → some PASS, some FAIL
4. **Fix failures** until all tests for that service PASS
5. **Move to next service**
6. **When all unit tests PASS** → run integration tests
7. **Fix integration failures** until all PASS
8. **Finally run E2E suite** via `npm run test:e2e`

---

## Test Status Legend

| Status | Meaning |
|--------|---------|
| FAIL (TDD) | Test written, implementation not yet built. Expected to fail. |
| PASS | Test passes — feature is implemented. |
| PENDING | Subagent still writing test file. Check after completion. |

---

*TDD suite version: 1.0 — 2026-09-06*
*Status: Tests written, features not yet implemented. All tests FAIL by design.*
