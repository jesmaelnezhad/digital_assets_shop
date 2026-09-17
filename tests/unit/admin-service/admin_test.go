package admin_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	return db, mock
}

func TestAdminService_GetStats(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.GetStats
}

func TestAdminService_ListUsers(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.ListUsers
}

func TestAdminService_DeleteUser(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.DeleteUser
}

func TestAdminService_ResetPassword(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.ResetPassword
}

func TestAdminService_ListAllOrders(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.ListAllOrders
}

func TestAdminService_ListCommunityPosts(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.ListCommunityPosts
}

func TestAdminService_DeleteCommunityPost(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.DeleteCommunityPost
}

func TestAdminService_ListReferrals(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.ListReferrals
}

func TestAdminService_ExportEmails(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.ExportEmails
}

func TestAdminService_GetSettings(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.GetSettings
}

func TestAdminService_SetSetting(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()
	// TODO: Call admin.Service.SetSetting
}

var _ = sql.ErrNoRows
