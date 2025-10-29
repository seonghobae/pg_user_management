package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/seonghobae/pg_user_management/internal/auth"
	"github.com/seonghobae/pg_user_management/internal/database"
	"github.com/seonghobae/pg_user_management/internal/hba"
	"github.com/seonghobae/pg_user_management/internal/permission"
	"github.com/seonghobae/pg_user_management/internal/role"
	"github.com/seonghobae/pg_user_management/internal/user"
	"github.com/seonghobae/pg_user_management/pkg/config"
)

// main is the program entry point for the pg_user_admin CLI.
// It parses the first command-line argument as a subcommand and dispatches to the corresponding command handler; on missing or unknown commands it prints usage and exits with status 1.
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create-user":
		createUserCmd()
	case "modify-user":
		modifyUserCmd()
	case "delete-user":
		deleteUserCmd()
	case "list-users":
		listUsersCmd()
	case "grant":
		grantCmd()
	case "revoke":
		revokeCmd()
	case "list-privileges":
		listPrivilegesCmd()
	case "hba-add":
		hbaAddCmd()
	case "hba-remove":
		hbaRemoveCmd()
	case "hba-list":
		hbaListCmd()
	case "hba-reload":
		hbaReloadCmd()
	case "create-role":
		createRoleCmd()
	case "delete-role":
		deleteRoleCmd()
	case "list-roles":
		listRolesCmd()
	case "grant-role":
		grantRoleCmd()
	case "revoke-role":
		revokeRoleCmd()
	case "list-role-members":
		listRoleMembersCmd()
	case "list-user-roles":
		listUserRolesCmd()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func createUserCmd() {
	fs := flag.NewFlagSet("create-user", flag.ExitOnError)
	username := fs.String("username", "", "Username to create (required)")
	password := fs.String("password", "", "Password for the user (required)")
	isSuperuser := fs.Bool("superuser", false, "Create as superuser")
	canLogin := fs.Bool("login", true, "Allow login")
	authMethod := fs.String("auth", "scram-sha-256", "Authentication method (md5 or scram-sha-256)")

	fs.Parse(os.Args[2:])

	if *username == "" || *password == "" {
		fmt.Println("Error: username and password are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	method, err := auth.ValidateAuthMethod(*authMethod)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	userMgr := user.NewManager(db.DB)
	opts := user.UserOptions{
		Username:   *username,
		Password:   *password,
		IsSuperuser: *isSuperuser,
		CanLogin:   *canLogin,
		AuthMethod: method,
	}

	if err := userMgr.CreateUser(opts); err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created user: %s\n", *username)
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Superuser: %v\n", *isSuperuser)
	fmt.Printf("  Can Login: %v\n", *canLogin)
	fmt.Printf("  Auth Method: %s\n", method)
}

func modifyUserCmd() {
	fs := flag.NewFlagSet("modify-user", flag.ExitOnError)
	username := fs.String("username", "", "Username to modify (required)")
	password := fs.String("password", "", "New password")
	isSuperuser := fs.Bool("superuser", false, "Set as superuser")
	noSuperuser := fs.Bool("no-superuser", false, "Remove superuser privilege")
	canLogin := fs.Bool("login", false, "Allow login")
	noLogin := fs.Bool("no-login", false, "Disallow login")
	authMethod := fs.String("auth", "scram-sha-256", "Authentication method (md5 or scram-sha-256)")

	fs.Parse(os.Args[2:])

	if *username == "" {
		fmt.Println("Error: username is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	method, err := auth.ValidateAuthMethod(*authMethod)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	userMgr := user.NewManager(db.DB)

	// Determine superuser status
	superuser := *isSuperuser
	if *noSuperuser {
		superuser = false
	}

	// Determine login status
	login := true
	if *canLogin {
		login = true
	}
	if *noLogin {
		login = false
	}

	opts := user.UserOptions{
		Username:   *username,
		Password:   *password,
		IsSuperuser: superuser,
		CanLogin:   login,
		AuthMethod: method,
	}

	if err := userMgr.ModifyUser(opts); err != nil {
		fmt.Printf("Error modifying user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully modified user: %s\n", *username)
}

func deleteUserCmd() {
	fs := flag.NewFlagSet("delete-user", flag.ExitOnError)
	username := fs.String("username", "", "Username to delete (required)")

	fs.Parse(os.Args[2:])

	if *username == "" {
		fmt.Println("Error: username is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	userMgr := user.NewManager(db.DB)

	if err := userMgr.DeleteUser(*username); err != nil {
		fmt.Printf("Error deleting user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully deleted user: %s\n", *username)
}

func listUsersCmd() {
	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	userMgr := user.NewManager(db.DB)

	users, err := userMgr.ListUsers()
	if err != nil {
		fmt.Printf("Error listing users: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PostgreSQL Users:")
	fmt.Println("--------------------------------------------------")
	for _, u := range users {
		fmt.Printf("Username: %s\n", u["username"])
		fmt.Printf("  Superuser: %v\n", u["is_superuser"])
		fmt.Printf("  Can Create DB: %v\n", u["can_create_db"])
		fmt.Printf("  Can Create Role: %v\n", u["can_create_role"])
		fmt.Println()
	}
}

func grantCmd() {
	fs := flag.NewFlagSet("grant", flag.ExitOnError)
	username := fs.String("username", "", "Username (required)")
	database := fs.String("database", "", "Database name")
	schema := fs.String("schema", "", "Schema name (default: public)")
	table := fs.String("table", "", "Table name (optional, grants on all tables if not specified)")
	privileges := fs.String("privileges", "", "Comma-separated privileges (SELECT,INSERT,UPDATE,DELETE,ALL)")
	grantFunctions := fs.Bool("grant-functions", false, "Grant EXECUTE on all functions in schema")

	fs.Parse(os.Args[2:])

	if *username == "" {
		fmt.Println("Error: username is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	if *schema == "" {
		*schema = "public"
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	permMgr := permission.NewManager(db.DB)

	// Grant database access if specified
	if *database != "" {
		if err := permMgr.GrantDatabaseAccess(*username, *database); err != nil {
			fmt.Printf("Error granting database access: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Granted CONNECT on database %s to %s\n", *database, *username)
	}

	// Grant schema usage
	if err := permMgr.GrantSchemaUsage(*username, *schema); err != nil {
		fmt.Printf("Error granting schema usage: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Granted USAGE on schema %s to %s\n", *schema, *username)

	// Grant table privileges
	if *privileges != "" {
		privList := strings.Split(*privileges, ",")
		var privs []permission.Privilege
		for _, p := range privList {
			privs = append(privs, permission.Privilege(strings.TrimSpace(strings.ToUpper(p))))
		}

		opts := permission.GrantOptions{
			Username:   *username,
			Database:   *database,
			Schema:     *schema,
			Table:      *table,
			Privileges: privs,
		}

		if err := permMgr.GrantTablePrivileges(opts); err != nil {
			fmt.Printf("Error granting table privileges: %v\n", err)
			os.Exit(1)
		}

		if *table != "" {
			fmt.Printf("Granted %s on table %s.%s to %s\n", *privileges, *schema, *table, *username)
		} else {
			fmt.Printf("Granted %s on all tables in schema %s to %s\n", *privileges, *schema, *username)
		}
	}

	// Grant function privileges
	if *grantFunctions {
		if err := permMgr.GrantAllFunctionsInSchema(*username, *schema); err != nil {
			fmt.Printf("Error granting function privileges: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Granted EXECUTE on all functions in schema %s to %s\n", *schema, *username)
	}
}

func revokeCmd() {
	fs := flag.NewFlagSet("revoke", flag.ExitOnError)
	username := fs.String("username", "", "Username (required)")
	schema := fs.String("schema", "public", "Schema name")
	table := fs.String("table", "", "Table name (optional, revokes from all tables if not specified)")
	privileges := fs.String("privileges", "", "Comma-separated privileges (SELECT,INSERT,UPDATE,DELETE,ALL)")

	fs.Parse(os.Args[2:])

	if *username == "" || *privileges == "" {
		fmt.Println("Error: username and privileges are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	permMgr := permission.NewManager(db.DB)

	privList := strings.Split(*privileges, ",")
	var privs []permission.Privilege
	for _, p := range privList {
		privs = append(privs, permission.Privilege(strings.TrimSpace(strings.ToUpper(p))))
	}

	opts := permission.GrantOptions{
		Username:   *username,
		Schema:     *schema,
		Table:      *table,
		Privileges: privs,
	}

	if err := permMgr.RevokeTablePrivileges(opts); err != nil {
		fmt.Printf("Error revoking privileges: %v\n", err)
		os.Exit(1)
	}

	if *table != "" {
		fmt.Printf("Revoked %s on table %s.%s from %s\n", *privileges, *schema, *table, *username)
	} else {
		fmt.Printf("Revoked %s on all tables in schema %s from %s\n", *privileges, *schema, *username)
	}
}

func listPrivilegesCmd() {
	fs := flag.NewFlagSet("list-privileges", flag.ExitOnError)
	username := fs.String("username", "", "Username (required)")

	fs.Parse(os.Args[2:])

	if *username == "" {
		fmt.Println("Error: username is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	permMgr := permission.NewManager(db.DB)

	privileges, err := permMgr.ListUserPrivileges(*username)
	if err != nil {
		fmt.Printf("Error listing privileges: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Privileges for user: %s\n", *username)
	fmt.Println("--------------------------------------------------")
	for _, p := range privileges {
		fmt.Printf("Database: %s, Schema: %s, Table: %s, Privilege: %s\n",
			p["database"], p["schema"], p["table"], p["privilege"])
	}
}

func hbaAddCmd() {
	fs := flag.NewFlagSet("hba-add", flag.ExitOnError)
	connType := fs.String("type", "host", "Connection type (local, host, hostssl, hostnossl)")
	database := fs.String("database", "all", "Database name")
	username := fs.String("user", "", "Username (required)")
	address := fs.String("address", "", "Address/CIDR (required for host types)")
	method := fs.String("method", "scram-sha-256", "Authentication method (md5, scram-sha-256, trust, etc.)")
	hbaFile := fs.String("hba-file", "/etc/postgresql/16/main/pg_hba.conf", "Path to pg_hba.conf")

	fs.Parse(os.Args[2:])

	if *username == "" {
		fmt.Println("Error: user is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	if (*connType == "host" || *connType == "hostssl" || *connType == "hostnossl") && *address == "" {
		fmt.Println("Error: address is required for host connection types")
		fs.PrintDefaults()
		os.Exit(1)
	}

	hbaMgr := hba.NewManager(*hbaFile)

	rule := hba.Rule{
		Type:     hba.ConnectionType(*connType),
		Database: *database,
		User:     *username,
		Address:  *address,
		Method:   hba.AuthMethod(*method),
	}

	if err := hbaMgr.AddRule(rule); err != nil {
		fmt.Printf("Error adding HBA rule: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully added HBA rule")
	fmt.Println("Remember to reload PostgreSQL configuration: pg_user_admin hba-reload")
}

func hbaRemoveCmd() {
	fs := flag.NewFlagSet("hba-remove", flag.ExitOnError)
	connType := fs.String("type", "host", "Connection type")
	database := fs.String("database", "all", "Database name")
	username := fs.String("user", "", "Username (required)")
	address := fs.String("address", "", "Address/CIDR")
	method := fs.String("method", "scram-sha-256", "Authentication method")
	hbaFile := fs.String("hba-file", "/etc/postgresql/16/main/pg_hba.conf", "Path to pg_hba.conf")

	fs.Parse(os.Args[2:])

	if *username == "" {
		fmt.Println("Error: user is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	hbaMgr := hba.NewManager(*hbaFile)

	rule := hba.Rule{
		Type:     hba.ConnectionType(*connType),
		Database: *database,
		User:     *username,
		Address:  *address,
		Method:   hba.AuthMethod(*method),
	}

	if err := hbaMgr.RemoveRule(rule); err != nil {
		fmt.Printf("Error removing HBA rule: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully removed HBA rule")
	fmt.Println("Remember to reload PostgreSQL configuration: pg_user_admin hba-reload")
}

func hbaListCmd() {
	fs := flag.NewFlagSet("hba-list", flag.ExitOnError)
	hbaFile := fs.String("hba-file", "/etc/postgresql/16/main/pg_hba.conf", "Path to pg_hba.conf")

	fs.Parse(os.Args[2:])

	hbaMgr := hba.NewManager(*hbaFile)

	rules, err := hbaMgr.ReadRules()
	if err != nil {
		fmt.Printf("Error reading HBA rules: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PostgreSQL HBA Rules:")
	fmt.Println("==================================================")
	for i, rule := range rules {
		if rule.IsComment {
			fmt.Printf("%d. %s\n", i+1, rule.Comment)
		} else {
			fmt.Printf("%d. TYPE: %s, DATABASE: %s, USER: %s, ADDRESS: %s, METHOD: %s\n",
				i+1, rule.Type, rule.Database, rule.User, rule.Address, rule.Method)
			if rule.Options != "" {
				fmt.Printf("   OPTIONS: %s\n", rule.Options)
			}
		}
	}
}

// hbaReloadCmd reloads the PostgreSQL configuration and verifies that HBA rules are active.
// It connects to the configured database, triggers a configuration reload, and prints a success message or exits on error.
func hbaReloadCmd() {
	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	reloader := hba.NewReloader(db.DB)

	if err := reloader.ReloadAndVerify(); err != nil {
		fmt.Printf("Error reloading PostgreSQL configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully reloaded PostgreSQL configuration")
	fmt.Println("HBA rules are now active")
}

// createRoleCmd parses command-line flags for the create-role command, validates inputs,
// connects to the database, and creates a PostgreSQL role according to the provided options.
// It requires a non-empty rolename, rejects names starting with "pg_", and requires a password
// when the role is created with login capability. On validation or creation failure the command
// prints an error and exits the process; on success it prints the created role type and name.
func createRoleCmd() {
	fs := flag.NewFlagSet("create-role", flag.ExitOnError)
	roleName := fs.String("rolename", "", "Role name to create (required)")
	canLogin := fs.Bool("login", false, "Allow login (creates a user role)")
	password := fs.String("password", "", "Password (only if login is true)")
	isSuperuser := fs.Bool("superuser", false, "Create as superuser")

	fs.Parse(os.Args[2:])

	if *roleName == "" {
		fmt.Println("Error: rolename is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	// Reject role names starting with pg_ (PostgreSQL system prefix, case-insensitive)
	if strings.HasPrefix(strings.ToLower(*roleName), "pg_") {
		fmt.Printf("Error: role names starting with 'pg_' (case-insensitive) are reserved for PostgreSQL system roles\n")
		os.Exit(1)
	}

	if *canLogin && *password == "" {
		fmt.Println("Error: password is required when login is enabled")
		fs.PrintDefaults()
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	roleMgr := role.NewManager(db.DB)

	opts := role.RoleOptions{
		RoleName:    *roleName,
		CanLogin:    *canLogin,
		Password:    *password,
		IsSuperuser: *isSuperuser,
	}

	if err := roleMgr.CreateRole(opts); err != nil {
		fmt.Printf("Error creating role: %v\n", err)
		os.Exit(1)
	}

	roleType := "group role (NOLOGIN)"
	if *canLogin {
		roleType = "user role (LOGIN)"
	}
	fmt.Printf("Successfully created %s: %s\n", roleType, *roleName)
}

// deleteRoleCmd parses command-line flags for the delete-role command and deletes the specified PostgreSQL role.
// It validates required inputs and prevents reassigning objects to the same role being removed. If `--reassign-to`
// or `--drop-owned` are provided, it performs a cleanup deletion that reassigns or drops owned objects; otherwise it
// performs a standard deletion which may fail if the role owns objects or has members. On errors it prints a message
// and exits the process with a non-zero status.
func deleteRoleCmd() {
	fs := flag.NewFlagSet("delete-role", flag.ExitOnError)
	roleName := fs.String("rolename", "", "Role name to delete (required)")
	reassignTo := fs.String("reassign-to", "", "Reassign owned objects to this role before deletion")
	dropOwned := fs.Bool("drop-owned", false, "Drop all objects owned by the role before deletion")

	fs.Parse(os.Args[2:])

	if *roleName == "" {
		fmt.Println("Error: rolename is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	// Validate that reassign-to is different from the role being deleted
	if *reassignTo != "" && *reassignTo == *roleName {
		fmt.Printf("Error: cannot reassign objects to the same role being deleted (%s)\n", *roleName)
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	roleMgr := role.NewManager(db.DB)

	// Use safe deletion if reassign-to or drop-owned is specified
	if *reassignTo != "" || *dropOwned {
		opts := role.DeleteRoleOptions{
			RoleName:   *roleName,
			ReassignTo: *reassignTo,
			DropOwned:  *dropOwned,
		}
		if err := roleMgr.DeleteRoleWithOptions(opts); err != nil {
			fmt.Printf("Error deleting role: %v\n", err)
			fmt.Println("\nHint: Check role members with:")
			fmt.Printf("  pg_user_admin list-role-members -rolename=%s\n", *roleName)
			os.Exit(1)
		}
		fmt.Printf("Successfully deleted role: %s (with cleanup)\n", *roleName)
	} else {
		// Simple deletion (may fail if role owns objects)
		if err := roleMgr.DeleteRole(*roleName); err != nil {
			fmt.Printf("Error deleting role: %v\n", err)
			fmt.Println("\nHint: If the role owns objects or has members, use:")
			fmt.Println("  --reassign-to <role>  to reassign owned objects")
			fmt.Println("  --drop-owned          to drop owned objects")
			fmt.Println("  list-role-members     to find and revoke all members first")
			os.Exit(1)
		}
		fmt.Printf("Successfully deleted role: %s\n", *roleName)
	}
}

// listRolesCmd lists PostgreSQL roles and prints a formatted summary to standard output.
// It connects to the configured database, retrieves roles via the role manager, and prints each role's name, whether it can log in, whether it is a superuser, and its connection limit.
// If a connection or listing error occurs the function prints the error and exits the process with status 1.
func listRolesCmd() {
	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	roleMgr := role.NewManager(db.DB)

	roles, err := roleMgr.ListRoles()
	if err != nil {
		fmt.Printf("Error listing roles: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PostgreSQL Roles:")
	fmt.Println("--------------------------------------------------")
	for _, r := range roles {
		fmt.Printf("Role: %s\n", r["role_name"])
		fmt.Printf("  Can Login: %v\n", r["can_login"])
		fmt.Printf("  Superuser: %v\n", r["is_superuser"])
		fmt.Printf("  Connection Limit: %v\n", r["connection_limit"])
		fmt.Println()
	}
}

// grantRoleCmd parses command-line flags and grants a PostgreSQL role to a user.
// 
// It requires the `-rolename` and `-username` flags, validates their presence,
// establishes a database connection, and uses the role manager to grant the
// specified role to the specified user. On error it prints a message and exits;
// on success it prints a confirmation.
func grantRoleCmd() {
	fs := flag.NewFlagSet("grant-role", flag.ExitOnError)
	roleName := fs.String("rolename", "", "Role name to grant (required)")
	username := fs.String("username", "", "Username to receive the role (required)")

	fs.Parse(os.Args[2:])

	if *roleName == "" || *username == "" {
		fmt.Println("Error: both rolename and username are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	roleMgr := role.NewManager(db.DB)

	if err := roleMgr.GrantRole(*roleName, *username); err != nil {
		fmt.Printf("Error granting role: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully granted role %s to user %s\n", *roleName, *username)
}

// revokeRoleCmd revokes a PostgreSQL role from a specified user using command-line flags.
// The command requires the "rolename" and "username" flags; it connects to the database,
// attempts the revocation, exits with a non-zero status and prints an error on failure,
// and prints a confirmation message on success.
func revokeRoleCmd() {
	fs := flag.NewFlagSet("revoke-role", flag.ExitOnError)
	roleName := fs.String("rolename", "", "Role name to revoke (required)")
	username := fs.String("username", "", "Username to revoke from (required)")

	fs.Parse(os.Args[2:])

	if *roleName == "" || *username == "" {
		fmt.Println("Error: both rolename and username are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	roleMgr := role.NewManager(db.DB)

	if err := roleMgr.RevokeRole(*roleName, *username); err != nil {
		fmt.Printf("Error revoking role: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully revoked role %s from user %s\n", *roleName, *username)
}

// listRoleMembersCmd parses the "rolename" flag, queries the database for members of that role, and prints the member list to stdout.
// It prints "No members found" when the role has no members. If the required flag is missing or a database operation fails, it prints an error and exits with status 1.
func listRoleMembersCmd() {
	fs := flag.NewFlagSet("list-role-members", flag.ExitOnError)
	roleName := fs.String("rolename", "", "Role name (required)")

	fs.Parse(os.Args[2:])

	if *roleName == "" {
		fmt.Println("Error: rolename is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	roleMgr := role.NewManager(db.DB)

	members, err := roleMgr.ListRoleMembers(*roleName)
	if err != nil {
		fmt.Printf("Error listing role members: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Members of role '%s':\n", *roleName)
	fmt.Println("--------------------------------------------------")
	if len(members) == 0 {
		fmt.Println("No members found")
	} else {
		for _, member := range members {
			fmt.Printf("  - %s\n", member)
		}
	}
}

// listUserRolesCmd lists roles granted to a specified database user.
// It parses the "username" flag, validates it, connects to the configured database, queries the user's granted roles, and prints them to stdout.
// On validation or runtime errors it prints an error message and exits with status 1.
func listUserRolesCmd() {
	fs := flag.NewFlagSet("list-user-roles", flag.ExitOnError)
	username := fs.String("username", "", "Username (required)")

	fs.Parse(os.Args[2:])

	if *username == "" {
		fmt.Println("Error: username is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	_, db, err := connectDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	roleMgr := role.NewManager(db.DB)

	roles, err := roleMgr.ListUserRoles(*username)
	if err != nil {
		fmt.Printf("Error listing user roles: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Roles granted to user '%s':\n", *username)
	fmt.Println("--------------------------------------------------")
	if len(roles) == 0 {
		fmt.Println("No roles found")
	} else {
		for _, role := range roles {
			fmt.Printf("  - %s\n", role)
		}
	}
}

// connectDB creates a Config from the environment and opens a database connection using it.
// It returns the Config, the connected *database.DB, and any error encountered while creating
// the configuration or establishing the connection.
func connectDB() (*config.Config, *database.DB, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, nil, err
	}

	db, err := database.Connect(cfg.ConnectionString())
	if err != nil {
		return nil, nil, err
	}

	return cfg, db, nil
}

// printUsage prints the CLI help text describing available commands, examples, and relevant environment variables.
// The output includes grouped command categories (user, role, permission, and HBA management), example invocations, and environment variable defaults.
func printUsage() {
	fmt.Println("PostgreSQL User Management Admin Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  pg_user_admin <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  User Management:")
	fmt.Println("    create-user       Create a new PostgreSQL user")
	fmt.Println("    modify-user       Modify an existing PostgreSQL user")
	fmt.Println("    delete-user       Delete a PostgreSQL user")
	fmt.Println("    list-users        List all PostgreSQL users")
	fmt.Println("")
	fmt.Println("  Role Management (PostgreSQL's native permission system):")
	fmt.Println("    create-role       Create a new role (group)")
	fmt.Println("    delete-role       Delete a role (use --reassign-to or --drop-owned for safe cleanup)")
	fmt.Println("    list-roles        List all roles")
	fmt.Println("    grant-role        Grant a role to a user")
	fmt.Println("    revoke-role       Revoke a role from a user")
	fmt.Println("    list-role-members List all members of a role")
	fmt.Println("    list-user-roles   List all roles granted to a user")
	fmt.Println("")
	fmt.Println("  Permission Management:")
	fmt.Println("    grant             Grant privileges to a user or role")
	fmt.Println("    revoke            Revoke privileges from a user or role")
	fmt.Println("    list-privileges   List privileges for a user or role")
	fmt.Println("")
	fmt.Println("  HBA Configuration:")
	fmt.Println("    hba-add           Add a rule to pg_hba.conf")
	fmt.Println("    hba-remove        Remove a rule from pg_hba.conf")
	fmt.Println("    hba-list          List all rules in pg_hba.conf")
	fmt.Println("    hba-reload        Reload PostgreSQL configuration")
	fmt.Println("")
	fmt.Println("  help              Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Create a regular user with scram-sha-256 authentication")
	fmt.Println("  pg_user_admin create-user -username=myuser -password=secret -auth=scram-sha-256")
	fmt.Println("")
	fmt.Println("  # Create a superuser")
	fmt.Println("  pg_user_admin create-user -username=admin -password=secret -superuser")
	fmt.Println("")
	fmt.Println("  # Grant SELECT and INSERT on a specific table")
	fmt.Println("  pg_user_admin grant -username=myuser -schema=public -table=users -privileges=SELECT,INSERT")
	fmt.Println("")
	fmt.Println("  # Grant all privileges on all tables in a schema")
	fmt.Println("  pg_user_admin grant -username=myuser -schema=public -privileges=ALL")
	fmt.Println("")
	fmt.Println("  # Grant function access and table privileges")
	fmt.Println("  pg_user_admin grant -username=myuser -schema=public -privileges=SELECT -grant-functions")
	fmt.Println("")
	fmt.Println("  # Create a group role and grant it to users (PostgreSQL standard)")
	fmt.Println("  pg_user_admin create-role -rolename=app_readonly")
	fmt.Println("  pg_user_admin grant -username=app_readonly -schema=public -privileges=SELECT -grant-functions")
	fmt.Println("  pg_user_admin grant-role -rolename=app_readonly -username=user1")
	fmt.Println("  pg_user_admin grant-role -rolename=app_readonly -username=user2")
	fmt.Println("")
	fmt.Println("  # Safely delete a role (reassign or drop owned objects)")
	fmt.Println("  pg_user_admin delete-role -rolename=old_role -reassign-to=postgres")
	fmt.Println("  pg_user_admin delete-role -rolename=old_role -drop-owned")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  PGHOST         PostgreSQL host (default: localhost)")
	fmt.Println("  PGPORT         PostgreSQL port (default: 5432)")
	fmt.Println("  PGUSER         PostgreSQL admin user (default: postgres)")
	fmt.Println("  PGPASSWORD     PostgreSQL admin password (required)")
	fmt.Println("  PGDATABASE     PostgreSQL database (default: postgres)")
	fmt.Println("  PGSSLMODE      SSL mode (default: disable)")
}