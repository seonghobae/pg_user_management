package permission

import (
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
			input: "tablename",
			want:  `"tablename"`,
		},
		{
			name:  "Identifier with spaces",
			input: "table name",
			want:  `"table name"`,
		},
		{
			name:  "Identifier with quotes",
			input: `table"name`,
			want:  `"table""name"`,
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

func TestManager_GrantTablePrivileges(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)

	tests := []struct {
		name        string
		opts        GrantOptions
		setupMock   func()
		expectError bool
	}{
		{
			name: "Grant SELECT on specific table",
			opts: GrantOptions{
				Username:   "testuser",
				Schema:     "public",
				Table:      "users",
				Privileges: []Privilege{SELECT},
			},
			setupMock: func() {
				mock.ExpectExec(`GRANT SELECT ON TABLE "public"."users" TO "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Grant multiple privileges on table",
			opts: GrantOptions{
				Username:   "testuser",
				Schema:     "public",
				Table:      "products",
				Privileges: []Privilege{SELECT, INSERT, UPDATE},
			},
			setupMock: func() {
				mock.ExpectExec(`GRANT SELECT, INSERT, UPDATE ON TABLE "public"."products" TO "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Grant ALL on all tables in schema",
			opts: GrantOptions{
				Username:   "testuser",
				Schema:     "public",
				Table:      "",
				Privileges: []Privilege{ALL},
			},
			setupMock: func() {
				mock.ExpectExec(`GRANT ALL ON ALL TABLES IN SCHEMA "public" TO "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(`ALTER DEFAULT PRIVILEGES IN SCHEMA "public" GRANT ALL ON TABLES TO "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Grant without privileges - should error",
			opts: GrantOptions{
				Username:   "testuser",
				Schema:     "public",
				Table:      "users",
				Privileges: []Privilege{},
			},
			setupMock:   func() {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := manager.GrantTablePrivileges(tt.opts)

			if tt.expectError {
				if err == nil {
					t.Errorf("GrantTablePrivileges() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GrantTablePrivileges() unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestManager_RevokeTablePrivileges(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)

	tests := []struct {
		name        string
		opts        GrantOptions
		setupMock   func()
		expectError bool
	}{
		{
			name: "Revoke SELECT from specific table",
			opts: GrantOptions{
				Username:   "testuser",
				Schema:     "public",
				Table:      "users",
				Privileges: []Privilege{SELECT},
			},
			setupMock: func() {
				mock.ExpectExec(`REVOKE SELECT ON TABLE "public"."users" FROM "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Revoke ALL from all tables in schema",
			opts: GrantOptions{
				Username:   "testuser",
				Schema:     "public",
				Table:      "",
				Privileges: []Privilege{ALL},
			},
			setupMock: func() {
				mock.ExpectExec(`REVOKE ALL ON ALL TABLES IN SCHEMA "public" FROM "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := manager.RevokeTablePrivileges(tt.opts)

			if tt.expectError {
				if err == nil {
					t.Errorf("RevokeTablePrivileges() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("RevokeTablePrivileges() unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestManager_GrantDatabaseAccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)

	tests := []struct {
		name        string
		username    string
		database    string
		setupMock   func()
		expectError bool
	}{
		{
			name:     "Grant database access",
			username: "testuser",
			database: "testdb",
			setupMock: func() {
				mock.ExpectExec(`GRANT CONNECT ON DATABASE "testdb" TO "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := manager.GrantDatabaseAccess(tt.username, tt.database)

			if tt.expectError {
				if err == nil {
					t.Errorf("GrantDatabaseAccess() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GrantDatabaseAccess() unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestManager_GrantSchemaUsage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)

	tests := []struct {
		name        string
		username    string
		schema      string
		setupMock   func()
		expectError bool
	}{
		{
			name:     "Grant schema usage",
			username: "testuser",
			schema:   "public",
			setupMock: func() {
				mock.ExpectExec(`GRANT USAGE ON SCHEMA "public" TO "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := manager.GrantSchemaUsage(tt.username, tt.schema)

			if tt.expectError {
				if err == nil {
					t.Errorf("GrantSchemaUsage() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GrantSchemaUsage() unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestManager_GrantAllFunctionsInSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)

	tests := []struct {
		name        string
		username    string
		schema      string
		setupMock   func()
		expectError bool
	}{
		{
			name:     "Grant function execute privileges",
			username: "testuser",
			schema:   "public",
			setupMock: func() {
				mock.ExpectExec(`GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA "public" TO "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(`ALTER DEFAULT PRIVILEGES IN SCHEMA "public" GRANT EXECUTE ON FUNCTIONS TO "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := manager.GrantAllFunctionsInSchema(tt.username, tt.schema)

			if tt.expectError {
				if err == nil {
					t.Errorf("GrantAllFunctionsInSchema() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GrantAllFunctionsInSchema() unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestManager_ListUserPrivileges(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)

	tests := []struct {
		name        string
		username    string
		setupMock   func()
		expectError bool
		wantCount   int
	}{
		{
			name:     "List user privileges successfully",
			username: "testuser",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"database", "schema", "table", "privilege"}).
					AddRow("testdb", "public", "users", "SELECT").
					AddRow("testdb", "public", "users", "INSERT")
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			expectError: false,
			wantCount:   2,
		},
		{
			name:     "List user privileges - empty result",
			username: "testuser",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"database", "schema", "table", "privilege"})
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			expectError: false,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			privileges, err := manager.ListUserPrivileges(tt.username)

			if tt.expectError {
				if err == nil {
					t.Errorf("ListUserPrivileges() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ListUserPrivileges() unexpected error: %v", err)
				}
				if len(privileges) != tt.wantCount {
					t.Errorf("ListUserPrivileges() returned %d privileges, want %d", len(privileges), tt.wantCount)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}
