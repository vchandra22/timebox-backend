# 0002. Keep Domain Entity Separate From DTO And Database Row

## Status

Accepted

## Context

Satu struct sering dipakai sekaligus untuk JSON response, business logic, dan
database scan. Cara itu cepat di awal, tetapi membuat layer saling bergantung.

## Decision

Pisahkan model menjadi tiga jenis:

```text
internal/entity/              # domain entity
internal/dto/<feature>/       # HTTP request/response DTO
internal/repository/<feature>/database/row.go
```

Rules:

- Entity tidak memiliki `json` atau `db` tag.
- DTO hanya dipakai handler.
- Row hanya dipakai database repository.
- Mapping DTO <-> entity dilakukan di handler.
- Mapping row <-> entity dilakukan di repository database.

## Consequences

Perubahan response API tidak memaksa perubahan domain. Perubahan schema database
tidak bocor ke service atau handler. Ada sedikit mapping code, tetapi masih
lebih mudah dirawat dibanding satu struct untuk semua layer.

