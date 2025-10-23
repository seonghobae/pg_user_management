package user

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/seonghobae/pg_user_management/internal/auth"
)

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple identifier",
			input: "username",
			want:  `"username"`,
		},
		{
			name:  "Identifier with spaces",
			input: "user name",
			want:  `"user name"`,
		},
		{
			name:  "Identifier with quotes",
			input: `user"name`,
			want:  `"user""name"`,
		},
		{
			name:  "Empty identifier",
			input: "",
			want:  `""`,
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

func TestEscapeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple string",
			input: "password",
			want:  "password",
		},
		{
			name:  "String with single quote",
			input: "pass'word",
			want:  "pass''word",
		},
		{
			name:  "String with multiple quotes",
			input: "it's a 'test'",
			want:  "it''s a ''test''",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeString(tt.input)
			if got != tt.want {
				t.Errorf("escapeString() = %v, want %v", got, tt.want)
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

func TestManager_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)

	tests := []struct {
		name        string
		opts        UserOptions
		setupMock   func()
		expectError bool
	}{
		{
			name: "Create regular user with SCRAM-SHA-256",
			opts: UserOptions{
				Username:    "testuser",
				Password:    "secret123",
				IsSuperuser: false,
				CanLogin:    true,
				AuthMethod:  auth.ScramSHA256,
			},
			setupMock: func() {
				mock.ExpectExec("SET password_encryption").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(`CREATE USER "testuser" WITH PASSWORD 'secret123' NOSUPERUSER LOGIN`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Create superuser",
			opts: UserOptions{
				Username:    "admin",
				Password:    "admin123",
				IsSuperuser: true,
				CanLogin:    true,
				AuthMethod:  auth.MD5,
			},
			setupMock: func() {
				mock.ExpectExec("SET password_encryption").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(`CREATE USER "admin" WITH PASSWORD 'admin123' SUPERUSER LOGIN`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Create user without login",
			opts: UserOptions{
				Username:    "nologin",
				Password:    "pass",
				IsSuperuser: false,
				CanLogin:    false,
				AuthMethod:  auth.ScramSHA256,
			},
			setupMock: func() {
				mock.ExpectExec("SET password_encryption").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(`CREATE USER "nologin" WITH PASSWORD 'pass' NOSUPERUSER NOLOGIN`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := manager.CreateUser(tt.opts)

			if tt.expectError {
				if err == nil {
					t.Errorf("CreateUser() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("CreateUser() unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestManager_DeleteUser(t *testing.T) {
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
	}{
		{
			name:     "Delete existing user",
			username: "testuser",
			setupMock: func() {
				// Mock for getting databases
				rows := sqlmock.NewRows([]string{"datname"}).
					AddRow("testdb")
				mock.ExpectQuery("SELECT datname FROM pg_database").WillReturnRows(rows)

				// Mock for revoke operations (they can fail, we just log)
				mock.ExpectExec("REVOKE ALL PRIVILEGES").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("REVOKE ALL PRIVILEGES").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("REVOKE ALL PRIVILEGES").WillReturnResult(sqlmock.NewResult(0, 0))

				// Mock for DROP USER
				mock.ExpectExec(`DROP USER IF EXISTS "testuser"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := manager.DeleteUser(tt.username)

			if tt.expectError {
				if err == nil {
					t.Errorf("DeleteUser() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("DeleteUser() unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestManager_ListUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)

	tests := []struct {
		name        string
		setupMock   func()
		expectError bool
		wantCount   int
	}{
		{
			name: "List users successfully",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"username", "is_superuser", "can_create_db", "can_create_role"}).
					AddRow("user1", false, false, false).
					AddRow("user2", true, true, true)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			expectError: false,
			wantCount:   2,
		},
		{
			name: "List users - empty result",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"username", "is_superuser", "can_create_db", "can_create_role"})
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			expectError: false,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			users, err := manager.ListUsers()

			if tt.expectError {
				if err == nil {
					t.Errorf("ListUsers() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ListUsers() unexpected error: %v", err)
				}
				if len(users) != tt.wantCount {
					t.Errorf("ListUsers() returned %d users, want %d", len(users), tt.wantCount)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestManager_ModifyUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	manager := NewManager(db)

	tests := []struct {
		name        string
		opts        UserOptions
		setupMock   func()
		expectError bool
	}{
		{
			name: "Modify user password",
			opts: UserOptions{
				Username:    "testuser",
				Password:    "newpass",
				IsSuperuser: false,
				CanLogin:    true,
				AuthMethod:  auth.ScramSHA256,
			},
			setupMock: func() {
				mock.ExpectExec("SET password_encryption").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(`ALTER USER "testuser" WITH PASSWORD 'newpass' NOSUPERUSER LOGIN`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Modify user to superuser",
			opts: UserOptions{
				Username:    "testuser",
				Password:    "pass",
				IsSuperuser: true,
				CanLogin:    true,
				AuthMethod:  auth.MD5,
			},
			setupMock: func() {
				mock.ExpectExec("SET password_encryption").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(`ALTER USER "testuser" WITH PASSWORD 'pass' SUPERUSER LOGIN`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := manager.ModifyUser(tt.opts)

			if tt.expectError {
				if err == nil {
					t.Errorf("ModifyUser() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ModifyUser() unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}
