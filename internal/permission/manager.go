package permission

import (
	"database/sql"
	"fmt"
	"strings"
)

// Manager handles permission management operations
type Manager struct {
	db *sql.DB
}

// NewManager creates a new permission manager
func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// Privilege represents a database privilege
type Privilege string

const (
	SELECT Privilege = "SELECT"
	INSERT Privilege = "INSERT"
	UPDATE Privilege = "UPDATE"
	DELETE Privilege = "DELETE"
	ALL    Privilege = "ALL"
)

// GrantOptions contains options for granting privileges
type GrantOptions struct {
	Username   string
	Database   string
	Schema     string
	Table      string
	Privileges []Privilege
}

// GrantTablePrivileges grants table-level privileges to a user
func (m *Manager) GrantTablePrivileges(opts GrantOptions) error {
	if len(opts.Privileges) == 0 {
		return fmt.Errorf("no privileges specified")
	}

	// Build privilege list
	privList := make([]string, len(opts.Privileges))
	for i, priv := range opts.Privileges {
		privList[i] = string(priv)
	}
	privileges := strings.Join(privList, ", ")

	// Grant on specific table
	if opts.Table != "" {
		query := fmt.Sprintf("GRANT %s ON TABLE %s.%s TO %s",
			privileges,
			quoteIdentifier(opts.Schema),
			quoteIdentifier(opts.Table),
			quoteIdentifier(opts.Username))

		if _, err := m.db.Exec(query); err != nil {
			return fmt.Errorf("failed to grant table privileges: %w", err)
		}
	} else if opts.Schema != "" {
		// Grant on all tables in schema
		query := fmt.Sprintf("GRANT %s ON ALL TABLES IN SCHEMA %s TO %s",
			privileges,
			quoteIdentifier(opts.Schema),
			quoteIdentifier(opts.Username))

		if _, err := m.db.Exec(query); err != nil {
			return fmt.Errorf("failed to grant schema privileges: %w", err)
		}

		// Set default privileges for future tables
		defaultQuery := fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT %s ON TABLES TO %s",
			quoteIdentifier(opts.Schema),
			privileges,
			quoteIdentifier(opts.Username))

		if _, err := m.db.Exec(defaultQuery); err != nil {
			return fmt.Errorf("failed to set default privileges: %w", err)
		}
	} else {
		return fmt.Errorf("either table or schema must be specified")
	}

	return nil
}

// RevokeTablePrivileges revokes table-level privileges from a user
func (m *Manager) RevokeTablePrivileges(opts GrantOptions) error {
	if len(opts.Privileges) == 0 {
		return fmt.Errorf("no privileges specified")
	}

	// Build privilege list
	privList := make([]string, len(opts.Privileges))
	for i, priv := range opts.Privileges {
		privList[i] = string(priv)
	}
	privileges := strings.Join(privList, ", ")

	// Revoke from specific table
	if opts.Table != "" {
		query := fmt.Sprintf("REVOKE %s ON TABLE %s.%s FROM %s",
			privileges,
			quoteIdentifier(opts.Schema),
			quoteIdentifier(opts.Table),
			quoteIdentifier(opts.Username))

		if _, err := m.db.Exec(query); err != nil {
			return fmt.Errorf("failed to revoke table privileges: %w", err)
		}
	} else if opts.Schema != "" {
		// Revoke from all tables in schema
		query := fmt.Sprintf("REVOKE %s ON ALL TABLES IN SCHEMA %s FROM %s",
			privileges,
			quoteIdentifier(opts.Schema),
			quoteIdentifier(opts.Username))

		if _, err := m.db.Exec(query); err != nil {
			return fmt.Errorf("failed to revoke schema privileges: %w", err)
		}
	} else {
		return fmt.Errorf("either table or schema must be specified")
	}

	return nil
}

// GrantDatabaseAccess grants connect privilege to a database
func (m *Manager) GrantDatabaseAccess(username, database string) error {
	query := fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO %s",
		quoteIdentifier(database),
		quoteIdentifier(username))

	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to grant database access: %w", err)
	}

	return nil
}

// GrantSchemaUsage grants usage privilege on a schema
func (m *Manager) GrantSchemaUsage(username, schema string) error {
	query := fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s",
		quoteIdentifier(schema),
		quoteIdentifier(username))

	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to grant schema usage: %w", err)
	}

	return nil
}

// GrantAllFunctionsInSchema grants execute privilege on all functions in a schema
func (m *Manager) GrantAllFunctionsInSchema(username, schema string) error {
	// Grant execute on all existing functions
	query := fmt.Sprintf("GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA %s TO %s",
		quoteIdentifier(schema),
		quoteIdentifier(username))

	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to grant function privileges: %w", err)
	}

	// Set default privileges for future functions
	defaultQuery := fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT EXECUTE ON FUNCTIONS TO %s",
		quoteIdentifier(schema),
		quoteIdentifier(username))

	if _, err := m.db.Exec(defaultQuery); err != nil {
		return fmt.Errorf("failed to set default function privileges: %w", err)
	}

	return nil
}

// GrantAllFunctionsInDatabase grants execute privilege on all functions in all schemas of a database
func (m *Manager) GrantAllFunctionsInDatabase(username, database string) error {
	// First, get all schemas in the database
	query := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE catalog_name = $1
		AND schema_name NOT IN ('pg_catalog', 'information_schema')
	`

	rows, err := m.db.Query(query, database)
	if err != nil {
		return fmt.Errorf("failed to get schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	// Grant execute on all functions in each schema
	for _, schema := range schemas {
		if err := m.GrantAllFunctionsInSchema(username, schema); err != nil {
			return fmt.Errorf("failed to grant functions in schema %s: %w", schema, err)
		}
	}

	return nil
}

// ListUserPrivileges lists all privileges for a user
func (m *Manager) ListUserPrivileges(username string) ([]map[string]interface{}, error) {
	query := `
		SELECT
			table_catalog as database,
			table_schema as schema,
			table_name as table,
			privilege_type as privilege
		FROM information_schema.table_privileges
		WHERE grantee = $1
		ORDER BY table_catalog, table_schema, table_name, privilege_type
	`

	rows, err := m.db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to list privileges: %w", err)
	}
	defer rows.Close()

	var privileges []map[string]interface{}
	for rows.Next() {
		var database, schema, table, privilege string

		if err := rows.Scan(&database, &schema, &table, &privilege); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		privileges = append(privileges, map[string]interface{}{
			"database":  database,
			"schema":    schema,
			"table":     table,
			"privilege": privilege,
		})
	}

	return privileges, nil
}

// quoteIdentifier quotes a PostgreSQL identifier
func quoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}
