package user

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/seonghobae/pg_user_management/internal/auth"
)

// Manager handles user management operations
type Manager struct {
	db *sql.DB
}

// NewManager creates a new user manager
func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// UserOptions contains options for creating/modifying a user
type UserOptions struct {
	Username   string
	Password   string
	IsSuperuser bool
	CanLogin   bool
	AuthMethod auth.AuthMethod
}

// CreateUser creates a new PostgreSQL user
func (m *Manager) CreateUser(opts UserOptions) error {
	// Set password encryption method
	if err := m.setPasswordEncryption(opts.AuthMethod); err != nil {
		return fmt.Errorf("failed to set password encryption: %w", err)
	}

	// Build CREATE USER statement with parameterized password
	query := fmt.Sprintf("CREATE USER %s WITH PASSWORD $1", quoteIdentifier(opts.Username))

	if opts.IsSuperuser {
		query += " SUPERUSER"
	} else {
		query += " NOSUPERUSER"
	}

	if opts.CanLogin {
		query += " LOGIN"
	} else {
		query += " NOLOGIN"
	}

	// Execute with parameterized password to avoid logging sensitive data
	if _, err := m.db.Exec(query, opts.Password); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// ModifyUser modifies an existing PostgreSQL user
func (m *Manager) ModifyUser(opts UserOptions) error {
	// Set password encryption method if changing password
	if opts.Password != "" {
		if err := m.setPasswordEncryption(opts.AuthMethod); err != nil {
			return fmt.Errorf("failed to set password encryption: %w", err)
		}
	}

	// Build ALTER USER statement with parameterized password
	var alterParts []string
	var args []interface{}

	if opts.Password != "" {
		alterParts = append(alterParts, "PASSWORD $1")
		args = append(args, opts.Password)
	}

	if opts.IsSuperuser {
		alterParts = append(alterParts, "SUPERUSER")
	} else {
		alterParts = append(alterParts, "NOSUPERUSER")
	}

	if opts.CanLogin {
		alterParts = append(alterParts, "LOGIN")
	} else {
		alterParts = append(alterParts, "NOLOGIN")
	}

	if len(alterParts) == 0 {
		return fmt.Errorf("no modifications specified")
	}

	query := fmt.Sprintf("ALTER USER %s WITH %s",
		quoteIdentifier(opts.Username),
		strings.Join(alterParts, " "))

	// Execute with parameterized password to avoid logging sensitive data
	if _, err := m.db.Exec(query, args...); err != nil {
		return fmt.Errorf("failed to modify user: %w", err)
	}

	return nil
}

// DeleteUser deletes a PostgreSQL user
func (m *Manager) DeleteUser(username string) error {
	// First, revoke all privileges
	if err := m.revokeAllPrivileges(username); err != nil {
		// Log warning but continue
		fmt.Printf("Warning: failed to revoke all privileges: %v\n", err)
	}

	// Drop the user
	query := fmt.Sprintf("DROP USER IF EXISTS %s", quoteIdentifier(username))
	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers lists all non-system users
func (m *Manager) ListUsers() ([]map[string]interface{}, error) {
	query := `
		SELECT
			rolname as username,
			rolsuper as is_superuser,
			rolcreatedb as can_create_db,
			rolcreaterole as can_create_role
		FROM pg_roles
		WHERE rolname NOT LIKE 'pg_%'
		AND rolcanlogin = true
		ORDER BY rolname
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var username string
		var isSuperuser, canCreateDB, canCreateRole bool

		if err := rows.Scan(&username, &isSuperuser, &canCreateDB, &canCreateRole); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		users = append(users, map[string]interface{}{
			"username":        username,
			"is_superuser":    isSuperuser,
			"can_create_db":   canCreateDB,
			"can_create_role": canCreateRole,
		})
	}

	return users, nil
}

// setPasswordEncryption sets the password encryption method for the current session
func (m *Manager) setPasswordEncryption(method auth.AuthMethod) error {
	query := fmt.Sprintf("SET password_encryption = '%s'", method.GetPasswordEncryption())
	_, err := m.db.Exec(query)
	return err
}

// revokeAllPrivileges revokes all privileges from a user
func (m *Manager) revokeAllPrivileges(username string) error {
	// Get all databases
	rows, err := m.db.Query("SELECT datname FROM pg_database WHERE datistemplate = false")
	if err != nil {
		return err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbname string
		if err := rows.Scan(&dbname); err != nil {
			continue
		}
		databases = append(databases, dbname)
	}

	// Revoke privileges on each database
	for _, dbname := range databases {
		query := fmt.Sprintf("REVOKE ALL PRIVILEGES ON DATABASE %s FROM %s",
			quoteIdentifier(dbname),
			quoteIdentifier(username))
		m.db.Exec(query)

		// Revoke privileges on all tables in public schema
		query = fmt.Sprintf("REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM %s",
			quoteIdentifier(username))
		m.db.Exec(query)

		// Revoke privileges on all sequences in public schema
		query = fmt.Sprintf("REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM %s",
			quoteIdentifier(username))
		m.db.Exec(query)
	}

	return nil
}

// quoteIdentifier quotes a PostgreSQL identifier
func quoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

