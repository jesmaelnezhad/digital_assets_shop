package database

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDB_ConnStringFromEnv(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://test:testpass@localhost:5432/testdb?sslmode=disable")
	defer os.Unsetenv("DATABASE_URL")

	db, err := InitDB()
	// This will fail to connect (no DB), but it should use the env var
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestInitDB_DefaultConnString(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	db, err := InitDB()
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestCreateTables(t *testing.T) {
	// Relies on a test DB being available
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer db.Close()

	err = CreateTables(db)
	assert.NoError(t, err)
}

func TestClose_NilDB(t *testing.T) {
	// Close on nil DB should not panic
	Close()
}
