# Testing Documentation

이 문서는 PostgreSQL User Management Admin Tool의 테스트 구조와 실행 방법을 설명합니다.

## 테스트 구조

### 단위 테스트 (Unit Tests)

각 패키지별로 단위 테스트가 작성되어 있습니다:

- **internal/auth/config_test.go**: 인증 방법 검증 테스트
- **internal/user/manager_test.go**: 사용자 관리 기능 테스트
- **internal/permission/manager_test.go**: 권한 관리 기능 테스트
- **pkg/config/config_test.go**: 설정 관리 테스트
- **internal/database/connection_test.go**: 데이터베이스 연결 테스트 (통합 테스트 필요)

### 통합 테스트 (Integration Tests)

- **test/integration_test.go**: 실제 PostgreSQL 데이터베이스를 사용한 통합 테스트

## 테스트 실행

### 1. 단위 테스트 실행

```bash
# 모든 단위 테스트 실행
make test

# 또는
go test -v ./...
```

### 2. 테스트 커버리지 확인

```bash
# 커버리지 리포트 생성
make test-coverage

# 또는
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

생성된 `coverage.html` 파일을 브라우저로 열어서 상세한 커버리지를 확인할 수 있습니다.

### 3. 통합 테스트 실행

통합 테스트는 실제 PostgreSQL 데이터베이스가 필요합니다.

```bash
# PostgreSQL 환경 변수 설정
export PGHOST=localhost
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=yourpassword
export PGDATABASE=postgres

# 통합 테스트 실행
make test-integration

# 또는
go test -tags=integration -v ./test/
```

## 테스트 커버리지 현황

현재 테스트 커버리지:

- **internal/auth**: 100% ✅
- **pkg/config**: 100% ✅
- **internal/user**: 84.2% ⭐
- **internal/permission**: 63.3% ⭐
- **internal/database**: 0% (통합 테스트 필요)
- **cmd/pg_user_admin**: 0% (CLI 엔드투엔드 테스트 필요)

## 테스트 작성 가이드

### sqlmock을 사용한 데이터베이스 모킹

데이터베이스 작업을 테스트할 때는 `go-sqlmock`을 사용합니다:

```go
import "github.com/DATA-DOG/go-sqlmock"

func TestSomeDatabaseOperation(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to create mock db: %v", err)
    }
    defer db.Close()

    // Mock 설정
    mock.ExpectExec("CREATE USER").WillReturnResult(sqlmock.NewResult(0, 1))

    // 테스트 실행
    manager := user.NewManager(db)
    err = manager.CreateUser(opts)

    // 검증
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("unfulfilled expectations: %v", err)
    }
}
```

### 테스트 케이스 구조

테이블 기반 테스트(Table-Driven Tests)를 사용합니다:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        want        string
        expectError bool
    }{
        {
            name:        "Test case 1",
            input:       "input1",
            want:        "output1",
            expectError: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)

            if tt.expectError {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
                if got != tt.want {
                    t.Errorf("got %v, want %v", got, tt.want)
                }
            }
        })
    }
}
```

## CI/CD 통합

GitHub Actions나 다른 CI 시스템에서 테스트를 실행하려면:

```yaml
# .github/workflows/test.yml 예시
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.19'
      - run: make test
      - run: make test-coverage
```

## 트러블슈팅

### 테스트 실패 시

1. **데이터베이스 연결 오류**
   - PostgreSQL이 실행 중인지 확인
   - 환경 변수가 올바르게 설정되었는지 확인

2. **Mock 기대값 불일치**
   - SQL 쿼리 문자열이 정확히 일치하는지 확인
   - 정규표현식 매칭을 사용할 수 있습니다: `mock.ExpectExec(regexp.QuoteMeta(query))`

3. **환경 변수 문제**
   - 테스트 실행 전에 `os.Clearenv()`를 호출하여 환경 초기화
   - 각 테스트에서 필요한 환경 변수를 명시적으로 설정

## 추가 정보

- [go-sqlmock 문서](https://github.com/DATA-DOG/go-sqlmock)
- [Go 테스트 가이드](https://golang.org/pkg/testing/)
- [테이블 기반 테스트](https://github.com/golang/go/wiki/TableDrivenTests)
