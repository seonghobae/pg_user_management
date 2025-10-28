package role

import (
	"database/sql"
	"fmt"
	"strings"
)

// Manager handles PostgreSQL role operations
type Manager struct {
	db *sql.DB
}

// NewManager creates a new role manager
func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// RoleOptions contains options for role creation
type RoleOptions struct {
	RoleName    string
	CanLogin    bool
	IsSuperuser bool
	Password    string
}

// CreateRole creates a new PostgreSQL role (typically a group role with NOLOGIN)
func (m *Manager) CreateRole(opts RoleOptions) error {
	// Build the base query with properly quoted identifier
	query := fmt.Sprintf("CREATE ROLE %s", quoteIdentifier(opts.RoleName))

	if opts.CanLogin {
		query += " LOGIN"
		if opts.Password != "" {
			// Use parameterized query for password to avoid SQL injection and log leakage
			query += " PASSWORD $1"
			if opts.IsSuperuser {
				query += " SUPERUSER"
			}
			if _, err := m.db.Exec(query, opts.Password); err != nil {
				return fmt.Errorf("failed to create role: %w", err)
			}
			return nil
		}
	} else {
		query += " NOLOGIN"
	}

	if opts.IsSuperuser {
		query += " SUPERUSER"
	}

	// Execute without password parameter
	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// DeleteRoleOptions contains options for safe role deletion
type DeleteRoleOptions struct {
	RoleName   string
	ReassignTo string // If set, reassign owned objects to this role before deletion
	DropOwned  bool   // If true, drop all objects owned by the role
}

// DeleteRole deletes a PostgreSQL role
// Note: This will fail if the role owns objects or has granted privileges.
// Use DeleteRoleWithOptions for safe deletion with cleanup.
func (m *Manager) DeleteRole(roleName string) error {
	// Check if the role has any members before deletion
	var memberCount int
	checkMembersQuery := `
		SELECT COUNT(*)
		FROM pg_auth_members
		JOIN pg_roles AS role ON pg_auth_members.roleid = role.oid
		WHERE role.rolname = $1
	`
	if err := m.db.QueryRow(checkMembersQuery, roleName).Scan(&memberCount); err != nil {
		return fmt.Errorf("failed to check role members: %w", err)
	}
	if memberCount > 0 {
		return fmt.Errorf("role %s still has %d member(s); revoke them before deletion using revoke-role command",
			roleName, memberCount)
	}

	query := fmt.Sprintf("DROP ROLE %s", quoteIdentifier(roleName))

	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// DeleteRoleWithOptions deletes a PostgreSQL role with transactional cleanup
// Supports REASSIGN OWNED and DROP OWNED to handle role dependencies
func (m *Manager) DeleteRoleWithOptions(opts DeleteRoleOptions) error {
	// Validate that ReassignTo is different from the role being deleted
	if opts.ReassignTo != "" && opts.ReassignTo == opts.RoleName {
		return fmt.Errorf("cannot reassign objects to the same role being deleted (%s)", opts.RoleName)
	}

	// Start a transaction for atomic cleanup + deletion
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// If ReassignTo is specified, reassign all owned objects
	if opts.ReassignTo != "" {
		reassignQuery := fmt.Sprintf("REASSIGN OWNED BY %s TO %s",
			quoteIdentifier(opts.RoleName),
			quoteIdentifier(opts.ReassignTo))
		if _, err := tx.Exec(reassignQuery); err != nil {
			return fmt.Errorf("failed to reassign owned objects: %w", err)
		}
	}

	// If DropOwned is true, drop all objects owned by the role
	if opts.DropOwned {
		// Use CASCADE to handle dependent objects
		dropOwnedQuery := fmt.Sprintf("DROP OWNED BY %s CASCADE",
			quoteIdentifier(opts.RoleName))
		if _, err := tx.Exec(dropOwnedQuery); err != nil {
			return fmt.Errorf("failed to drop owned objects: %w", err)
		}
	}

	// Check if the role has any members before deletion
	var memberCount int
	checkMembersQuery := `
		SELECT COUNT(*)
		FROM pg_auth_members
		JOIN pg_roles AS role ON pg_auth_members.roleid = role.oid
		WHERE role.rolname = $1
	`
	if err := tx.QueryRow(checkMembersQuery, opts.RoleName).Scan(&memberCount); err != nil {
		return fmt.Errorf("failed to check role members: %w", err)
	}
	if memberCount > 0 {
		return fmt.Errorf("role %s still has %d member(s); revoke them before deletion using revoke-role command",
			opts.RoleName, memberCount)
	}

	// Finally, drop the role
	dropRoleQuery := fmt.Sprintf("DROP ROLE %s", quoteIdentifier(opts.RoleName))
	if _, err := tx.Exec(dropRoleQuery); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GrantRole grants a role to a user (adds user to role group)
func (m *Manager) GrantRole(roleName, username string) error {
	query := fmt.Sprintf("GRANT %s TO %s",
		quoteIdentifier(roleName),
		quoteIdentifier(username))

	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to grant role: %w", err)
	}

	return nil
}

// RevokeRole revokes a role from a user (removes user from role group)
func (m *Manager) RevokeRole(roleName, username string) error {
	query := fmt.Sprintf("REVOKE %s FROM %s",
		quoteIdentifier(roleName),
		quoteIdentifier(username))

	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	return nil
}

// ListRoles lists all roles (excluding system roles)
func (m *Manager) ListRoles() ([]map[string]interface{}, error) {
	query := `
		SELECT
			rolname as role_name,
			rolsuper as is_superuser,
			rolcanlogin as can_login,
			rolconnlimit as connection_limit
		FROM pg_roles
		WHERE rolname NOT LIKE 'pg_%'
		AND rolname != 'postgres'
		ORDER BY rolname
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []map[string]interface{}
	for rows.Next() {
		var roleName string
		var isSuperuser, canLogin bool
		var connLimit int

		if err := rows.Scan(&roleName, &isSuperuser, &canLogin, &connLimit); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		role := map[string]interface{}{
			"role_name":         roleName,
			"is_superuser":      isSuperuser,
			"can_login":         canLogin,
			"connection_limit":  connLimit,
		}
		roles = append(roles, role)
	}

	// Check for errors from iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during role iteration: %w", err)
	}

	return roles, nil
}

// ListRoleMembers lists all members of a specific role
func (m *Manager) ListRoleMembers(roleName string) ([]string, error) {
	query := `
		SELECT member.rolname
		FROM pg_auth_members
		JOIN pg_roles AS member ON pg_auth_members.member = member.oid
		JOIN pg_roles AS role ON pg_auth_members.roleid = role.oid
		WHERE role.rolname = $1
		ORDER BY member.rolname
	`

	rows, err := m.db.Query(query, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to list role members: %w", err)
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var memberName string
		if err := rows.Scan(&memberName); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, memberName)
	}

	// Check for errors from iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during member iteration: %w", err)
	}

	return members, nil
}

// ListUserRoles lists all roles granted to a specific user
func (m *Manager) ListUserRoles(username string) ([]string, error) {
	query := `
		SELECT role.rolname
		FROM pg_auth_members
		JOIN pg_roles AS member ON pg_auth_members.member = member.oid
		JOIN pg_roles AS role ON pg_auth_members.roleid = role.oid
		WHERE member.rolname = $1
		ORDER BY role.rolname
	`

	rows, err := m.db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to list user roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, roleName)
	}

	// Check for errors from iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during user roles iteration: %w", err)
	}

	return roles, nil
}

// quoteIdentifier quotes an identifier to prevent SQL injection
// Escapes embedded double quotes by doubling them per PostgreSQL spec
func quoteIdentifier(name string) string {
	// Escape any double quotes in the identifier by doubling them
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}
