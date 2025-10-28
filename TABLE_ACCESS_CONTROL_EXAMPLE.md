# Table-Specific Access Control in Public Schema

This document demonstrates how to grant access to **specific tables only** in the public schema, while keeping all functions accessible.

## Feature Overview

The tool supports granular table-level access control:
- Grant access to **specific tables** (not all tables in schema)
- Grant access to **all tables** in a schema
- Grant **function execution** privileges independently
- Combine both for precise access control

## Use Case: Limited Table Access with Full Function Access

### Scenario
You want a user to:
- ✅ Access ONLY `public.allowed_table` (not other tables)
- ✅ Execute ALL functions in `public` schema
- ❌ NOT access `public.restricted_table` or other tables

## Step-by-Step Example

### 1. Create User with Authentication Method

```bash
# Create user with SCRAM-SHA-256 authentication
./pg_user_admin create-user \
  -username limited_user \
  -password "SecurePass123" \
  -auth scram-sha-256
```

### 2. Grant Access to Specific Table ONLY

```bash
# Grant SELECT, INSERT on ONLY the allowed_table
./pg_user_admin grant \
  -username limited_user \
  -schema public \
  -table allowed_table \
  -privileges SELECT,INSERT,UPDATE,DELETE
```

**Key Point**: The `-table` flag specifies **one specific table**. Without this flag, it would grant on ALL tables.

### 3. Grant Function Execution Privileges

```bash
# Grant EXECUTE on all functions in public schema
./pg_user_admin grant \
  -username limited_user \
  -schema public \
  -grant-functions
```

### 4. Add pg_hba.conf Rule (if needed)

```bash
# Allow scram-sha-256 authentication for this user
./pg_user_admin hba-add \
  -type host \
  -database postgres \
  -user limited_user \
  -address 127.0.0.1/32 \
  -method scram-sha-256

# Reload PostgreSQL configuration
./pg_user_admin hba-reload
```

## Verification

### What the User CAN Do

```sql
-- ✅ Access allowed_table
SELECT * FROM public.allowed_table;
INSERT INTO public.allowed_table (data) VALUES ('new data');
UPDATE public.allowed_table SET data = 'modified' WHERE id = 1;
DELETE FROM public.allowed_table WHERE id = 1;

-- ✅ Execute functions in public schema
SELECT public.some_function();
```

### What the User CANNOT Do

```sql
-- ❌ Access restricted_table (permission denied)
SELECT * FROM public.restricted_table;
ERROR:  permission denied for table restricted_table

-- ❌ Access other tables in public schema
SELECT * FROM public.other_table;
ERROR:  permission denied for table other_table
```

## Granting Multiple Specific Tables

If you need to grant access to multiple specific tables (but not all), run the grant command multiple times:

```bash
# Grant access to table1
./pg_user_admin grant \
  -username limited_user \
  -schema public \
  -table table1 \
  -privileges SELECT,INSERT

# Grant access to table2
./pg_user_admin grant \
  -username limited_user \
  -schema public \
  -table table2 \
  -privileges SELECT,INSERT

# Grant access to table3
./pg_user_admin grant \
  -username limited_user \
  -schema public \
  -table table3 \
  -privileges SELECT,UPDATE,DELETE

# Functions still work for all
./pg_user_admin grant \
  -username limited_user \
  -schema public \
  -grant-functions
```

## Implementation Details

### Code Reference

The table-specific access control is implemented in:
- **CLI**: `cmd/pg_user_admin/main.go:232` - `-table` flag
- **Core Logic**: `internal/permission/manager.go:54-87`

### How It Works

```go
// When -table is specified:
if opts.Table != "" {
    query := fmt.Sprintf("GRANT %s ON TABLE %s.%s TO %s",
        privileges,
        quoteIdentifier(opts.Schema),
        quoteIdentifier(opts.Table),
        quoteIdentifier(opts.Username))
    // Grants ONLY on the specific table
}

// When -table is NOT specified:
else if opts.Schema != "" {
    query := fmt.Sprintf("GRANT %s ON ALL TABLES IN SCHEMA %s TO %s",
        privileges,
        quoteIdentifier(opts.Schema),
        quoteIdentifier(opts.Username))
    // Grants on ALL tables in the schema
}
```

### SQL Generated

**Specific Table Access:**
```sql
-- Schema access (required)
GRANT USAGE ON SCHEMA public TO limited_user;

-- Specific table only
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.allowed_table TO limited_user;

-- All functions
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO limited_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO limited_user;
```

**Result:**
- User can access `public.allowed_table` with specified privileges
- User can execute any function in `public` schema
- User CANNOT access any other tables in `public` schema

## Revoking Specific Table Access

To revoke access from a specific table:

```bash
./pg_user_admin revoke \
  -username limited_user \
  -schema public \
  -table allowed_table \
  -privileges SELECT,INSERT,UPDATE,DELETE
```

## Listing User Privileges

To see what tables a user has access to:

```bash
./pg_user_admin list-privileges -username limited_user
```

Example output:
```
Privileges for user: limited_user
--------------------------------------------------
Table: public.allowed_table
  Privilege: SELECT
  Grantable: NO

Table: public.allowed_table
  Privilege: INSERT
  Grantable: NO

Table: public.allowed_table
  Privilege: UPDATE
  Grantable: NO

Table: public.allowed_table
  Privilege: DELETE
  Grantable: NO

Function Execute Privileges:
  Schema: public - EXECUTE on ALL FUNCTIONS
```

## Summary

✅ **YES**, the tool fully supports table-specific access control in the public schema (or any schema):

1. Use `-table <table_name>` to grant on **specific table only**
2. Omit `-table` to grant on **all tables** in schema
3. Use `-grant-functions` to give function execution rights
4. Combine both for precise control: specific tables + all functions

This functionality was implemented from the beginning and has been tested with real PostgreSQL 16.
