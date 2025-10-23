#!/bin/bash

# PostgreSQL User Management Admin Tool - Example Usage Script
# This script demonstrates how to use pg_user_admin tool

# Environment variables for PostgreSQL connection
# Uncomment and set these before running the script
# export PGHOST=localhost
# export PGPORT=5432
# export PGUSER=postgres
# export PGPASSWORD=yourpassword
# export PGDATABASE=postgres
# export PGSSLMODE=disable

echo "=== PostgreSQL User Management Admin Tool - Example Usage ==="
echo ""

# Example 1: Create a regular user with SCRAM-SHA-256 authentication
echo "Example 1: Creating a regular user 'myuser' with SCRAM-SHA-256 authentication"
echo "Command: ./pg_user_admin create-user -username=myuser -password=secret123 -auth=scram-sha-256"
echo ""

# Example 2: Create a superuser
echo "Example 2: Creating a superuser 'dbadmin'"
echo "Command: ./pg_user_admin create-user -username=dbadmin -password=admin123 -superuser -auth=scram-sha-256"
echo ""

# Example 3: List all users
echo "Example 3: Listing all users"
echo "Command: ./pg_user_admin list-users"
echo ""

# Example 4: Grant database access
echo "Example 4: Granting database access to 'myuser' on 'mydb'"
echo "Command: ./pg_user_admin grant -username=myuser -database=mydb"
echo ""

# Example 5: Grant table privileges
echo "Example 5: Granting SELECT and INSERT on specific table"
echo "Command: ./pg_user_admin grant -username=myuser -schema=public -table=users -privileges=SELECT,INSERT"
echo ""

# Example 6: Grant all privileges on all tables in schema
echo "Example 6: Granting all privileges on all tables in 'public' schema"
echo "Command: ./pg_user_admin grant -username=myuser -schema=public -privileges=ALL"
echo ""

# Example 7: Grant function execution privileges
echo "Example 7: Granting function execution privileges and table access"
echo "Command: ./pg_user_admin grant -username=myuser -schema=public -privileges=SELECT,INSERT,UPDATE,DELETE -grant-functions"
echo ""

# Example 8: Create a read-only user
echo "Example 8: Creating a read-only user"
echo "Commands:"
echo "  ./pg_user_admin create-user -username=readonly -password=readonly123 -auth=scram-sha-256"
echo "  ./pg_user_admin grant -username=readonly -database=mydb"
echo "  ./pg_user_admin grant -username=readonly -schema=public -privileges=SELECT -grant-functions"
echo ""

# Example 9: Modify user
echo "Example 9: Modifying user password"
echo "Command: ./pg_user_admin modify-user -username=myuser -password=newsecret123 -auth=scram-sha-256"
echo ""

# Example 10: Revoke privileges
echo "Example 10: Revoking privileges from user"
echo "Command: ./pg_user_admin revoke -username=myuser -schema=public -table=users -privileges=INSERT,UPDATE"
echo ""

# Example 11: List user privileges
echo "Example 11: Listing privileges for user"
echo "Command: ./pg_user_admin list-privileges -username=myuser"
echo ""

# Example 12: Delete user
echo "Example 12: Deleting user"
echo "Command: ./pg_user_admin delete-user -username=myuser"
echo ""

echo "=== Complete Workflow Example ==="
echo ""
echo "# 1. Create a new application user"
echo "./pg_user_admin create-user -username=appuser -password=apppass123 -auth=scram-sha-256"
echo ""
echo "# 2. Grant database connection"
echo "./pg_user_admin grant -username=appuser -database=myapp"
echo ""
echo "# 3. Grant specific table privileges"
echo "./pg_user_admin grant -username=appuser -schema=public -table=users -privileges=SELECT,INSERT,UPDATE"
echo "./pg_user_admin grant -username=appuser -schema=public -table=products -privileges=SELECT"
echo ""
echo "# 4. Grant all function execution privileges"
echo "./pg_user_admin grant -username=appuser -schema=public -grant-functions"
echo ""
echo "# 5. Verify privileges"
echo "./pg_user_admin list-privileges -username=appuser"
echo ""
