# 0005. Use Explicit Response Package

## Status

Accepted

## Context

Nama seperti `util`, `helper`, atau `apiresponse` bisa membingungkan. Tim perlu
package yang jelas untuk format response HTTP standar.

## Decision

Gunakan package:

```text
internal/response
```

Dengan type:

```go
SuccessResponse[T]
MessageResponse
ErrorResponse
```

Dan helper:

```go
response.WithData(...)
response.WithoutData(...)
response.Error(...)
```

## Consequences

Handler memiliki format response yang konsisten dan mudah dibaca. Package ini
tidak menjadi tempat utility umum; isinya hanya response HTTP.

