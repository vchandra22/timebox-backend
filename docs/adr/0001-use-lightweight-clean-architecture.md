# 0001. Use Lightweight Clean Architecture

## Status

Accepted

## Context

Project ini adalah backend barebone yang perlu mudah dipakai tim, tetapi tetap
siap tumbuh untuk skala menengah. Full Clean Architecture bisa membuat terlalu
banyak folder dan ceremony untuk project awal.

## Decision

Gunakan clean architecture secara ringan:

```text
handler -> service -> repository interface -> repository implementation
```

Dengan boundary:

- `handler` menangani HTTP, request binding, response, mapping DTO, dan mapping
  service error ke HTTP status.
- `service` berisi business logic, memakai domain entity, dan memetakan
  repository error ke service error.
- `repository` interface memakai domain entity dan repository error.
- `repository/*/database` menangani SQL, DB row, mapping ke entity, dan mapping
  error driver database ke repository error.
- `.go-arch-lint.yml` menjadi guard CI untuk dependency direction dan vendor
  import per layer.

## Consequences

Kode tetap familiar untuk tim yang biasa memakai service-repository pattern,
tetapi domain tidak bocor ke detail HTTP atau database. Struktur masih ringkas
dan tidak membutuhkan framework dependency injection. Jika feature baru membuat
package layer baru, rule `.go-arch-lint.yml` perlu ikut diupdate.
