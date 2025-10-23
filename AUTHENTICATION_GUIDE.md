# PostgreSQL 인증 가이드

## 인증의 두 가지 측면

PostgreSQL 인증은 두 가지 설정이 함께 작동합니다:

### 1. 비밀번호 저장 방식 (사용자 생성 시)
사용자를 생성할 때 `-auth` 플래그로 비밀번호가 데이터베이스에 **저장되는 방식**을 지정합니다.

```bash
# MD5로 비밀번호 저장
./pg_user_admin create-user -username=myuser -password=secret -auth=md5

# SCRAM-SHA-256로 비밀번호 저장 (권장)
./pg_user_admin create-user -username=myuser -password=secret -auth=scram-sha-256
```

**데이터베이스 내부:**
- MD5: `md526baf277223da9f6f4b495734a30bfc9`
- SCRAM-SHA-256: `SCRAM-SHA-256$4096:VGvBeojL16YlGakuy1TcrA==$...`

### 2. 인증 프로토콜 (pg_hba.conf)
클라이언트가 접속할 때 **어떤 인증 방법**을 사용할지 pg_hba.conf에서 지정합니다.

```bash
# pg_hba.conf에 규칙 추가
./pg_user_admin hba-add -user=myuser -database=mydb -address=192.168.1.0/24 -method=scram-sha-256
```

## 호환성 매트릭스

| 저장된 비밀번호 | pg_hba.conf 설정 | 작동 여부 |
|----------------|-----------------|----------|
| MD5            | md5             | ✅ 작동   |
| MD5            | scram-sha-256   | ❌ 실패   |
| SCRAM-SHA-256  | scram-sha-256   | ✅ 작동   |
| SCRAM-SHA-256  | md5             | ❌ 실패   |
| 둘 다          | trust           | ✅ 작동 (비밀번호 무시) |
| 둘 다          | peer            | ✅ 작동 (OS 인증) |

## 완전한 사용 예제

### 예제 1: SCRAM-SHA-256 사용자 (권장)

```bash
# 1. 사용자 생성 (SCRAM-SHA-256로 비밀번호 저장)
export PGHOST=127.0.0.1 PGPORT=5432 PGUSER=postgres PGPASSWORD=admin PGDATABASE=postgres PGSSLMODE=disable
./pg_user_admin create-user -username=appuser -password=securepass123 -auth=scram-sha-256

# 2. 권한 부여
./pg_user_admin grant -username=appuser -database=myapp -schema=public -privileges=ALL -grant-functions

# 3. 네트워크 접근 허용 (pg_hba.conf에 SCRAM-SHA-256 규칙 추가)
./pg_user_admin hba-add -user=appuser -database=myapp -address=10.0.0.0/8 -method=scram-sha-256

# 4. 설정 리로드
./pg_user_admin hba-reload

# 5. 접속 테스트
PGPASSWORD=securepass123 psql -h 10.0.0.5 -U appuser -d myapp
# ✅ 성공! (비밀번호가 SCRAM-SHA-256로 저장되었고, pg_hba.conf도 scram-sha-256)
```

### 예제 2: MD5 사용자 (레거시)

```bash
# 1. 사용자 생성 (MD5로 비밀번호 저장)
./pg_user_admin create-user -username=oldapp -password=oldpass123 -auth=md5

# 2. 권한 부여
./pg_user_admin grant -username=oldapp -database=olddb -schema=public -privileges=ALL

# 3. 네트워크 접근 허용 (pg_hba.conf에 MD5 규칙 추가)
./pg_user_admin hba-add -user=oldapp -database=olddb -address=192.168.1.0/24 -method=md5

# 4. 설정 리로드
./pg_user_admin hba-reload

# 5. 접속 테스트
PGPASSWORD=oldpass123 psql -h 192.168.1.10 -U oldapp -d olddb
# ✅ 성공! (비밀번호가 MD5로 저장되었고, pg_hba.conf도 md5)
```

### 예제 3: 로컬 접속 (peer 인증)

```bash
# 1. 사용자 생성 (비밀번호는 저장되지만 사용되지 않음)
./pg_user_admin create-user -username=dbadmin -password=unused -auth=scram-sha-256

# 2. HBA 규칙 추가 (local 연결에 peer 인증)
./pg_user_admin hba-add -user=dbadmin -database=all -method=peer -type=local

# 3. 설정 리로드
./pg_user_admin hba-reload

# 4. 접속 테스트 (OS 사용자 이름과 PostgreSQL 사용자 이름이 같아야 함)
psql -U dbadmin -d postgres
# ✅ 성공! (peer 인증은 OS 인증을 사용하므로 비밀번호 불필요)
```

## 일반적인 오류와 해결

### 오류 1: "FATAL: password authentication failed"
```
원인: 저장된 비밀번호 방식과 pg_hba.conf의 인증 방법이 불일치
해결:
- 사용자를 다시 생성하거나
- pg_hba.conf 규칙을 수정
```

### 오류 2: "no pg_hba.conf entry for host"
```
원인: pg_hba.conf에 해당 호스트/사용자/데이터베이스 조합에 대한 규칙이 없음
해결:
./pg_user_admin hba-add -user=username -database=dbname -address=client_ip/32 -method=scram-sha-256
./pg_user_admin hba-reload
```

### 오류 3: "SCRAM authentication requires libpq version 10 or above"
```
원인: 클라이언트 라이브러리가 SCRAM-SHA-256를 지원하지 않음
해결:
- PostgreSQL 클라이언트를 10 이상으로 업그레이드하거나
- MD5로 다시 설정
```

## 보안 권장사항

### ✅ 권장: SCRAM-SHA-256
- 더 안전한 해시 알고리즘
- Salt 및 iteration 사용
- PostgreSQL 10+ 지원

```bash
./pg_user_admin create-user -username=user -password=pass -auth=scram-sha-256
./pg_user_admin hba-add -user=user -database=db -address=x.x.x.x/x -method=scram-sha-256
```

### ⚠️ 레거시: MD5
- 구형 애플리케이션과의 호환성을 위해서만 사용
- 가능하면 SCRAM-SHA-256로 마이그레이션

```bash
./pg_user_admin create-user -username=user -password=pass -auth=md5
./pg_user_admin hba-add -user=user -database=db -address=x.x.x.x/x -method=md5
```

### 🔒 가장 안전: SSL/TLS + SCRAM-SHA-256
```bash
# 사용자 생성
./pg_user_admin create-user -username=secureuser -password=strongpass -auth=scram-sha-256

# SSL 필수 연결만 허용
./pg_user_admin hba-add -user=secureuser -database=securedb -address=0.0.0.0/0 -method=scram-sha-256 -type=hostssl
```

## 마이그레이션: MD5 → SCRAM-SHA-256

기존 MD5 사용자를 SCRAM-SHA-256로 마이그레이션:

```bash
# 1. 사용자 비밀번호를 SCRAM-SHA-256로 재설정
./pg_user_admin modify-user -username=olduser -password=newpass -auth=scram-sha-256

# 2. pg_hba.conf 규칙 업데이트
./pg_user_admin hba-remove -user=olduser -database=olddb -address=x.x.x.x/x -method=md5
./pg_user_admin hba-add -user=olduser -database=olddb -address=x.x.x.x/x -method=scram-sha-256

# 3. 설정 리로드
./pg_user_admin hba-reload

# 4. 애플리케이션 설정 업데이트 및 재시작
```

## 확인 방법

### 저장된 비밀번호 해시 확인
```sql
SELECT usename, passwd FROM pg_shadow WHERE usename = 'myuser';
```

### 현재 HBA 규칙 확인
```bash
./pg_user_admin hba-list
```

### 인증 시도 로그 확인
```bash
tail -f /var/log/postgresql/postgresql-16-main.log
```

## 요약

1. **사용자 생성 시**: `-auth` 플래그로 비밀번호 저장 방식 지정
2. **네트워크 접근 시**: `hba-add -method`로 인증 프로토콜 지정
3. **두 가지가 호환**되어야 인증 성공
4. **SCRAM-SHA-256 권장**: 더 안전하고 현대적인 방법
5. **항상 리로드**: HBA 규칙 변경 후 `hba-reload` 실행
