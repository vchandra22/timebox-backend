# Boilerplate Golang

Boilerplate Golang adalah barebone REST API berbasis Go, Gin, PostgreSQL, Redis,
Cobra, Viper, dan Zap. Struktur project dibuat untuk tim yang terbiasa dengan
service-repository pattern, tetapi tetap menjaga boundary ala clean architecture
secara ringan.

## Status Project

Project ini sudah bagus untuk starter/MVP REST API dan cukup rapi untuk
dikembangkan menjadi service production. Baseline runtime hardening sudah ada:
validasi config, CORS allowlist, safe error response, timeout HTTP server,
graceful shutdown, dan health endpoint untuk liveness/readiness.

Project ini belum sepenuhnya production-complete. Hal yang masih perlu
disiapkan: secret/config production, database production, pipeline deploy,
observability, dan backup.

## Tech Stack

- Go 1.26.4
- Gin untuk HTTP router
- Cobra untuk command CLI
- Viper untuk config
- Zap untuk logger
- sqlx + pgx/stdlib untuk PostgreSQL
- go-redis untuk Redis

## Struktur Project

```text
.
|-- cmd/                         # CLI commands
|-- internal/
|   |-- bootstrap/               # init config, logger, database, redis
|   |-- config/                  # config structs
|   |-- dto/                     # request/response DTO per feature
|   |   `-- user/
|   |-- entity/                  # domain entities
|   |-- handler/                 # HTTP handlers/controllers
|   |-- middleware/              # HTTP middleware
|   |-- repository/              # repository registry and contracts
|   |   |-- dbexecutor/          # small sqlx executor wrapper
|   |   |-- health/
|   |   `-- user/
|   |       `-- database/        # PostgreSQL implementation and rows
|   |-- response/                # standard API response helpers
|   |-- router/                  # Gin router setup
|   |-- service/                 # business logic
|   `-- utils/                   # shared small filters/helpers
|-- migrations/                  # SQL migration files
|-- config.yaml                  # local config
|-- go.mod
`-- main.go
```

## Dependency Flow

```text
handler -> dto + entity + service
service -> entity + repository interface
repository interface -> entity + repository error
repository database implementation -> SQL + database row + repository error
```

Rules:

- Handler boleh tahu HTTP, DTO, response helper, dan service.
- Handler hanya memetakan service error ke HTTP status.
- Service tidak boleh import Gin, DTO HTTP, database row, atau driver database.
- Service memetakan repository error ke service error.
- Repository interface hanya expose domain entity dan repository error.
- Database implementation boleh tahu sqlx, query SQL, database row, dan error
  driver database.
- DTO hanya untuk input/output HTTP.
- Entity adalah domain object, tanpa `json` atau `db` tag.
- Dependency boundary dijaga oleh `.go-arch-lint.yml` dan CI.

## Prerequisites

- Go sesuai versi di `go.mod`
- PostgreSQL berjalan sesuai `config.yaml`
- Redis berjalan sesuai `config.yaml`

## Config

Aplikasi membaca config dari `config.yaml` di root project.
Saat binary dijalankan, working directory harus berisi `config.yaml`.
Untuk production, mount file config ke working directory service atau jalankan
service dari folder yang berisi config tersebut.

`app.gin_mode` mengontrol mode Gin. Pakai `release` untuk production agar Gin
tidak mencetak warning debug dan daftar route saat startup.

Default config:

```yaml
app:
  name: timebox-backend
  gin_mode: release
  host: localhost
  port: 8080
  cors_allowed_origins:
    - http://localhost:3000

database:
  pgsql:
    db_name:
      host: localhost
      port: 5432
      username: postgres
      password: postgres
      dbname: timebox_database

redis:
  host: localhost
  port: 6379
  username:
  password:
  dbname:
  dbindex: 0

jwt:
  secret:
  access_ttl_seconds: 900
  refresh_ttl_seconds: 2592000

external:
  aws:
    region:
    access_key_id:
    secret_access_key:
    s3_bucket:
  cloudinary:
    cloud_name:
    api_key:
    api_secret:
  rest_client:
    base_url:
    api_key:
    timeout_seconds: 10
```

## Install Dependencies

```bash
go mod tidy
```

Jika Go cache di environment read-only, gunakan:

```bash
GOCACHE=/tmp/timebox-backend-gocache go mod tidy
```

## Run Application

Root command:

```bash
go run .
```

Run REST API:

```bash
go run . serve
```

Jika perlu custom Go cache:

```bash
GOCACHE=/tmp/timebox-backend-gocache go run . serve
```

Server berjalan di port dari `config.yaml`, default:

```text
http://localhost:8080
```

## Run Checks

```bash
gofmt -w .
go vet ./...
go run github.com/fe3dback/go-arch-lint@v1.15.0 check --project-path .
go build ./...
```

Jika perlu custom Go cache:

```bash
GOCACHE=/tmp/timebox-backend-gocache go build ./...
```

## Production Readiness

Review terakhir:

- Struktur handler-service-repository sudah cukup bersih untuk project kecil.
- Runtime API sudah punya timeout, graceful shutdown, CORS allowlist, dan health
  check liveness/readiness.
- Query database sudah parameterized.
- Format, vet, dan build lokal sudah lolos.

Sudah diimplementasikan:

- Startup memvalidasi `app.port`, `app.cors_allowed_origins`, dan config
  PostgreSQL.
- Redis wajib connect saat startup karena auth menyimpan refresh token di Redis.
- HTTP server memakai read-header, read, write, dan idle timeout.
- Shutdown menangani `SIGINT`/`SIGTERM`, lalu menutup HTTP server dan koneksi
  database secara rapi.
- CORS memakai allowlist dari `app.cors_allowed_origins`, bukan wildcard.
- Handler tidak mengirim raw internal error ke client.
- Path parameter user `:id` divalidasi sebagai UUID sebelum query database.
- Error service user not found dipetakan ke `404`; duplicate email dipetakan
  ke `409`.
- Health endpoint dibagi menjadi liveness dan readiness.
- CI menjalankan `gofmt`, `go vet ./...`, dan `go-arch-lint`.

Masih perlu disiapkan sebelum production sungguhan:

- Untuk production, pastikan `config.yaml` tidak berisi secret yang bocor ke
  source control publik.
- Jalankan migration sebagai langkah deploy yang eksplisit dan terukur.
- Tambahkan observability production: metrics, tracing jika perlu, dan alerting.
- Siapkan backup/restore strategy untuk PostgreSQL.

Batasan teknis saat ini:

- Loader config hanya mencari `config.yaml` di working directory saat proses
  dijalankan.
- Migration runner hanya aman untuk SQL sederhana; jika migration memakai
  function/procedure dengan semicolon di body, gunakan migration tool yang
  parser SQL-nya lebih lengkap.
- Validasi DTO user baru mencakup `required` dan `email`; aturan domain seperti
  minimum panjang nama perlu ditambah di DTO saat dibutuhkan.
Nice to have setelah minimum aman:

- Request ID di middleware logger.
- Metrics endpoint untuk latency, status code, dan database errors.
- Rate limit untuk endpoint publik.
- Dockerfile dan contoh compose untuk local production-like run.

## Migrations

Migration SQL disimpan di:

```text
migrations/
```

Jalankan migration SQL dengan command bawaan:

```bash
go run . migrate
```

Runner membuat tabel metadata berikut jika belum ada:

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

File migration dijalankan urut berdasarkan nama file dan dicatat di
`schema_migrations.version`.

Cara membuat migration baru:

1. Buat file baru di `migrations/` dengan format nama:

```text
migrations/YYYYMMDDHHMM_deskripsi_perubahan.sql
```

2. Pastikan nama file lebih besar dari migration terakhir supaya urutan benar.

3. Tulis SQL perubahan schema. Contoh menambahkan kolom `address` ke tabel
   `users`:

```sql
ALTER TABLE users
ADD COLUMN IF NOT EXISTS address TEXT;
```

Jika kolom wajib diisi, beri default agar data lama tetap valid:

```sql
ALTER TABLE users
ADD COLUMN IF NOT EXISTS address TEXT NOT NULL DEFAULT '';
```

4. Jalankan migration:

```bash
go run . migrate
```

Contoh manual dengan `psql` jika command aplikasi tidak dipakai:

```bash
psql "postgres://postgres:postgres@localhost:5432/timebox_database?sslmode=disable" \
  -f migrations/202606300001_create_users_table.sql
```

## API Endpoints

Base path:

```text
/api/v1
```

Health:

```text
GET /api/v1/health/
GET /api/v1/health/live
GET /api/v1/health/ready
```

`/health/live` hanya memastikan proses HTTP hidup. `/health/ready` dan
`/health/` mengecek kesiapan dependency database.

Auth:

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
```

Register request:

```json
{
  "full_name": "John Doe",
  "email": "john@example.com",
  "password": "Secret123!",
  "timezone": "Asia/Jakarta"
}
```

Login request:

```json
{
  "email": "john@example.com",
  "password": "Secret123!"
}
```

Refresh/logout request:

```json
{
  "refresh_token": "jwt-refresh-token"
}
```

Error response:

```json
{
  "status": false,
  "message": "invalid request",
  "error": "validation error"
}
```

## Cara Menambah Feature Baru

Contoh feature `menu`.

1. Buat entity domain:

```text
internal/entity/menu.go
```

2. Buat DTO:

```text
internal/dto/menu/request.go
internal/dto/menu/response.go
```

3. Buat repository interface:

```text
internal/repository/menu/repository.go
```

4. Buat database implementation:

```text
internal/repository/menu/database/query.go
internal/repository/menu/database/row.go
internal/repository/menu/database/repository.go
```

5. Register repository di:

```text
internal/repository/repository.go
```

6. Buat service:

```text
internal/service/menu.go
```

7. Register service di:

```text
internal/service/service.go
```

8. Buat handler dan route:

```text
internal/handler/menu.go
```

Handler feature memiliki method:

```go
func (h *MenuHandler) RegisterRoutes(routeGroup *gin.RouterGroup)
```

9. Register handler di:

```text
internal/handler/handler.go
internal/router/api.go
```

10. Tambahkan migration:

```text
migrations/YYYYMMDDHHMM_create_menus_table.sql
```

11. Tambahkan component feature ke `.go-arch-lint.yml` jika membuat package
    repository atau database baru, lalu jalankan:

```bash
go run github.com/fe3dback/go-arch-lint@v1.15.0 check --project-path .
```

## Naming Guidelines

- Package gunakan nama pendek dan jelas: `service`, `handler`, `repository`.
- Jangan gunakan `util`, `helper`, atau `common` untuk logic yang spesifik.
- Gunakan `entity` untuk domain object.
- Gunakan `dto` untuk HTTP request/response.
- Gunakan `Row` untuk struct database scan result.
- Gunakan `Repository` untuk interface dan implementation dalam package masing-masing.
- Gunakan `Service` untuk business logic.
- Gunakan `Handler` untuk HTTP delivery.
- Acronym Go ditulis konsisten: `ID`, `DB`, `API`, `JWT`, `SQL`.

## ADR

Architecture Decision Records disimpan di:

```text
docs/adr/
```

Gunakan ADR untuk mencatat keputusan arsitektur yang mempengaruhi struktur,
dependency direction, data mapping, atau standar tim.
