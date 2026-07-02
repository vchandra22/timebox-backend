# 0004. Register Routes In Feature Handlers

## Status

Accepted

## Context

Jika setiap feature memiliki file `api_<feature>.go` di router, package router
akan terus menumpuk saat entity bertambah. Route detail juga menjadi jauh dari
handler yang menjalankan logic HTTP.

## Decision

Setiap handler feature memiliki method:

```go
func (h *UserHandler) RegisterRoutes(routeGroup *gin.RouterGroup)
```

Router utama hanya membuat base group dan memanggil registration:

```go
api := r.Group("/api/v1")
handlers.Health.RegisterRoutes(api)
handlers.User.RegisterRoutes(api)
```

## Consequences

Route feature dekat dengan handler feature. Package router tetap tipis. Saat
menambah feature baru, tim cukup membuat handler baru dan menambahkan satu baris
registration di router utama.

