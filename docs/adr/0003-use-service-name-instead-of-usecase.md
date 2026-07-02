# 0003. Use Service Name Instead Of Usecase

## Status

Accepted

## Context

Tim lebih familiar dengan pola Spring Boot: controller -> service ->
repository. Istilah `usecase` valid dalam Clean Architecture, tetapi dapat
membuat onboarding lebih sulit untuk tim yang terbiasa dengan `service`.

## Decision

Gunakan package `internal/service` untuk business logic.

Service tetap mengikuti boundary clean architecture:

- Service tidak import Gin.
- Service tidak import DTO HTTP.
- Service tidak import database row.
- Service menerima dan mengembalikan domain entity.

## Consequences

Nama folder lebih familiar untuk tim, tetapi dependency rule tetap bersih.
Jika business logic tumbuh sangat kompleks, package service dapat dipecah per
feature tanpa mengganti konsep utama.

