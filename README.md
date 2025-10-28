# PostgreSQL User Management Admin Tool

PostgreSQL 사용자 및 권한을 관리하기 위한 CLI 도구입니다.

## 기능

1. **사용자 관리**
   - 사용자 생성 (일반 사용자 또는 superuser)
   - 사용자 수정 (권한 변경, 비밀번호 변경)
   - 사용자 삭제
   - 사용자 목록 조회

2. **인증 방법 지원**
   - MD5
   - SCRAM-SHA-256 (기본값, 권장)

3. **Role 기반 권한 관리** ⭐ NEW! (PostgreSQL 표준 방식)
   - Group role 생성 및 삭제
   - 사용자에게 role 부여/회수
   - Role 멤버 조회
   - 중앙화된 권한 관리
   - 자세한 내용: [ROLE_BASED_PERMISSIONS.md](ROLE_BASED_PERMISSIONS.md)

4. **권한 관리**
   - 데이터베이스 레벨 권한 (CONNECT)
   - 스키마 레벨 권한 (USAGE)
   - 테이블 레벨 권한 (SELECT, INSERT, UPDATE, DELETE, ALL)
   - 특정 테이블 또는 스키마 내 모든 테이블에 권한 부여
   - 스키마 내 모든 함수에 EXECUTE 권한 자동 부여

5. **pg_hba.conf 관리**
   - HBA 규칙 추가/삭제
   - HBA 규칙 목록 조회
   - PostgreSQL 설정 자동 리로드
   - 백업 자동 생성

6. **세부 권한 제어**
   - 테이블별 세부 권한 설정
   - 함수에 대한 전체 액세스 권한

## 빌드

```bash
go build -o pg_user_admin ./cmd/pg_user_admin
```

크로스 플랫폼 빌드:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o pg_user_admin-linux ./cmd/pg_user_admin

# macOS
GOOS=darwin GOARCH=amd64 go build -o pg_user_admin-macos ./cmd/pg_user_admin

# Windows
GOOS=windows GOARCH=amd64 go build -o pg_user_admin.exe ./cmd/pg_user_admin
```

## 환경 변수 설정

PostgreSQL 연결 정보는 환경 변수로 설정합니다:

```bash
export PGHOST=localhost          # PostgreSQL 호스트 (기본값: localhost)
export PGPORT=5432               # PostgreSQL 포트 (기본값: 5432)
export PGUSER=postgres           # PostgreSQL 관리자 사용자 (기본값: postgres)
export PGPASSWORD=yourpassword   # PostgreSQL 관리자 비밀번호 (필수)
export PGDATABASE=postgres       # PostgreSQL 데이터베이스 (기본값: postgres)
export PGSSLMODE=disable         # SSL 모드 (기본값: disable)
```

## 사용 방법

### 1. 사용자 생성

일반 사용자 생성 (SCRAM-SHA-256 인증):
```bash
./pg_user_admin create-user -username=myuser -password=secret -auth=scram-sha-256
```

Superuser 생성:
```bash
./pg_user_admin create-user -username=admin -password=secret -superuser -auth=scram-sha-256
```

MD5 인증 방식으로 사용자 생성:
```bash
./pg_user_admin create-user -username=myuser -password=secret -auth=md5
```

로그인 불가능한 사용자 생성:
```bash
./pg_user_admin create-user -username=myuser -password=secret -login=false
```

### 2. 사용자 수정

비밀번호 변경:
```bash
./pg_user_admin modify-user -username=myuser -password=newsecret -auth=scram-sha-256
```

Superuser 권한 부여:
```bash
./pg_user_admin modify-user -username=myuser -superuser
```

Superuser 권한 제거:
```bash
./pg_user_admin modify-user -username=myuser -no-superuser
```

로그인 권한 제거:
```bash
./pg_user_admin modify-user -username=myuser -no-login
```

### 3. 사용자 삭제

```bash
./pg_user_admin delete-user -username=myuser
```

### 4. 사용자 목록 조회

```bash
./pg_user_admin list-users
```

### 5. 권한 부여 (GRANT)

특정 테이블에 SELECT, INSERT 권한 부여:
```bash
./pg_user_admin grant -username=myuser -schema=public -table=users -privileges=SELECT,INSERT
```

스키마 내 모든 테이블에 SELECT 권한 부여:
```bash
./pg_user_admin grant -username=myuser -schema=public -privileges=SELECT
```

스키마 내 모든 테이블에 모든 권한 부여:
```bash
./pg_user_admin grant -username=myuser -schema=public -privileges=ALL
```

스키마 내 모든 함수에 EXECUTE 권한 부여 및 테이블 권한 부여:
```bash
./pg_user_admin grant -username=myuser -schema=public -privileges=SELECT,INSERT,UPDATE,DELETE -grant-functions
```

데이터베이스 접속 권한 부여:
```bash
./pg_user_admin grant -username=myuser -database=mydb
```

### 6. 권한 회수 (REVOKE)

특정 테이블에서 권한 회수:
```bash
./pg_user_admin revoke -username=myuser -schema=public -table=users -privileges=INSERT,UPDATE
```

스키마 내 모든 테이블에서 권한 회수:
```bash
./pg_user_admin revoke -username=myuser -schema=public -privileges=ALL
```

### 7. 사용자 권한 조회

```bash
./pg_user_admin list-privileges -username=myuser
```

### 8. Role 기반 권한 관리 (PostgreSQL 표준 방식)

**Group role 생성 (NOLOGIN - 권한 그룹용):**
```bash
./pg_user_admin create-role -rolename=app_readonly
```

**Role에 권한 부여:**
```bash
# 스키마의 모든 테이블에 SELECT 권한 + 모든 함수 실행 권한
./pg_user_admin grant -username=app_readonly -schema=public -privileges=SELECT -grant-functions
```

**사용자에게 role 부여 (사용자가 role의 모든 권한 상속):**
```bash
./pg_user_admin grant-role -rolename=app_readonly -username=user1
./pg_user_admin grant-role -rolename=app_readonly -username=user2
```

**사용자로부터 role 회수:**
```bash
./pg_user_admin revoke-role -rolename=app_readonly -username=user1
```

**Role 목록 조회:**
```bash
./pg_user_admin list-roles
```

**Role의 멤버 조회:**
```bash
./pg_user_admin list-role-members -rolename=app_readonly
```

**사용자가 속한 role 조회:**
```bash
./pg_user_admin list-user-roles -username=user1
```

**Role 삭제 (안전한 방법):**
```bash
# 방법 1: 멤버 revoke 후 삭제 (가장 안전)
./pg_user_admin revoke-role -rolename=app_readonly -username=user1
./pg_user_admin revoke-role -rolename=app_readonly -username=user2
./pg_user_admin delete-role -rolename=app_readonly

# 방법 2: 소유 객체를 다른 role에 재할당하고 삭제
./pg_user_admin delete-role -rolename=app_readonly -reassign-to=postgres

# 방법 3: 소유 객체를 모두 삭제하고 role 삭제 (주의!)
./pg_user_admin delete-role -rolename=app_readonly -drop-owned
```

💡 **왜 Role을 사용해야 하나요?**
- ✅ 권한 관리가 중앙화됨 (role에만 권한 부여하면 모든 멤버에 적용)
- ✅ 새 사용자 추가가 간단함 (grant-role 한 번만)
- ✅ PostgreSQL 표준 방식
- 자세한 내용: [ROLE_BASED_PERMISSIONS.md](ROLE_BASED_PERMISSIONS.md)

### 9. pg_hba.conf 관리

HBA 규칙 목록 조회:
```bash
./pg_user_admin hba-list
```

HBA 규칙 추가 (특정 네트워크에서 사용자 접속 허용):
```bash
./pg_user_admin hba-add -user=myuser -database=mydb -address=192.168.1.0/24 -method=scram-sha-256 -type=host
```

HBA 규칙 추가 (로컬 연결):
```bash
./pg_user_admin hba-add -user=myuser -database=mydb -method=scram-sha-256 -type=local
```

HBA 규칙 삭제:
```bash
./pg_user_admin hba-remove -user=myuser -database=mydb -address=192.168.1.0/24 -method=scram-sha-256 -type=host
```

PostgreSQL 설정 리로드 (HBA 규칙 활성화):
```bash
./pg_user_admin hba-reload
```

## 일반적인 사용 시나리오

### 시나리오 1: 읽기 전용 사용자 생성

```bash
# 1. 사용자 생성
./pg_user_admin create-user -username=readonly -password=secret -auth=scram-sha-256

# 2. 데이터베이스 접속 권한 부여
./pg_user_admin grant -username=readonly -database=mydb

# 3. 스키마 내 모든 테이블에 SELECT 권한 및 함수 실행 권한 부여
./pg_user_admin grant -username=readonly -schema=public -privileges=SELECT -grant-functions
```

### 시나리오 2: 특정 테이블만 관리할 수 있는 사용자 생성

```bash
# 1. 사용자 생성
./pg_user_admin create-user -username=appuser -password=secret -auth=scram-sha-256

# 2. 데이터베이스 접속 권한 부여
./pg_user_admin grant -username=appuser -database=mydb

# 3. users 테이블에 모든 권한 부여
./pg_user_admin grant -username=appuser -schema=public -table=users -privileges=ALL

# 4. products 테이블에 SELECT, INSERT 권한 부여
./pg_user_admin grant -username=appuser -schema=public -table=products -privileges=SELECT,INSERT

# 5. 모든 함수 실행 권한 부여
./pg_user_admin grant -username=appuser -schema=public -grant-functions
```

### 시나리오 3: 완전한 권한을 가진 관리자 사용자 생성

```bash
# Superuser로 생성
./pg_user_admin create-user -username=dbadmin -password=secret -superuser -auth=scram-sha-256
```

### 시나리오 4: Role 기반 읽기 전용 사용자 그룹 생성 (권장 방식⭐)

```bash
# 1. Group role 생성 (권한 그룹)
./pg_user_admin create-role -rolename=app_readonly

# 2. Role에 권한 부여 (한 번만!)
./pg_user_admin grant -username=app_readonly -database=mydb
./pg_user_admin grant -username=app_readonly -schema=public -privileges=SELECT -grant-functions

# 3. 사용자 생성
./pg_user_admin create-user -username=analyst1 -password=secret1 -auth=scram-sha-256
./pg_user_admin create-user -username=analyst2 -password=secret2 -auth=scram-sha-256
./pg_user_admin create-user -username=analyst3 -password=secret3 -auth=scram-sha-256

# 4. 사용자들을 role에 추가 (간단!)
./pg_user_admin grant-role -rolename=app_readonly -username=analyst1
./pg_user_admin grant-role -rolename=app_readonly -username=analyst2
./pg_user_admin grant-role -rolename=app_readonly -username=analyst3

# 이제 3명의 analyst가 모두 app_readonly의 권한을 상속받습니다!
# 새로운 analyst 추가는 grant-role 한 번만 하면 됩니다.
```

**장점:**
- 새 analyst 추가: `grant-role` 한 번만
- 권한 변경: role에만 수정하면 모든 멤버에 적용
- 권한 구조 명확

## 보안 고려사항

1. **인증 방법**: 가능하면 SCRAM-SHA-256를 사용하세요 (MD5보다 안전)
2. **비밀번호**: 강력한 비밀번호를 사용하세요
3. **최소 권한 원칙**: 필요한 최소한의 권한만 부여하세요
4. **환경 변수**: PGPASSWORD는 안전하게 관리하세요 (.env 파일 사용 권장)

## 프로젝트 구조

```
pg_user_management/
├── cmd/
│   └── pg_user_admin/
│       └── main.go              # CLI 진입점
├── internal/
│   ├── auth/
│   │   └── config.go            # 인증 방법 관리
│   ├── database/
│   │   └── connection.go        # 데이터베이스 연결
│   ├── hba/
│   │   ├── manager.go           # pg_hba.conf 관리
│   │   └── reload.go            # PostgreSQL 설정 리로드
│   ├── permission/
│   │   └── manager.go           # 권한 관리
│   ├── role/
│   │   └── manager.go           # ROLE 관리 (PostgreSQL 표준)
│   └── user/
│       └── manager.go           # 사용자 관리
├── pkg/
│   └── config/
│       └── config.go            # 설정 관리
├── go.mod
└── README.md
```

## 요구사항

- Go 1.16 이상
- PostgreSQL 9.6 이상
- 관리자 권한을 가진 PostgreSQL 계정

## 라이선스

MIT License
