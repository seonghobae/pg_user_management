// +build integration

package test

import (
	"os"
	"testing"

	"github.com/seonghobae/pg_user_management/internal/auth"
	"github.com/seonghobae/pg_user_management/internal/database"
	"github.com/seonghobae/pg_user_management/internal/permission"
	"github.com/seonghobae/pg_user_management/internal/user"
	"github.com/seonghobae/pg_user_management/pkg/config"
)

// Integration tests require a real PostgreSQL database
// Run with: go test -tags=integration ./test/

func getTestDB(t *testing.T) *database.DB {
	// Set default test environment if not set
	if os.Getenv("PGHOST") == "" {
		os.Setenv("PGHOST", "localhost")
	}
	if os.Getenv("PGPORT") == "" {
		os.Setenv("PGPORT", "5432")
	}
	if os.Getenv("PGUSER") == "" {
		os.Setenv("PGUSER", "postgres")
	}
	if os.Getenv("PGDATABASE") == "" {
		os.Setenv("PGDATABASE", "postgres")
	}
	if os.Getenv("PGSSLMODE") == "" {
		os.Setenv("PGSSLMODE", "disable")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
		return nil
	}

	db, err := database.Connect(cfg.ConnectionString())
	if err != nil {
		t.Skipf("Skipping integration test - cannot connect to database: %v", err)
		return nil
	}

	return db
}

func TestIntegration_UserLifecycle(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	userMgr := user.NewManager(db.DB)
	testUsername := "test_integration_user"

	// Clean up any existing test user
	userMgr.DeleteUser(testUsername)

	// Test: Create user
	t.Run("CreateUser", func(t *testing.T) {
		opts := user.UserOptions{
			Username:    testUsername,
			Password:    "testpass123",
			IsSuperuser: false,
			CanLogin:    true,
			AuthMethod:  auth.ScramSHA256,
		}

		err := userMgr.CreateUser(opts)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	})

	// Test: List users (should include our test user)
	t.Run("ListUsers", func(t *testing.T) {
		users, err := userMgr.ListUsers()
		if err != nil {
			t.Fatalf("Failed to list users: %v", err)
		}

		found := false
		for _, u := range users {
			if u["username"] == testUsername {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Created user %s not found in user list", testUsername)
		}
	})

	// Test: Modify user
	t.Run("ModifyUser", func(t *testing.T) {
		opts := user.UserOptions{
			Username:    testUsername,
			Password:    "newpass456",
			IsSuperuser: false,
			CanLogin:    true,
			AuthMethod:  auth.ScramSHA256,
		}

		err := userMgr.ModifyUser(opts)
		if err != nil {
			t.Fatalf("Failed to modify user: %v", err)
		}
	})

	// Clean up
	t.Run("DeleteUser", func(t *testing.T) {
		err := userMgr.DeleteUser(testUsername)
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}
	})
}

func TestIntegration_PermissionManagement(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	userMgr := user.NewManager(db.DB)
	permMgr := permission.NewManager(db.DB)
	testUsername := "test_perm_user"

	// Create test user
	userMgr.DeleteUser(testUsername)
	opts := user.UserOptions{
		Username:    testUsername,
		Password:    "testpass123",
		IsSuperuser: false,
		CanLogin:    true,
		AuthMethod:  auth.ScramSHA256,
	}
	if err := userMgr.CreateUser(opts); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer userMgr.DeleteUser(testUsername)

	// Test: Grant schema usage
	t.Run("GrantSchemaUsage", func(t *testing.T) {
		err := permMgr.GrantSchemaUsage(testUsername, "public")
		if err != nil {
			t.Fatalf("Failed to grant schema usage: %v", err)
		}
	})

	// Test: Grant function execute privileges
	t.Run("GrantFunctionPrivileges", func(t *testing.T) {
		err := permMgr.GrantAllFunctionsInSchema(testUsername, "public")
		if err != nil {
			t.Fatalf("Failed to grant function privileges: %v", err)
		}
	})

	// Test: Grant table privileges
	t.Run("GrantTablePrivileges", func(t *testing.T) {
		grantOpts := permission.GrantOptions{
			Username:   testUsername,
			Schema:     "public",
			Privileges: []permission.Privilege{permission.SELECT},
		}

		err := permMgr.GrantTablePrivileges(grantOpts)
		if err != nil {
			t.Fatalf("Failed to grant table privileges: %v", err)
		}
	})

	// Test: List user privileges
	t.Run("ListUserPrivileges", func(t *testing.T) {
		privileges, err := permMgr.ListUserPrivileges(testUsername)
		if err != nil {
			t.Fatalf("Failed to list user privileges: %v", err)
		}

		if len(privileges) == 0 {
			t.Log("Warning: No privileges found (this might be expected if no tables exist)")
		}
	})
}
