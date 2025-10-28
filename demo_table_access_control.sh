#!/bin/bash

# Demo: Table-Specific Access Control in Public Schema
# This shows how to grant access to specific tables only, while keeping all functions accessible

set -e

echo "=========================================="
echo "Table-Specific Access Control Demo"
echo "=========================================="
echo ""

# Setup: Assume we have these tables in public schema:
# - public.users (allowed)
# - public.orders (allowed)
# - public.secrets (NOT allowed)
# - public.audit_logs (NOT allowed)

echo "Step 1: Create user with SCRAM-SHA-256 authentication"
echo "-----------------------------------------------"
echo "./pg_user_admin create-user \\"
echo "  -username limited_user \\"
echo "  -password 'SecurePass123' \\"
echo "  -auth scram-sha-256"
echo ""

echo "Step 2: Grant access to SPECIFIC tables only (not all tables)"
echo "-----------------------------------------------"
echo ""
echo "# Grant access to 'users' table"
echo "./pg_user_admin grant \\"
echo "  -username limited_user \\"
echo "  -schema public \\"
echo "  -table users \\"
echo "  -privileges SELECT,INSERT,UPDATE"
echo ""
echo "# Grant access to 'orders' table"
echo "./pg_user_admin grant \\"
echo "  -username limited_user \\"
echo "  -schema public \\"
echo "  -table orders \\"
echo "  -privileges SELECT,INSERT"
echo ""
echo "NOTE: We are NOT granting on 'secrets' or 'audit_logs' tables!"
echo ""

echo "Step 3: Grant function execution privileges"
echo "-----------------------------------------------"
echo "./pg_user_admin grant \\"
echo "  -username limited_user \\"
echo "  -schema public \\"
echo "  -grant-functions"
echo ""

echo "Step 4: Add pg_hba.conf authentication rule"
echo "-----------------------------------------------"
echo "./pg_user_admin hba-add \\"
echo "  -type host \\"
echo "  -database postgres \\"
echo "  -user limited_user \\"
echo "  -address 127.0.0.1/32 \\"
echo "  -method scram-sha-256"
echo ""
echo "./pg_user_admin hba-reload"
echo ""

echo "=========================================="
echo "Result: What limited_user CAN and CANNOT do"
echo "=========================================="
echo ""

echo "✅ CAN access these tables:"
echo "  - SELECT, INSERT, UPDATE on public.users"
echo "  - SELECT, INSERT on public.orders"
echo ""

echo "✅ CAN execute all functions:"
echo "  - public.calculate_total()"
echo "  - public.format_date()"
echo "  - public.any_other_function()"
echo ""

echo "❌ CANNOT access these tables:"
echo "  - public.secrets (ERROR: permission denied)"
echo "  - public.audit_logs (ERROR: permission denied)"
echo "  - public.any_other_table (ERROR: permission denied)"
echo ""

echo "=========================================="
echo "SQL Queries Generated Behind the Scenes"
echo "=========================================="
echo ""
echo "-- Schema access (required for any table access)"
echo "GRANT USAGE ON SCHEMA public TO limited_user;"
echo ""
echo "-- Specific table access only"
echo "GRANT SELECT, INSERT, UPDATE ON TABLE public.users TO limited_user;"
echo "GRANT SELECT, INSERT ON TABLE public.orders TO limited_user;"
echo ""
echo "-- Function execution (ALL functions)"
echo "GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO limited_user;"
echo "ALTER DEFAULT PRIVILEGES IN SCHEMA public"
echo "  GRANT EXECUTE ON FUNCTIONS TO limited_user;"
echo ""

echo "=========================================="
echo "Verification"
echo "=========================================="
echo ""
echo "./pg_user_admin list-privileges -username limited_user"
echo ""
echo "Expected output:"
echo "  Table: public.users - SELECT, INSERT, UPDATE"
echo "  Table: public.orders - SELECT, INSERT"
echo "  Functions: EXECUTE on ALL FUNCTIONS in public schema"
echo ""

echo "=========================================="
echo "Key Points"
echo "=========================================="
echo ""
echo "1. Use -table flag to grant on SPECIFIC table"
echo "2. Omit -table to grant on ALL tables (not what we want here)"
echo "3. -grant-functions works independently of table grants"
echo "4. Combine both for precise control"
echo "5. Repeat grant command for each table you want to allow"
echo ""

echo "This feature has been available since the initial implementation!"
