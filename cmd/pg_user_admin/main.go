package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/seonghobae/pg_user_management/internal/auth"
	"github.com/seonghobae/pg_user_management/internal/database"
	"github.com/seonghobae/pg_user_management/internal/permission"
	"github.com/seonghobae/pg_user_management/internal/user"
	"github.com/seonghobae/pg_user_management/pkg/config"
)

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

func printUsage() {
	fmt.Println("PostgreSQL User Management Admin Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  pg_user_admin <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  create-user       Create a new PostgreSQL user")
	fmt.Println("  modify-user       Modify an existing PostgreSQL user")
	fmt.Println("  delete-user       Delete a PostgreSQL user")
	fmt.Println("  list-users        List all PostgreSQL users")
	fmt.Println("  grant             Grant privileges to a user")
	fmt.Println("  revoke            Revoke privileges from a user")
	fmt.Println("  list-privileges   List privileges for a user")
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
	fmt.Println("Environment Variables:")
	fmt.Println("  PGHOST         PostgreSQL host (default: localhost)")
	fmt.Println("  PGPORT         PostgreSQL port (default: 5432)")
	fmt.Println("  PGUSER         PostgreSQL admin user (default: postgres)")
	fmt.Println("  PGPASSWORD     PostgreSQL admin password (required)")
	fmt.Println("  PGDATABASE     PostgreSQL database (default: postgres)")
	fmt.Println("  PGSSLMODE      SSL mode (default: disable)")
}
