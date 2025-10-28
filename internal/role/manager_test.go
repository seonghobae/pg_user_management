package role

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple identifier",
			input: "rolename",
			want:  `"rolename"`,
		},
		{
			name:  "Identifier with spaces",
			input: "role name",
			want:  `"role name"`,
		},
		{
			name:  "Identifier with quotes",
			input: `role"name`,
			want:  `"role""name"`,
		},
		{
			name:  "Identifier with multiple quotes",
			input: `ro"le"na"me`,
			want:  `"ro""le""na""me"`,
		},
		{
			name:  "SQL injection attempt",
			input: `foo"; DROP ROLE postgres; --`,
			want:  `"foo""; DROP ROLE postgres; --"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quoteIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("quoteIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuoteIdentifier_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("quoteIdentifier() did not panic on empty string")
		}
	}()
	quoteIdentifier("")
}

func TestNewManager(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)
	if manager == nil {
		t.Error("NewManager() returned nil")
	}
	if manager.db != db {
		t.Error("NewManager() did not set db correctly")
	}
}

func TestManager_CreateRole(t *testing.T) {
	tests := []struct {
		name      string
		opts      RoleOptions
		wantQuery string
		mockError error
		wantError bool
		setupMock func(sqlmock.Sqlmock)
	}{
		{
			name: "Create NOLOGIN role",
			opts: RoleOptions{
				RoleName: "app_readonly",
				CanLogin: false,
			},
			wantQuery: `CREATE ROLE "app_readonly" NOLOGIN`,
			mockError: nil,
			wantError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`CREATE ROLE "app_readonly" NOLOGIN`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Create LOGIN role with password",
			opts: RoleOptions{
				RoleName: "app_user",
				CanLogin: true,
				Password: "secret123",
			},
			wantQuery: `CREATE ROLE "app_user" LOGIN PASSWORD \$1`,
			mockError: nil,
			wantError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`CREATE ROLE "app_user" LOGIN PASSWORD \$1`).
					WithArgs("secret123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Create LOGIN role without password",
			opts: RoleOptions{
				RoleName: "app_user2",
				CanLogin: true,
				Password: "",
			},
			wantQuery: `CREATE ROLE "app_user2" LOGIN`,
			mockError: nil,
			wantError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`CREATE ROLE "app_user2" LOGIN`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Create SUPERUSER role",
			opts: RoleOptions{
				RoleName:    "admin_role",
				CanLogin:    false,
				IsSuperuser: true,
			},
			wantQuery: `CREATE ROLE "admin_role" NOLOGIN SUPERUSER`,
			mockError: nil,
			wantError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`CREATE ROLE "admin_role" NOLOGIN SUPERUSER`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Create LOGIN SUPERUSER with password",
			opts: RoleOptions{
				RoleName:    "superadmin",
				CanLogin:    true,
				Password:    "admin123",
				IsSuperuser: true,
			},
			wantQuery: `CREATE ROLE "superadmin" LOGIN PASSWORD \$1 SUPERUSER`,
			mockError: nil,
			wantError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`CREATE ROLE "superadmin" LOGIN PASSWORD \$1 SUPERUSER`).
					WithArgs("admin123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Role with special characters in name",
			opts: RoleOptions{
				RoleName: `special"role`,
				CanLogin: false,
			},
			wantQuery: `CREATE ROLE "special""role" NOLOGIN`,
			mockError: nil,
			wantError: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`CREATE ROLE "special""role" NOLOGIN`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Database error",
			opts: RoleOptions{
				RoleName: "error_role",
				CanLogin: false,
			},
			wantQuery: `CREATE ROLE "error_role" NOLOGIN`,
			mockError: errors.New("database error"),
			wantError: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`CREATE ROLE "error_role" NOLOGIN`).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			manager := NewManager(db)

			// Setup expectations
			tt.setupMock(mock)

			err = manager.CreateRole(tt.opts)

			if (err != nil) != tt.wantError {
				t.Errorf("CreateRole() error = %v, wantError %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestManager_DeleteRole(t *testing.T) {
	tests := []struct {
		name         string
		roleName     string
		memberCount  int
		memberError  error
		deleteError  error
		wantError    bool
		errorContains string
	}{
		{
			name:        "Delete role successfully",
			roleName:    "old_role",
			memberCount: 0,
			memberError: nil,
			deleteError: nil,
			wantError:   false,
		},
		{
			name:          "Fail when role has members",
			roleName:      "role_with_members",
			memberCount:   3,
			memberError:   nil,
			deleteError:   nil,
			wantError:     true,
			errorContains: "still has 3 member(s)",
		},
		{
			name:          "Error checking members",
			roleName:      "check_error",
			memberCount:   0,
			memberError:   errors.New("query error"),
			deleteError:   nil,
			wantError:     true,
			errorContains: "failed to check role members",
		},
		{
			name:          "Error deleting role",
			roleName:      "delete_error",
			memberCount:   0,
			memberError:   nil,
			deleteError:   errors.New("cannot drop role"),
			wantError:     true,
			errorContains: "failed to delete role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			manager := NewManager(db)

			// Mock member count check
			rows := sqlmock.NewRows([]string{"count"}).AddRow(tt.memberCount)
			if tt.memberError != nil {
				mock.ExpectQuery("SELECT COUNT").WithArgs(tt.roleName).WillReturnError(tt.memberError)
			} else {
				mock.ExpectQuery("SELECT COUNT").WithArgs(tt.roleName).WillReturnRows(rows)
			}

			// Only expect DELETE if no members
			if tt.memberCount == 0 && tt.memberError == nil {
				expectation := mock.ExpectExec(`DROP ROLE "` + tt.roleName + `"`)
				if tt.deleteError != nil {
					expectation.WillReturnError(tt.deleteError)
				} else {
					expectation.WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}

			err = manager.DeleteRole(tt.roleName)

			if (err != nil) != tt.wantError {
				t.Errorf("DeleteRole() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.wantError && tt.errorContains != "" {
				if err == nil || !contains(err.Error(), tt.errorContains) {
					t.Errorf("DeleteRole() error = %v, want error containing %v", err, tt.errorContains)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestManager_DeleteRoleWithOptions(t *testing.T) {
	tests := []struct {
		name          string
		opts          DeleteRoleOptions
		memberCount   int
		wantError     bool
		errorContains string
	}{
		{
			name: "Delete role with reassign-to",
			opts: DeleteRoleOptions{
				RoleName:   "old_role",
				ReassignTo: "new_owner",
			},
			memberCount: 0,
			wantError:   false,
		},
		{
			name: "Delete role with drop-owned",
			opts: DeleteRoleOptions{
				RoleName:  "old_role",
				DropOwned: true,
			},
			memberCount: 0,
			wantError:   false,
		},
		{
			name: "Delete role with both options",
			opts: DeleteRoleOptions{
				RoleName:   "old_role",
				ReassignTo: "new_owner",
				DropOwned:  true,
			},
			memberCount: 0,
			wantError:   false,
		},
		{
			name: "Fail when reassign-to equals role being deleted",
			opts: DeleteRoleOptions{
				RoleName:   "same_role",
				ReassignTo: "same_role",
			},
			memberCount:   0,
			wantError:     true,
			errorContains: "cannot reassign objects to the same role being deleted",
		},
		{
			name: "Fail when role has members",
			opts: DeleteRoleOptions{
				RoleName:  "role_with_members",
				DropOwned: true,
			},
			memberCount:   2,
			wantError:     true,
			errorContains: "still has 2 member(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			manager := NewManager(db)

			// Early validation for same role reassignment
			if tt.opts.ReassignTo != "" && tt.opts.ReassignTo == tt.opts.RoleName {
				err = manager.DeleteRoleWithOptions(tt.opts)
				if !tt.wantError || err == nil {
					t.Errorf("DeleteRoleWithOptions() should fail for same role reassignment")
				}
				return
			}

			// Begin transaction
			mock.ExpectBegin()

			// Expect REASSIGN if specified
			if tt.opts.ReassignTo != "" {
				mock.ExpectExec("REASSIGN OWNED BY").
					WillReturnResult(sqlmock.NewResult(0, 0))
			}

			// Expect DROP OWNED if specified
			if tt.opts.DropOwned {
				mock.ExpectExec("DROP OWNED BY .* CASCADE").
					WillReturnResult(sqlmock.NewResult(0, 0))
			}

			// Mock member count check
			rows := sqlmock.NewRows([]string{"count"}).AddRow(tt.memberCount)
			mock.ExpectQuery("SELECT COUNT").WithArgs(tt.opts.RoleName).WillReturnRows(rows)

			// Only expect DELETE and COMMIT if no members
			if tt.memberCount == 0 {
				mock.ExpectExec("DROP ROLE").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}

			err = manager.DeleteRoleWithOptions(tt.opts)

			if (err != nil) != tt.wantError {
				t.Errorf("DeleteRoleWithOptions() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.wantError && tt.errorContains != "" {
				if err == nil || !contains(err.Error(), tt.errorContains) {
					t.Errorf("DeleteRoleWithOptions() error = %v, want error containing %v", err, tt.errorContains)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestManager_GrantRole(t *testing.T) {
	tests := []struct {
		name      string
		roleName  string
		username  string
		mockError error
		wantError bool
	}{
		{
			name:      "Grant role successfully",
			roleName:  "app_readonly",
			username:  "app_user1",
			mockError: nil,
			wantError: false,
		},
		{
			name:      "Grant role with special characters",
			roleName:  `special"role`,
			username:  `user"name`,
			mockError: nil,
			wantError: false,
		},
		{
			name:      "Database error",
			roleName:  "app_readonly",
			username:  "app_user1",
			mockError: errors.New("grant failed"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			manager := NewManager(db)

			expectation := mock.ExpectExec("GRANT")
			if tt.mockError != nil {
				expectation.WillReturnError(tt.mockError)
			} else {
				expectation.WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err = manager.GrantRole(tt.roleName, tt.username)

			if (err != nil) != tt.wantError {
				t.Errorf("GrantRole() error = %v, wantError %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestManager_RevokeRole(t *testing.T) {
	tests := []struct {
		name      string
		roleName  string
		username  string
		mockError error
		wantError bool
	}{
		{
			name:      "Revoke role successfully",
			roleName:  "app_readonly",
			username:  "app_user1",
			mockError: nil,
			wantError: false,
		},
		{
			name:      "Revoke role with special characters",
			roleName:  `special"role`,
			username:  `user"name`,
			mockError: nil,
			wantError: false,
		},
		{
			name:      "Database error",
			roleName:  "app_readonly",
			username:  "app_user1",
			mockError: errors.New("revoke failed"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			manager := NewManager(db)

			expectation := mock.ExpectExec("REVOKE")
			if tt.mockError != nil {
				expectation.WillReturnError(tt.mockError)
			} else {
				expectation.WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err = manager.RevokeRole(tt.roleName, tt.username)

			if (err != nil) != tt.wantError {
				t.Errorf("RevokeRole() error = %v, wantError %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestManager_ListRoles(t *testing.T) {
	tests := []struct {
		name        string
		mockRows    *sqlmock.Rows
		mockError   error
		wantCount   int
		wantError   bool
	}{
		{
			name: "List roles successfully",
			mockRows: sqlmock.NewRows([]string{"role_name", "is_superuser", "can_login", "connection_limit"}).
				AddRow("app_readonly", false, false, -1).
				AddRow("app_readwrite", false, false, -1).
				AddRow("app_user1", false, true, 10),
			mockError: nil,
			wantCount: 3,
			wantError: false,
		},
		{
			name:      "List roles - empty result",
			mockRows:  sqlmock.NewRows([]string{"role_name", "is_superuser", "can_login", "connection_limit"}),
			mockError: nil,
			wantCount: 0,
			wantError: false,
		},
		{
			name:      "Database error",
			mockRows:  nil,
			mockError: errors.New("query failed"),
			wantCount: 0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			manager := NewManager(db)

			if tt.mockError != nil {
				mock.ExpectQuery("SELECT").WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery("SELECT").WillReturnRows(tt.mockRows)
			}

			roles, err := manager.ListRoles()

			if (err != nil) != tt.wantError {
				t.Errorf("ListRoles() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && len(roles) != tt.wantCount {
				t.Errorf("ListRoles() returned %d roles, want %d", len(roles), tt.wantCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestManager_ListRoleMembers(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		mockRows    *sqlmock.Rows
		mockError   error
		wantMembers []string
		wantError   bool
	}{
		{
			name:     "List role members successfully",
			roleName: "app_readonly",
			mockRows: sqlmock.NewRows([]string{"rolname"}).
				AddRow("app_user1").
				AddRow("app_user2").
				AddRow("app_user3"),
			mockError:   nil,
			wantMembers: []string{"app_user1", "app_user2", "app_user3"},
			wantError:   false,
		},
		{
			name:        "List role members - empty result",
			roleName:    "empty_role",
			mockRows:    sqlmock.NewRows([]string{"rolname"}),
			mockError:   nil,
			wantMembers: nil,
			wantError:   false,
		},
		{
			name:        "Database error",
			roleName:    "error_role",
			mockRows:    nil,
			mockError:   errors.New("query failed"),
			wantMembers: nil,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			manager := NewManager(db)

			if tt.mockError != nil {
				mock.ExpectQuery("SELECT member.rolname").WithArgs(tt.roleName).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery("SELECT member.rolname").WithArgs(tt.roleName).WillReturnRows(tt.mockRows)
			}

			members, err := manager.ListRoleMembers(tt.roleName)

			if (err != nil) != tt.wantError {
				t.Errorf("ListRoleMembers() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				if len(members) != len(tt.wantMembers) {
					t.Errorf("ListRoleMembers() returned %d members, want %d", len(members), len(tt.wantMembers))
				}
				for i, member := range members {
					if i < len(tt.wantMembers) && member != tt.wantMembers[i] {
						t.Errorf("ListRoleMembers() member[%d] = %v, want %v", i, member, tt.wantMembers[i])
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestManager_ListUserRoles(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		mockRows  *sqlmock.Rows
		mockError error
		wantRoles []string
		wantError bool
	}{
		{
			name:     "List user roles successfully",
			username: "app_user1",
			mockRows: sqlmock.NewRows([]string{"rolname"}).
				AddRow("app_readonly").
				AddRow("reporting_access"),
			mockError: nil,
			wantRoles: []string{"app_readonly", "reporting_access"},
			wantError: false,
		},
		{
			name:      "List user roles - empty result",
			username:  "new_user",
			mockRows:  sqlmock.NewRows([]string{"rolname"}),
			mockError: nil,
			wantRoles: nil,
			wantError: false,
		},
		{
			name:      "Database error",
			username:  "error_user",
			mockRows:  nil,
			mockError: errors.New("query failed"),
			wantRoles: nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			manager := NewManager(db)

			if tt.mockError != nil {
				mock.ExpectQuery("SELECT role.rolname").WithArgs(tt.username).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery("SELECT role.rolname").WithArgs(tt.username).WillReturnRows(tt.mockRows)
			}

			roles, err := manager.ListUserRoles(tt.username)

			if (err != nil) != tt.wantError {
				t.Errorf("ListUserRoles() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				if len(roles) != len(tt.wantRoles) {
					t.Errorf("ListUserRoles() returned %d roles, want %d", len(roles), len(tt.wantRoles))
				}
				for i, role := range roles {
					if i < len(tt.wantRoles) && role != tt.wantRoles[i] {
						t.Errorf("ListUserRoles() role[%d] = %v, want %v", i, role, tt.wantRoles[i])
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
