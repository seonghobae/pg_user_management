# PostgreSQL Role-Based Permission Management

## Overview

PostgreSQL의 표준 권한 관리 방식은 **ROLE (역할)** 기반입니다. 이 방식은 개별 사용자에게 직접 권한을 부여하는 것보다 훨씬 효율적이고 관리하기 쉽습니다.

## Why Role-Based Permissions?

### ❌ 기존 방식의 문제점 (테이블 단위 직접 부여)

```bash
# 각 사용자마다 권한을 일일이 부여해야 함
./pg_user_admin grant -username=user1 -table=users -privileges=SELECT
./pg_user_admin grant -username=user1 -table=orders -privileges=SELECT
./pg_user_admin grant -username=user2 -table=users -privileges=SELECT
./pg_user_admin grant -username=user2 -table=orders -privileges=SELECT
./pg_user_admin grant -username=user3 -table=users -privileges=SELECT
./pg_user_admin grant -username=user3 -table=orders -privileges=SELECT
```

**문제:**
- 사용자 추가/제거 시 모든 권한을 다시 설정해야 함
- 권한 변경 시 모든 사용자에 대해 반복해야 함
- 권한 구조가 불명확함
- 유지보수가 어려움

### ✅ PostgreSQL 표준 방식 (ROLE 기반)

```bash
# 1. Group role 생성 (한 번만)
./pg_user_admin create-role -rolename=app_readonly

# 2. Role에 권한 부여 (한 번만)
./pg_user_admin grant -username=app_readonly -schema=public -privileges=SELECT -grant-functions

# 3. 사용자를 role에 추가 (간단!)
./pg_user_admin grant-role -rolename=app_readonly -username=user1
./pg_user_admin grant-role -rolename=app_readonly -username=user2
./pg_user_admin grant-role -rolename=app_readonly -username=user3
```

**장점:**
- ✅ 권한 관리가 중앙화됨
- ✅ 새 사용자 추가가 간단함
- ✅ 권한 변경 시 role만 수정하면 모든 멤버에 적용됨
- ✅ 권한 구조가 명확함
- ✅ PostgreSQL 표준 방식

## Complete Example: Read-Only Application Role

### Step 1: Create Group Role

```bash
# Group role 생성 (NOLOGIN - 직접 로그인 불가)
./pg_user_admin create-role -rolename=app_readonly
```

### Step 2: Grant Permissions to Role

```bash
# Database 접근 권한
./pg_user_admin grant -username=app_readonly -database=mydb

# Schema 사용 권한 + 모든 테이블 SELECT + 모든 함수 실행
./pg_user_admin grant \
  -username=app_readonly \
  -schema=public \
  -privileges=SELECT \
  -grant-functions
```

**생성되는 SQL:**
```sql
-- Schema access
GRANT USAGE ON SCHEMA public TO app_readonly;

-- All tables (current and future)
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO app_readonly;

-- All functions (current and future)
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO app_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO app_readonly;
```

### Step 3: Create Users and Assign Role

```bash
# User 생성
./pg_user_admin create-user -username=app_user1 -password=pass1 -auth=scram-sha-256
./pg_user_admin create-user -username=app_user2 -password=pass2 -auth=scram-sha-256
./pg_user_admin create-user -username=app_user3 -password=pass3 -auth=scram-sha-256

# Role 부여 (각 사용자가 app_readonly 권한을 상속받음)
./pg_user_admin grant-role -rolename=app_readonly -username=app_user1
./pg_user_admin grant-role -rolename=app_readonly -username=app_user2
./pg_user_admin grant-role -rolename=app_readonly -username=app_user3
```

### Step 4: Verify

```bash
# Role의 멤버 확인
./pg_user_admin list-role-members -rolename=app_readonly

# 특정 사용자의 role 확인
./pg_user_admin list-user-roles -username=app_user1

# 권한 확인
./pg_user_admin list-privileges -username=app_readonly
```

## Common Role Patterns

### Pattern 1: Read-Only Access

```bash
# 1. Create role
./pg_user_admin create-role -rolename=readonly_group

# 2. Grant SELECT on all tables + function execution
./pg_user_admin grant \
  -username=readonly_group \
  -schema=public \
  -privileges=SELECT \
  -grant-functions

# 3. Add users to role
./pg_user_admin grant-role -rolename=readonly_group -username=analyst1
./pg_user_admin grant-role -rolename=readonly_group -username=analyst2
```

### Pattern 2: Read-Write Access

```bash
# 1. Create role
./pg_user_admin create-role -rolename=readwrite_group

# 2. Grant SELECT, INSERT, UPDATE, DELETE
./pg_user_admin grant \
  -username=readwrite_group \
  -schema=public \
  -privileges=SELECT,INSERT,UPDATE,DELETE \
  -grant-functions

# 3. Add users to role
./pg_user_admin grant-role -rolename=readwrite_group -username=app_user1
./pg_user_admin grant-role -rolename=readwrite_group -username=app_user2
```

### Pattern 3: Limited Table Access

```bash
# 1. Create role
./pg_user_admin create-role -rolename=limited_access

# 2. Grant access to specific tables only
./pg_user_admin grant \
  -username=limited_access \
  -schema=public \
  -table=users \
  -privileges=SELECT,INSERT

./pg_user_admin grant \
  -username=limited_access \
  -schema=public \
  -table=orders \
  -privileges=SELECT

# 3. Grant function access
./pg_user_admin grant \
  -username=limited_access \
  -schema=public \
  -grant-functions

# 4. Add users to role
./pg_user_admin grant-role -rolename=limited_access -username=partner_user
```

### Pattern 4: Multi-Schema Access

```bash
# 1. Create role
./pg_user_admin create-role -rolename=multi_schema_access

# 2. Grant access to multiple schemas
./pg_user_admin grant \
  -username=multi_schema_access \
  -schema=public \
  -privileges=SELECT

./pg_user_admin grant \
  -username=multi_schema_access \
  -schema=reporting \
  -privileges=SELECT

./pg_user_admin grant \
  -username=multi_schema_access \
  -schema=analytics \
  -privileges=SELECT

# 3. Add users to role
./pg_user_admin grant-role -rolename=multi_schema_access -username=data_scientist
```

## Role Hierarchy (Advanced)

Role은 계층적으로 구성할 수 있습니다:

```bash
# Base role: 기본 읽기 권한
./pg_user_admin create-role -rolename=base_readonly
./pg_user_admin grant -username=base_readonly -schema=public -privileges=SELECT

# Extended role: 기본 + 쓰기 권한
./pg_user_admin create-role -rolename=extended_readwrite
./pg_user_admin grant-role -rolename=base_readonly -username=extended_readwrite
./pg_user_admin grant -username=extended_readwrite -schema=public -privileges=INSERT,UPDATE

# User는 extended_readwrite만 부여받아도 base_readonly 권한도 상속받음
./pg_user_admin grant-role -rolename=extended_readwrite -username=app_user
```

## Managing Roles

### List All Roles

```bash
./pg_user_admin list-roles
```

Output:

```text
PostgreSQL Roles:
--------------------------------------------------
Role: app_readonly
  Can Login: false
  Superuser: false
  Connection Limit: -1

Role: app_readwrite
  Can Login: false
  Superuser: false
  Connection Limit: -1
```

### List Role Members

```bash
./pg_user_admin list-role-members -rolename=app_readonly
```

Output:

```text
Members of role 'app_readonly':
--------------------------------------------------
  - app_user1
  - app_user2
  - app_user3
```

### List User's Roles

```bash
./pg_user_admin list-user-roles -username=app_user1
```

Output:

```text
Roles granted to user 'app_user1':
--------------------------------------------------
  - app_readonly
  - reporting_access
```

### Revoke Role from User

```bash
./pg_user_admin revoke-role -rolename=app_readonly -username=app_user1
```

### Delete Role

Role 삭제는 다음 3가지 방법이 있습니다:

#### 방법 1: 안전한 삭제 (권장)

```bash
# 1단계: 모든 멤버 제거
./pg_user_admin list-role-members -rolename=app_readonly
./pg_user_admin revoke-role -rolename=app_readonly -username=app_user1
./pg_user_admin revoke-role -rolename=app_readonly -username=app_user2

# 2단계: Role 삭제
./pg_user_admin delete-role -rolename=app_readonly
```

#### 방법 2: 소유 객체를 다른 role에 재할당

```bash
# Role이 소유한 객체(테이블, 함수 등)를 postgres role에 재할당하고 삭제
./pg_user_admin delete-role -rolename=app_readonly -reassign-to=postgres
```

이 방법은 다음 SQL을 실행합니다:
```sql
BEGIN;
  REASSIGN OWNED BY app_readonly TO postgres;
  DROP ROLE app_readonly;
COMMIT;
```

#### 방법 3: 소유 객체를 모두 삭제 (주의!)

```bash
# ⚠️ 경고: Role이 소유한 모든 객체(테이블, 함수 등)가 삭제됩니다!
./pg_user_admin delete-role -rolename=app_readonly -drop-owned
```

이 방법은 다음 SQL을 실행합니다:
```sql
BEGIN;
  DROP OWNED BY app_readonly;
  DROP ROLE app_readonly;
COMMIT;
```

**주의사항:**
- 방법 1이 가장 안전합니다
- 방법 2는 객체를 보존하면서 소유권을 이전합니다
- 방법 3은 모든 객체를 삭제하므로 주의해서 사용하세요

## Best Practices

### 1. Use Group Roles for Common Permissions

❌ **Don't do this:**
```bash
# 100명의 사용자에게 각각 권한 부여
for user in user1 user2 ... user100; do
  ./pg_user_admin grant -username=$user -schema=public -privileges=SELECT
done
```

✅ **Do this instead:**
```bash
# Role에 한 번만 권한 부여
./pg_user_admin create-role -rolename=app_users
./pg_user_admin grant -username=app_users -schema=public -privileges=SELECT

# 사용자들을 role에 추가
for user in user1 user2 ... user100; do
  ./pg_user_admin grant-role -rolename=app_users -username=$user
done
```

### 2. Name Roles Clearly

```bash
# Good: 목적이 명확한 이름
app_readonly
app_readwrite
reporting_analysts
data_scientists
partner_api_access

# Bad: 의미 불명확
role1
group_a
temp_role
```

### 3. Document Role Purposes

각 role의 목적을 문서화하세요:

```bash
# Create role with clear purpose
./pg_user_admin create-role -rolename=reporting_analysts
# Purpose: Analysts who need read-only access to all tables for reporting
```

### 4. Use NOLOGIN for Group Roles

Group role은 직접 로그인하지 않아야 합니다:

```bash
# Correct: NOLOGIN (default)
./pg_user_admin create-role -rolename=app_readonly

# Wrong: LOGIN enabled for group role
./pg_user_admin create-role -rolename=app_readonly -login -password=secret
```

### 5. Regular Audits

정기적으로 role 멤버십을 검토하세요:

```bash
# 모든 role 확인
./pg_user_admin list-roles

# 각 role의 멤버 확인
./pg_user_admin list-role-members -rolename=app_readonly

# 각 사용자의 role 확인
./pg_user_admin list-user-roles -username=app_user1
```

## Migration from Direct Permissions to Roles

기존 직접 권한 방식에서 role 기반으로 마이그레이션:

### Step 1: Identify Permission Patterns

현재 사용자들의 권한을 분석:
```bash
./pg_user_admin list-privileges -username=user1
./pg_user_admin list-privileges -username=user2
./pg_user_admin list-privileges -username=user3
```

### Step 2: Create Roles Based on Patterns

공통 권한 패턴을 role로 생성:
```bash
./pg_user_admin create-role -rolename=pattern1_readonly
./pg_user_admin create-role -rolename=pattern2_readwrite
```

### Step 3: Grant Permissions to Roles

```bash
./pg_user_admin grant -username=pattern1_readonly -schema=public -privileges=SELECT
./pg_user_admin grant -username=pattern2_readwrite -schema=public -privileges=SELECT,INSERT,UPDATE
```

### Step 4: Assign Users to Roles

```bash
./pg_user_admin grant-role -rolename=pattern1_readonly -username=user1
./pg_user_admin grant-role -rolename=pattern2_readwrite -username=user2
```

### Step 5: (Optional) Revoke Direct Permissions

기존 직접 권한을 제거하고 role 권한만 사용:
```bash
./pg_user_admin revoke -username=user1 -schema=public -privileges=SELECT
# Now user1 only has permissions through pattern1_readonly role
```

## Comparison: Direct vs Role-Based

| Aspect | Direct Permissions | Role-Based Permissions |
|--------|-------------------|------------------------|
| **Initial Setup** | Simple | Slightly more complex |
| **Adding Users** | Repeat for each user | One command per user |
| **Modifying Permissions** | Repeat for ALL users | Modify role once |
| **Clarity** | Scattered | Centralized |
| **Scalability** | Poor (100+ users) | Excellent |
| **PostgreSQL Standard** | No | Yes ✅ |
| **Best Practice** | No | Yes ✅ |

## Summary

✅ **Use ROLE-based permissions for:**
- Multiple users with same permissions
- Centralized permission management
- Easy user onboarding/offboarding
- Clear permission structure
- PostgreSQL best practices

❌ **Use direct permissions only for:**
- One-off special cases
- Single user with unique permissions
- Quick testing/debugging

**Rule of thumb:** If 2명 이상의 사용자가 동일한 권한을 가진다면, ROLE을 사용하세요!
