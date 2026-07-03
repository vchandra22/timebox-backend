# API Contract — Timebox Space Backend Golang

**Produk:** Timebox Space  
**Versi API:** v1  
**Base Path:** `/api/v1`  
**Backend Stack:** Go 1.26.4, Gin, Cobra, Viper, Zap, sqlx, pgx/stdlib, go-redis, PostgreSQL, Redis  
**Format Dokumen:** REST API Contract untuk implementasi backend Golang dengan response envelope mengikuti boilerplate project  
**Target Konsumen API:** Vue 3 Web App, Worker internal, WebSocket client, Telegram webhook

---

## 1. Prinsip Umum API

### 1.1 Base URL

Development:

```txt
http://localhost:8080/api/v1
```

Staging:

```txt
https://staging-api.timeboxspace.app/api/v1
```

Production:

```txt
https://api.timeboxspace.app/api/v1
```

WebSocket:

```txt
ws://localhost:8080/ws?token=<access_token>
wss://api.timeboxspace.app/ws?token=<access_token>
```

### 1.2 Format Data

Semua request dan response menggunakan JSON.

```http
Content-Type: application/json
Accept: application/json
```

### 1.3 Timezone dan Timestamp

Semua timestamp yang dikirim dan disimpan backend menggunakan format **RFC3339** dengan timezone eksplisit.

Contoh:

```json
"2026-07-03T09:30:00+07:00"
```

Backend menyimpan timestamp ke PostgreSQL sebagai `timestamptz`. Konversi tampilan mengikuti timezone user/workspace.

### 1.4 UUID

Semua entity utama menggunakan UUID string.

Contoh:

```json
"id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1"
```

### 1.5 Header Standar

Endpoint yang membutuhkan autentikasi wajib mengirim header berikut:

```http
Authorization: Bearer <access_token>
X-Request-Id: <optional-client-generated-request-id>
```

Jika `X-Request-Id` tidak dikirim oleh client, backend membuat request id otomatis dan mengembalikannya di response.

### 1.6 Response Envelope Standar

Response API wajib mengikuti helper response default project pada package `internal/response`. Field utama yang digunakan adalah `status`, `message`, `data`, `meta`, dan `error`.

Response sukses non-list:

```json
{
  "status": true,
  "message": "workspace fetched",
  "data": {},
  "meta": null
}
```

Response sukses list dengan pagination:

```json
{
  "status": true,
  "message": "tasks fetched",
  "data": [],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 35,
    "totalPages": 4
  }
}
```

Response error:

```json
{
  "status": false,
  "message": "invalid request",
  "error": "validation error"
}
```

Catatan:

- Jangan gunakan field `success` pada body response. Gunakan `status` bernilai boolean.
- Jangan kirim raw internal error ke client. Handler hanya mengirim pesan aman sesuai mapping service error.
- `request_id` boleh tetap dipakai untuk logging Zap dan header `X-Request-Id`, tetapi tidak menjadi field wajib di body response.
- `meta` hanya dikirim saat dibutuhkan, terutama response list/pagination.

### 1.7 Pagination Standar

Query parameter mengikuti project default:

```txt
?page=1&limit=10
```

Meta response:

```json
{
  "page": 1,
  "limit": 10,
  "total": 35,
  "totalPages": 4
}
```

Batas default:

| Parameter | Default | Maksimum |
|---|---:|---:|
| page | 1 | - |
| limit | 10 | 100 |

### 1.8 Response Message Naming

Message mengikuti pola sederhana dan konsisten seperti project default:

- List: `users fetched`, `tasks fetched`, `workspaces fetched`.
- Detail: `user fetched`, `task fetched`, `workspace fetched`.
- Create: `user created`, `task created`, `timebox created`.
- Update: `user updated`, `task updated`, `timebox updated`.
- Delete/Archive: `task deleted`, `goal archived`.
- Error binding/validation: `invalid request`.

### 1.9 Sorting Standar

Query parameter:

```txt
?sort=created_at&order=desc
```

Nilai `order`:

```txt
asc | desc
```

### 1.10 Soft Delete

Sebagian besar data utama menggunakan soft delete:

- workspace member
- goal
- task
- timebox
- comment
- attachment
- category/tag jika masih dipakai

Data histori tidak boleh hilang karena dipakai untuk report, audit trail, dan time log.

---

## 2. Status Code Standar

| HTTP Status | Penggunaan |
|---:|---|
| 200 | Request sukses |
| 201 | Data berhasil dibuat |
| 202 | Request diterima untuk diproses worker/asinkron |
| 204 | Sukses tanpa body |
| 400 | Request tidak valid |
| 401 | Token tidak valid / belum login |
| 403 | Tidak punya permission |
| 404 | Resource tidak ditemukan |
| 409 | Konflik data, misalnya email/slug sudah dipakai |
| 422 | Validasi bisnis gagal |
| 429 | Rate limit tercapai |
| 500 | Internal server error |
| 503 | Dependency tidak tersedia, misalnya Redis/PostgreSQL down |

---

## 3. Error Code Standar

| Code | HTTP | Keterangan |
|---|---:|---|
| `VALIDATION_ERROR` | 400/422 | Input tidak valid |
| `UNAUTHORIZED` | 401 | Token kosong/invalid/expired |
| `FORBIDDEN` | 403 | Tidak punya akses |
| `NOT_FOUND` | 404 | Resource tidak ditemukan |
| `CONFLICT` | 409 | Data bentrok |
| `RATE_LIMITED` | 429 | Terlalu banyak request |
| `TOKEN_EXPIRED` | 401 | Access/refresh token expired |
| `REFRESH_TOKEN_REVOKED` | 401 | Refresh token sudah dicabut |
| `WORKSPACE_INACTIVE` | 403 | Workspace nonaktif |
| `INVALID_TIME_RANGE` | 422 | Waktu mulai/selesai tidak valid |
| `TIMER_ALREADY_RUNNING` | 409 | User sudah punya timer aktif |
| `TIMEBOX_NOT_RUNNING` | 422 | Aksi timer tidak sesuai status |
| `OVERLAP_WARNING` | 200/422 | Jadwal overlap; warning bukan blokir |
| `TELEGRAM_LINK_TOKEN_EXPIRED` | 422 | Token linking Telegram expired |
| `CLOUDINARY_SIGNATURE_FAILED` | 500 | Gagal generate signature upload |
| `DEPENDENCY_UNAVAILABLE` | 503 | PostgreSQL/Redis/service eksternal tidak tersedia |

---

## 4. Enum Global

### 4.1 Role Workspace

```txt
super_admin
owner
admin
member
viewer
```

### 4.2 Status Workspace Member

```txt
active
invited
inactive
removed
```

### 4.3 Task Status

```txt
backlog
scheduled
in_progress
done
cancelled
```

### 4.4 Task Priority

```txt
low
medium
high
urgent
```

### 4.5 Timebox Status

```txt
planned
running
paused
completed
overrun
skipped
cancelled
```

### 4.6 Time Log Source

```txt
timer
manual
system
```

### 4.7 Notification Channel

```txt
in_app
email
telegram
```

### 4.8 Notification Trigger Type

```txt
timebox_reminder
timebox_overrun
daily_planning_reminder
daily_summary
mention
shared_timebox_invitation
streak_milestone
workspace_invitation
team_invitation
```

### 4.9 Resource Type

```txt
workspace
goal
task
timebox
comment
attachment
category
tag
team
notification
telegram_link
```

### 4.10 Recurrence Frequency

```txt
daily
weekdays
weekly
custom_interval
```

### 4.11 Report Export Format

```txt
json
pdf
xlsx
csv
```

---

## 5. Authentication Contract

### 5.1 Register

```http
POST /api/v1/auth/register
```

Auth: public

Request body:

```json
{
  "full_name": "Vincent Chandra",
  "email": "vincent@example.com",
  "password": "SecretPassword123!",
  "timezone": "Asia/Jakarta"
}
```

Validasi:

| Field | Rule |
|---|---|
| full_name | required, max 120 |
| email | required, valid email, unique |
| password | required, min 8, harus kuat |
| timezone | required, valid IANA timezone |

Response 201:

```json
{
  "status": true,
  "message": "Account registered successfully",
  "data": {
    "user": {
      "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
      "full_name": "Vincent Chandra",
      "email": "vincent@example.com",
      "timezone": "Asia/Jakarta",
      "avatar_url": null,
      "email_verified_at": null,
      "created_at": "2026-07-03T09:30:00+07:00"
    },
    "tokens": {
      "access_token": "jwt-access-token",
      "refresh_token": "jwt-refresh-token",
      "token_type": "Bearer",
      "expires_in": 900
    }
  },
  "meta": null
}
```

Efek backend:

- Password di-hash memakai bcrypt/argon2.
- Access token berlaku sesuai `JWT_ACCESS_TTL`.
- Refresh token disimpan di Redis dengan TTL sesuai `JWT_REFRESH_TTL`.
- User mendapatkan workspace pribadi default jika business rule MVP mengaktifkannya.

---

### 5.2 Login

```http
POST /api/v1/auth/login
```

Auth: public

Request body:

```json
{
  "email": "vincent@example.com",
  "password": "SecretPassword123!"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
      "full_name": "Vincent Chandra",
      "email": "vincent@example.com",
      "timezone": "Asia/Jakarta",
      "avatar_url": "https://res.cloudinary.com/demo/image/upload/avatar.png"
    },
    "tokens": {
      "access_token": "jwt-access-token",
      "refresh_token": "jwt-refresh-token",
      "token_type": "Bearer",
      "expires_in": 900
    }
  },
  "meta": null
}
```

Catatan implementasi:

- Endpoint login wajib rate limited via Redis.
- Login gagal tidak boleh membocorkan apakah email terdaftar atau password salah.
- Login history dicatat.

---

### 5.3 Refresh Token

```http
POST /api/v1/auth/refresh
```

Auth: public dengan refresh token

Request body:

```json
{
  "refresh_token": "jwt-refresh-token"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Token refreshed",
  "data": {
    "access_token": "new-jwt-access-token",
    "refresh_token": "new-jwt-refresh-token",
    "token_type": "Bearer",
    "expires_in": 900
  },
  "meta": null
}
```

Efek backend:

- Refresh token lama direvoke.
- Refresh token baru disimpan ke Redis.
- Rotasi token wajib untuk mengurangi risiko replay attack.

---

### 5.4 Logout

```http
POST /api/v1/auth/logout
```

Auth: required

Request body:

```json
{
  "refresh_token": "jwt-refresh-token"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Logout successful",
  "data": null,
  "meta": null
}
```

Efek backend:

- Refresh token dihapus/direvoke dari Redis.
- Session aktif terkait device dicabut.

---

### 5.5 Forgot Password

```http
POST /api/v1/auth/forgot-password
```

Auth: public

Request body:

```json
{
  "email": "vincent@example.com"
}
```

Response 200:

```json
{
  "status": true,
  "message": "If the email exists, reset instructions will be sent",
  "data": null,
  "meta": null
}
```

Catatan:

- Response harus sama baik email ditemukan atau tidak.
- Token reset password disimpan di Redis dengan TTL singkat.

---

### 5.6 Reset Password

```http
POST /api/v1/auth/reset-password
```

Auth: public

Request body:

```json
{
  "token": "reset-token",
  "new_password": "NewSecretPassword123!"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Password reset successfully",
  "data": null,
  "meta": null
}
```

Efek backend:

- Password hash diperbarui.
- Semua refresh token user dicabut dari Redis.
- User perlu login ulang di semua device.

---

## 6. Profile & Session Contract

### 6.1 Get Current User

```http
GET /api/v1/me
```

Auth: required

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
    "full_name": "Vincent Chandra",
    "email": "vincent@example.com",
    "timezone": "Asia/Jakarta",
    "avatar_url": null,
    "work_hours": {
      "start": "08:00",
      "end": "17:00"
    },
    "created_at": "2026-07-03T09:30:00+07:00"
  },
  "meta": null
}
```

---

### 6.2 Update Profile

```http
PATCH /api/v1/me
```

Auth: required

Request body:

```json
{
  "full_name": "Vincent C.",
  "timezone": "Asia/Jakarta",
  "avatar_url": "https://res.cloudinary.com/demo/image/upload/avatar.png",
  "work_hours": {
    "start": "08:00",
    "end": "17:00"
  }
}
```

Response 200:

```json
{
  "status": true,
  "message": "Profile updated",
  "data": {
    "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
    "full_name": "Vincent C.",
    "email": "vincent@example.com",
    "timezone": "Asia/Jakarta",
    "avatar_url": "https://res.cloudinary.com/demo/image/upload/avatar.png",
    "work_hours": {
      "start": "08:00",
      "end": "17:00"
    },
    "updated_at": "2026-07-03T10:00:00+07:00"
  },
  "meta": null
}
```

---

### 6.3 Change Password

```http
PATCH /api/v1/me/password
```

Auth: required

Request body:

```json
{
  "current_password": "OldPassword123!",
  "new_password": "NewPassword123!"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Password changed successfully",
  "data": null,
  "meta": null
}
```

Efek backend:

- Semua refresh token lain dicabut.
- Device saat ini dapat tetap login atau diminta login ulang sesuai policy.

---

### 6.4 List Active Sessions

```http
GET /api/v1/me/sessions
```

Auth: required

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "sess_01JZ4VP4BT9YTMDC9KSW7TZCG8",
      "device_name": "Firefox on Ubuntu",
      "ip_address": "127.0.0.1",
      "last_active_at": "2026-07-03T10:00:00+07:00",
      "is_current": true
    }
  ],
  "meta": null
}
```

---

### 6.5 Revoke Session

```http
DELETE /api/v1/me/sessions/:sessionId
```

Auth: required

Response 200:

```json
{
  "status": true,
  "message": "Session revoked",
  "data": null,
  "meta": null
}
```

---

## 7. Workspace Contract

### 7.1 List Workspaces

```http
GET /api/v1/workspaces
```

Auth: required

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| q | string | no | Search nama/slug |
| status | string | no | active/inactive |
| page | int | no | Pagination |
| limit | int | no | Pagination |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
      "name": "Timebox Personal",
      "slug": "timebox-personal",
      "logo_url": null,
      "timezone": "Asia/Jakarta",
      "role": "owner",
      "member_count": 1,
      "status": "active",
      "created_at": "2026-07-03T09:30:00+07:00"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "totalPages": 1
  }
}
```

---

### 7.2 Create Workspace

```http
POST /api/v1/workspaces
```

Auth: required

Request body:

```json
{
  "name": "Thursina IT Team",
  "slug": "thursina-it-team",
  "timezone": "Asia/Jakarta",
  "logo_url": null
}
```

Response 201:

```json
{
  "status": true,
  "message": "Workspace created",
  "data": {
    "id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "name": "Thursina IT Team",
    "slug": "thursina-it-team",
    "timezone": "Asia/Jakarta",
    "logo_url": null,
    "owner_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
    "status": "active",
    "created_at": "2026-07-03T10:10:00+07:00"
  },
  "meta": null
}
```

Business rule:

- Pembuat workspace otomatis menjadi `owner`.
- Slug harus unik global.

---

### 7.3 Get Workspace Detail

```http
GET /api/v1/workspaces/:id
```

Auth: required

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "name": "Thursina IT Team",
    "slug": "thursina-it-team",
    "timezone": "Asia/Jakarta",
    "logo_url": null,
    "owner": {
      "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
      "full_name": "Vincent Chandra"
    },
    "settings": {
      "leaderboard_enabled": false,
      "default_planner_interval_minutes": 30
    },
    "member_count": 5,
    "created_at": "2026-07-03T10:10:00+07:00"
  },
  "meta": null
}
```

---

### 7.4 Update Workspace

```http
PATCH /api/v1/workspaces/:id
```

Auth: owner/admin

Request body:

```json
{
  "name": "Thursina IT Department",
  "slug": "thursina-it-department",
  "timezone": "Asia/Jakarta",
  "logo_url": "https://res.cloudinary.com/demo/image/upload/logo.png",
  "settings": {
    "leaderboard_enabled": true,
    "default_planner_interval_minutes": 30
  }
}
```

Response 200:

```json
{
  "status": true,
  "message": "Workspace updated",
  "data": {
    "id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "name": "Thursina IT Department",
    "slug": "thursina-it-department",
    "timezone": "Asia/Jakarta",
    "logo_url": "https://res.cloudinary.com/demo/image/upload/logo.png",
    "settings": {
      "leaderboard_enabled": true,
      "default_planner_interval_minutes": 30
    },
    "updated_at": "2026-07-03T10:20:00+07:00"
  },
  "meta": null
}
```

---

### 7.5 Invite Workspace Member

```http
POST /api/v1/workspaces/:id/invite
```

Auth: owner/admin

Request body:

```json
{
  "email": "member@example.com",
  "role": "member",
  "team_ids": [
    "89f8e070-953f-4507-a777-11fb7f7c39af"
  ]
}
```

Response 201:

```json
{
  "status": true,
  "message": "Invitation created",
  "data": {
    "invite_id": "inv_01JZ4VZEX15R5X3GJQY4AXZXXP",
    "email": "member@example.com",
    "role": "member",
    "status": "invited",
    "expires_at": "2026-07-10T10:20:00+07:00"
  },
  "meta": null
}
```

---

### 7.6 List Workspace Members

```http
GET /api/v1/workspaces/:id/members
```

Auth: workspace member

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| role | string | no | owner/admin/member/viewer |
| status | string | no | active/invited/inactive/removed |
| team_id | uuid | no | Filter team |
| q | string | no | Search nama/email |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "user_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
      "full_name": "Vincent Chandra",
      "email": "vincent@example.com",
      "avatar_url": null,
      "role": "owner",
      "status": "active",
      "teams": [
        {
          "id": "89f8e070-953f-4507-a777-11fb7f7c39af",
          "name": "Backend Team"
        }
      ],
      "joined_at": "2026-07-03T10:10:00+07:00"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "totalPages": 1
  }
}
```

---

### 7.7 Update Workspace Member

```http
PATCH /api/v1/workspaces/:id/members/:userId
```

Auth: owner/admin

Request body:

```json
{
  "role": "admin",
  "status": "active",
  "team_ids": [
    "89f8e070-953f-4507-a777-11fb7f7c39af"
  ]
}
```

Response 200:

```json
{
  "status": true,
  "message": "Member updated",
  "data": {
    "user_id": "b432ab9f-1330-454b-bf03-0b9186a51b27",
    "role": "admin",
    "status": "active",
    "team_ids": [
      "89f8e070-953f-4507-a777-11fb7f7c39af"
    ],
    "updated_at": "2026-07-03T10:25:00+07:00"
  },
  "meta": null
}
```

---

## 8. Team Contract

### 8.1 List Teams

```http
GET /api/v1/workspaces/:wsId/teams
```

Auth: workspace member

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "89f8e070-953f-4507-a777-11fb7f7c39af",
      "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
      "name": "Backend Team",
      "description": "Golang backend team",
      "member_count": 3,
      "created_at": "2026-07-03T10:10:00+07:00"
    }
  ],
  "meta": null
}
```

---

### 8.2 Create Team

```http
POST /api/v1/workspaces/:wsId/teams
```

Auth: owner/admin

Request body:

```json
{
  "name": "Backend Team",
  "description": "Golang backend team"
}
```

Response 201:

```json
{
  "status": true,
  "message": "Team created",
  "data": {
    "id": "89f8e070-953f-4507-a777-11fb7f7c39af",
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "name": "Backend Team",
    "description": "Golang backend team",
    "created_at": "2026-07-03T10:30:00+07:00"
  },
  "meta": null
}
```

---

### 8.3 Update Team

```http
PATCH /api/v1/teams/:id
```

Auth: owner/admin

Request body:

```json
{
  "name": "Core Backend Team",
  "description": "Core API and worker team"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Team updated",
  "data": {
    "id": "89f8e070-953f-4507-a777-11fb7f7c39af",
    "name": "Core Backend Team",
    "description": "Core API and worker team",
    "updated_at": "2026-07-03T10:35:00+07:00"
  },
  "meta": null
}
```

---

### 8.4 Delete Team

```http
DELETE /api/v1/teams/:id
```

Auth: owner/admin

Response 200:

```json
{
  "status": true,
  "message": "Team deleted",
  "data": null,
  "meta": null
}
```

Catatan:

- Menghapus team tidak menghapus user.
- Relasi `team_members` di-soft-delete.

---

## 9. Goal Contract

### 9.1 List Goals

```http
GET /api/v1/workspaces/:wsId/goals
```

Auth: workspace member

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| q | string | no | Search title/description |
| status | string | no | active/archived/completed |
| pinned | bool | no | Goal yang dipin |
| page | int | no | Pagination |
| limit | int | no | Pagination |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
      "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
      "title": "Rilis MVP Timebox Space",
      "description": "Goal untuk menyelesaikan MVP backend dan frontend.",
      "target_date": "2026-08-30",
      "status": "active",
      "is_pinned": true,
      "progress_percent": 35.5,
      "created_at": "2026-07-03T10:40:00+07:00"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "totalPages": 1
  }
}
```

---

### 9.2 Create Goal

```http
POST /api/v1/workspaces/:wsId/goals
```

Auth: owner/admin/member

Request body:

```json
{
  "title": "Rilis MVP Timebox Space",
  "description": "Goal untuk menyelesaikan MVP backend dan frontend.",
  "target_date": "2026-08-30",
  "is_pinned": true
}
```

Response 201:

```json
{
  "status": true,
  "message": "Goal created",
  "data": {
    "id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "title": "Rilis MVP Timebox Space",
    "description": "Goal untuk menyelesaikan MVP backend dan frontend.",
    "target_date": "2026-08-30",
    "status": "active",
    "is_pinned": true,
    "created_at": "2026-07-03T10:40:00+07:00"
  },
  "meta": null
}
```

---

### 9.3 Get Goal Detail

```http
GET /api/v1/goals/:id
```

Auth: workspace member

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "title": "Rilis MVP Timebox Space",
    "description": "Goal untuk menyelesaikan MVP backend dan frontend.",
    "target_date": "2026-08-30",
    "status": "active",
    "is_pinned": true,
    "progress": {
      "planned_minutes": 2400,
      "actual_minutes": 860,
      "completed_timeboxes": 12,
      "progress_percent": 35.5
    },
    "created_by": {
      "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
      "full_name": "Vincent Chandra"
    },
    "created_at": "2026-07-03T10:40:00+07:00",
    "updated_at": "2026-07-03T10:40:00+07:00"
  },
  "meta": null
}
```

---

### 9.4 Update Goal

```http
PATCH /api/v1/goals/:id
```

Auth: owner/admin/member pembuat goal

Request body:

```json
{
  "title": "Rilis MVP Timebox Space v1",
  "description": "Menyelesaikan API, UI, dan worker MVP.",
  "target_date": "2026-09-15",
  "status": "active",
  "is_pinned": false
}
```

Response 200:

```json
{
  "status": true,
  "message": "Goal updated",
  "data": {
    "id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
    "title": "Rilis MVP Timebox Space v1",
    "description": "Menyelesaikan API, UI, dan worker MVP.",
    "target_date": "2026-09-15",
    "status": "active",
    "is_pinned": false,
    "updated_at": "2026-07-03T10:45:00+07:00"
  },
  "meta": null
}
```

---

### 9.5 Delete/Archive Goal

```http
DELETE /api/v1/goals/:id
```

Auth: owner/admin/member pembuat goal

Response 200:

```json
{
  "status": true,
  "message": "Goal archived",
  "data": null,
  "meta": null
}
```

Catatan:

- Delete pada MVP direkomendasikan menjadi archive/soft delete.
- Task/timebox historis tetap menyimpan relasi goal.

---

## 10. Task & Backlog Contract

### 10.1 List Tasks

```http
GET /api/v1/workspaces/:wsId/tasks
```

Auth: workspace member

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| q | string | no | Search title/description |
| status | string | no | backlog/scheduled/in_progress/done/cancelled |
| priority | string | no | low/medium/high/urgent |
| goal_id | uuid | no | Filter goal |
| assignee_id | uuid | no | Filter assignee |
| category_id | uuid | no | Filter category |
| tag_ids | string | no | UUID dipisah koma |
| include_done | bool | no | Default false |
| sort | string | no | position/created_at/due_date/priority |
| order | string | no | asc/desc |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
      "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
      "goal_id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
      "title": "Buat module auth Golang",
      "description": "Register, login, refresh token, logout.",
      "status": "backlog",
      "priority": "high",
      "estimated_minutes": 180,
      "position": 1000,
      "assignee": {
        "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
        "full_name": "Vincent Chandra",
        "avatar_url": null
      },
      "goal": {
        "id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
        "title": "Rilis MVP Timebox Space"
      },
      "tags": [
        {
          "id": "9a801852-8742-4d04-ad15-974b4f252db6",
          "name": "backend"
        }
      ],
      "created_at": "2026-07-03T10:50:00+07:00",
      "updated_at": "2026-07-03T10:50:00+07:00"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "totalPages": 1
  }
}
```

---

### 10.2 Create Task

```http
POST /api/v1/workspaces/:wsId/tasks
```

Auth: owner/admin/member

Request body:

```json
{
  "goal_id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
  "assignee_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
  "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
  "title": "Buat module auth Golang",
  "description": "Register, login, refresh token, logout.",
  "priority": "high",
  "estimated_minutes": 180,
  "tag_ids": [
    "9a801852-8742-4d04-ad15-974b4f252db6"
  ],
  "checklist": [
    {
      "title": "Buat endpoint register"
    },
    {
      "title": "Buat endpoint login"
    }
  ]
}
```

Validasi:

| Field | Rule |
|---|---|
| title | required, max 200 |
| estimated_minutes | optional, min 1 |
| priority | optional enum |
| goal_id | nullable, harus dalam workspace yang sama |
| assignee_id | nullable, harus member workspace |

Response 201:

```json
{
  "status": true,
  "message": "Task created",
  "data": {
    "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "goal_id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
    "assignee_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
    "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
    "title": "Buat module auth Golang",
    "description": "Register, login, refresh token, logout.",
    "status": "backlog",
    "priority": "high",
    "estimated_minutes": 180,
    "position": 1000,
    "created_at": "2026-07-03T10:50:00+07:00"
  },
  "meta": null
}
```

---

### 10.3 Get Task Detail

```http
GET /api/v1/tasks/:id
```

Auth: workspace member

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "goal_id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
    "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
    "title": "Buat module auth Golang",
    "description": "Register, login, refresh token, logout.",
    "status": "backlog",
    "priority": "high",
    "estimated_minutes": 180,
    "position": 1000,
    "assignee": {
      "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
      "full_name": "Vincent Chandra",
      "avatar_url": null
    },
    "checklist": [
      {
        "id": "fa0f7ed8-30ef-4311-8f54-381c9c928af6",
        "title": "Buat endpoint register",
        "is_done": false,
        "position": 1000
      }
    ],
    "tags": [
      {
        "id": "9a801852-8742-4d04-ad15-974b4f252db6",
        "name": "backend"
      }
    ],
    "timeboxes_count": 0,
    "created_at": "2026-07-03T10:50:00+07:00",
    "updated_at": "2026-07-03T10:50:00+07:00"
  },
  "meta": null
}
```

---

### 10.4 Update Task

```http
PATCH /api/v1/tasks/:id
```

Auth: owner/admin/member owner/assignee

Request body:

```json
{
  "goal_id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
  "assignee_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
  "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
  "title": "Buat module authentication Golang",
  "description": "Auth lengkap JWT access + refresh token.",
  "status": "in_progress",
  "priority": "urgent",
  "estimated_minutes": 240,
  "tag_ids": [
    "9a801852-8742-4d04-ad15-974b4f252db6"
  ]
}
```

Response 200:

```json
{
  "status": true,
  "message": "Task updated",
  "data": {
    "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
    "title": "Buat module authentication Golang",
    "status": "in_progress",
    "priority": "urgent",
    "estimated_minutes": 240,
    "updated_at": "2026-07-03T11:00:00+07:00"
  },
  "meta": null
}
```

---

### 10.5 Delete Task

```http
DELETE /api/v1/tasks/:id
```

Auth: owner/admin/member owner/assignee

Response 200:

```json
{
  "status": true,
  "message": "Task deleted",
  "data": null,
  "meta": null
}
```

Catatan:

- Delete task adalah soft delete.
- Timebox historis tetap dapat menampilkan judul task yang pernah terkait.

---

### 10.6 Move Task / Kanban Drag and Drop

```http
PATCH /api/v1/tasks/:id/move
```

Auth: owner/admin/member owner/assignee

Request body:

```json
{
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "to_status": "in_progress",
  "position": 2000,
  "before_task_id": null,
  "after_task_id": "f3f4d021-c3ce-42bd-bb13-6bc41cb1549e"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Task moved",
  "data": {
    "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
    "from_status": "backlog",
    "to_status": "in_progress",
    "position": 2000,
    "updated_at": "2026-07-03T11:05:00+07:00"
  },
  "meta": null
}
```

Efek backend:

- Update `tasks.status` dan `tasks.position` dalam satu transaksi.
- Tulis activity log.
- Publish event Redis Pub/Sub untuk WebSocket:

```json
{
  "event": "task.updated",
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "payload": {
    "task_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
    "status": "in_progress",
    "position": 2000
  }
}
```

---

### 10.7 Convert Task to Timebox

```http
POST /api/v1/tasks/:id/convert-to-timebox
```

Auth: owner/admin/member owner/assignee

Request body:

```json
{
  "owner_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
  "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
  "scheduled_start": "2026-07-04T09:00:00+07:00",
  "scheduled_end": "2026-07-04T11:00:00+07:00",
  "split": false
}
```

Response 201:

```json
{
  "status": true,
  "message": "Task converted to timebox",
  "data": {
    "task": {
      "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
      "status": "scheduled"
    },
    "timeboxes": [
      {
        "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
        "task_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
        "title": "Buat module authentication Golang",
        "scheduled_start": "2026-07-04T09:00:00+07:00",
        "scheduled_end": "2026-07-04T11:00:00+07:00",
        "status": "planned"
      }
    ],
    "warnings": []
  },
  "meta": null
}
```

Jika overlap:

```json
"warnings": [
  {
    "code": "SCHEDULE_OVERLAP",
    "message": "Timebox overlaps with another planned timebox",
    "conflict_timebox_id": "44d571f2-b4d2-49f0-82c4-01f4a10aa322"
  }
]
```

---

## 11. Kanban Board Contract

### 11.1 Get Kanban Board

```http
GET /api/v1/workspaces/:wsId/kanban
```

Auth: workspace member

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| goal_id | uuid | no | Filter goal |
| assignee_id | uuid | no | Filter assignee |
| category_id | uuid | no | Filter category |
| tag_ids | string | no | UUID dipisah koma |
| q | string | no | Search task |
| mode | string | no | my/team, default my |
| swimlane | string | no | none/goal/assignee |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "columns": [
      {
        "status": "backlog",
        "title": "Backlog",
        "is_visible": true,
        "wip_limit": null,
        "total": 1,
        "tasks": [
          {
            "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
            "title": "Buat module auth Golang",
            "priority": "high",
            "estimated_minutes": 180,
            "position": 1000,
            "assignee": {
              "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
              "full_name": "Vincent Chandra",
              "avatar_url": null
            },
            "goal": {
              "id": "6eb12bbd-d8e1-4593-8cc1-3a7a8235673d",
              "title": "Rilis MVP Timebox Space"
            },
            "category": {
              "id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
              "name": "Deep Work",
              "color": "#2563EB"
            },
            "tags": [
              {
                "id": "9a801852-8742-4d04-ad15-974b4f252db6",
                "name": "backend"
              }
            ]
          }
        ]
      },
      {
        "status": "in_progress",
        "title": "In Progress",
        "is_visible": true,
        "wip_limit": 3,
        "total": 0,
        "tasks": []
      },
      {
        "status": "done",
        "title": "Done",
        "is_visible": true,
        "wip_limit": null,
        "total": 0,
        "tasks": []
      }
    ],
    "settings": {
      "visible_columns": [
        "backlog",
        "scheduled",
        "in_progress",
        "done",
        "cancelled"
      ],
      "wip_limits": {
        "in_progress": 3
      }
    }
  },
  "meta": null
}
```

---

### 11.2 Update Kanban Settings

```http
PATCH /api/v1/workspaces/:wsId/kanban/settings
```

Auth: workspace member, berlaku sebagai user preference pribadi; owner/admin dapat mengatur default workspace.

Request body:

```json
{
  "scope": "user",
  "visible_columns": [
    "backlog",
    "scheduled",
    "in_progress",
    "done"
  ],
  "wip_limits": {
    "in_progress": 3,
    "scheduled": 10
  }
}
```

Response 200:

```json
{
  "status": true,
  "message": "Kanban settings updated",
  "data": {
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "scope": "user",
    "visible_columns": [
      "backlog",
      "scheduled",
      "in_progress",
      "done"
    ],
    "wip_limits": {
      "in_progress": 3,
      "scheduled": 10
    }
  },
  "meta": null
}
```

---

## 12. Timebox Contract

### 12.1 List Timeboxes

```http
GET /api/v1/timeboxes
```

Auth: required

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| workspace_id | uuid | yes | Workspace konteks |
| date | date | no | Untuk view day |
| start_date | date | no | Untuk range/week |
| end_date | date | no | Untuk range/week |
| view | string | no | day/week/range |
| owner_id | uuid | no | Filter pemilik |
| team_id | uuid | no | Filter team |
| status | string | no | planned/running/paused/completed/overrun/skipped/cancelled |
| category_id | uuid | no | Filter kategori |
| goal_id | uuid | no | Filter goal dari task/goal relasi |
| include_logs | bool | no | Include time log ringkas |

Contoh:

```http
GET /api/v1/timeboxes?workspace_id=a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4&date=2026-07-04&view=day
```

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
      "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
      "task_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
      "owner_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
      "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
      "title": "Buat module authentication Golang",
      "description": "Auth lengkap JWT access + refresh token.",
      "scheduled_start": "2026-07-04T09:00:00+07:00",
      "scheduled_end": "2026-07-04T11:00:00+07:00",
      "planned_minutes": 120,
      "actual_minutes": 0,
      "status": "planned",
      "is_buffer": false,
      "owner": {
        "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
        "full_name": "Vincent Chandra",
        "avatar_url": null
      },
      "category": {
        "id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
        "name": "Deep Work",
        "color": "#2563EB"
      },
      "task": {
        "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
        "title": "Buat module authentication Golang"
      },
      "warnings": []
    }
  ],
  "meta": null
}
```

---

### 12.2 Create Timebox

```http
POST /api/v1/timeboxes
```

Auth: owner/admin/member

Request body:

```json
{
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "task_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
  "owner_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
  "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
  "title": "Buat module authentication Golang",
  "description": "Auth lengkap JWT access + refresh token.",
  "scheduled_start": "2026-07-04T09:00:00+07:00",
  "scheduled_end": "2026-07-04T11:00:00+07:00",
  "is_buffer": false,
  "participant_ids": []
}
```

Validasi:

| Field | Rule |
|---|---|
| workspace_id | required |
| owner_id | required, member workspace |
| category_id | required, category workspace |
| title | required jika task_id null |
| scheduled_start | required |
| scheduled_end | required, harus lebih besar dari start |
| participant_ids | optional, semua harus member workspace |

Response 201:

```json
{
  "status": true,
  "message": "Timebox created",
  "data": {
    "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "task_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
    "owner_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
    "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
    "title": "Buat module authentication Golang",
    "scheduled_start": "2026-07-04T09:00:00+07:00",
    "scheduled_end": "2026-07-04T11:00:00+07:00",
    "planned_minutes": 120,
    "status": "planned",
    "is_buffer": false,
    "warnings": [],
    "created_at": "2026-07-03T11:10:00+07:00"
  },
  "meta": null
}
```

Jika jadwal overlap, response tetap 201 jika policy MVP hanya warning:

```json
"warnings": [
  {
    "code": "SCHEDULE_OVERLAP",
    "message": "This timebox overlaps with another planned timebox",
    "conflict_timebox_id": "44d571f2-b4d2-49f0-82c4-01f4a10aa322"
  }
]
```

---

### 12.3 Get Timebox Detail

```http
GET /api/v1/timeboxes/:id
```

Auth: workspace member sesuai permission

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "task_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
    "owner_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
    "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
    "title": "Buat module authentication Golang",
    "description": "Auth lengkap JWT access + refresh token.",
    "scheduled_start": "2026-07-04T09:00:00+07:00",
    "scheduled_end": "2026-07-04T11:00:00+07:00",
    "planned_minutes": 120,
    "actual_minutes": 0,
    "status": "planned",
    "is_buffer": false,
    "owner": {
      "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
      "full_name": "Vincent Chandra",
      "avatar_url": null
    },
    "participants": [],
    "logs": [],
    "attachments_count": 0,
    "comments_count": 0,
    "created_at": "2026-07-03T11:10:00+07:00",
    "updated_at": "2026-07-03T11:10:00+07:00"
  },
  "meta": null
}
```

---

### 12.4 Update Timebox

```http
PATCH /api/v1/timeboxes/:id
```

Auth: owner/admin/member owner atau permission edit timebox orang lain

Request body:

```json
{
  "task_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
  "owner_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
  "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
  "title": "Buat module auth dan refresh token",
  "description": "Auth lengkap JWT access + refresh token.",
  "scheduled_start": "2026-07-04T09:30:00+07:00",
  "scheduled_end": "2026-07-04T11:30:00+07:00",
  "status": "planned",
  "is_buffer": false,
  "participant_ids": []
}
```

Response 200:

```json
{
  "status": true,
  "message": "Timebox updated",
  "data": {
    "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "title": "Buat module auth dan refresh token",
    "scheduled_start": "2026-07-04T09:30:00+07:00",
    "scheduled_end": "2026-07-04T11:30:00+07:00",
    "planned_minutes": 120,
    "status": "planned",
    "warnings": [],
    "updated_at": "2026-07-03T11:20:00+07:00"
  },
  "meta": null
}
```

Efek backend:

- Publish `timebox.updated` via Redis Pub/Sub.
- Activity log menyimpan old value dan new value.

---

### 12.5 Delete Timebox

```http
DELETE /api/v1/timeboxes/:id
```

Auth: owner/admin/member owner atau permission edit timebox orang lain

Response 200:

```json
{
  "status": true,
  "message": "Timebox deleted",
  "data": null,
  "meta": null
}
```

Business rule:

- Tidak boleh delete hard timebox yang sudah memiliki time log, kecuali super admin maintenance.
- Untuk timebox dengan log, lakukan soft delete/cancel.

---

### 12.6 Duplicate Timebox

```http
POST /api/v1/timeboxes/:id/duplicate
```

Auth: owner/admin/member owner

Request body:

```json
{
  "target_date": "2026-07-05",
  "scheduled_start_time": "09:00",
  "keep_duration": true
}
```

Response 201:

```json
{
  "status": true,
  "message": "Timebox duplicated",
  "data": {
    "id": "7155f0d1-5fc3-46d2-9e60-2210ac493545",
    "source_timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "scheduled_start": "2026-07-05T09:00:00+07:00",
    "scheduled_end": "2026-07-05T11:00:00+07:00",
    "status": "planned",
    "warnings": []
  },
  "meta": null
}
```

---

### 12.7 Move Missed Timebox to Next Day

```http
POST /api/v1/timeboxes/:id/move-to-next-day
```

Auth: owner/admin/member owner

Request body:

```json
{
  "keep_same_time": true
}
```

Response 200:

```json
{
  "status": true,
  "message": "Timebox moved to next day",
  "data": {
    "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "scheduled_start": "2026-07-05T09:30:00+07:00",
    "scheduled_end": "2026-07-05T11:30:00+07:00",
    "status": "planned",
    "warnings": []
  },
  "meta": null
}
```

---

## 13. Focus Timer Contract

### 13.1 Get Active Timer

```http
GET /api/v1/timer/active
```

Auth: required

Response jika ada timer aktif:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "status": "running",
    "started_at": "2026-07-04T09:30:00+07:00",
    "paused_at": null,
    "planned_minutes": 120,
    "elapsed_seconds": 900,
    "remaining_seconds": 6300,
    "server_time": "2026-07-04T09:45:00+07:00",
    "timebox": {
      "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
      "title": "Buat module auth dan refresh token",
      "scheduled_start": "2026-07-04T09:30:00+07:00",
      "scheduled_end": "2026-07-04T11:30:00+07:00"
    }
  },
  "meta": null
}
```

Response jika tidak ada timer aktif:

```json
{
  "status": true,
  "message": "No active timer",
  "data": null,
  "meta": null
}
```

Redis key:

```txt
timebox-space:<env>:timer:<user_id>
```

---

### 13.2 Start Timebox Timer

```http
POST /api/v1/timeboxes/:id/start
```

Auth: owner/admin/member owner

Request body:

```json
{
  "start_mode": "now"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Timer started",
  "data": {
    "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "status": "running",
    "started_at": "2026-07-04T09:30:00+07:00",
    "server_time": "2026-07-04T09:30:00+07:00",
    "planned_minutes": 120,
    "elapsed_seconds": 0,
    "remaining_seconds": 7200
  },
  "meta": null
}
```

Business rule:

- Hanya satu timer aktif per user.
- Jika ada timer aktif lain, return 409 `TIMER_ALREADY_RUNNING`.
- Timebox status berubah menjadi `running`.
- Buat time log segment awal dengan `started_at`, `ended_at = null`.
- Simpan state timer di Redis.
- Publish event `timer.sync` dan `timebox.updated`.

---

### 13.3 Pause Timer

```http
POST /api/v1/timeboxes/:id/pause
```

Auth: owner/admin/member owner

Request body:

```json
{
  "reason": "Istirahat sebentar"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Timer paused",
  "data": {
    "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "status": "paused",
    "paused_at": "2026-07-04T10:00:00+07:00",
    "elapsed_seconds": 1800,
    "remaining_seconds": 5400
  },
  "meta": null
}
```

Efek backend:

- Tutup time log segment berjalan dengan `ended_at = now`.
- Update Redis timer state ke `paused`.

---

### 13.4 Resume Timer

```http
POST /api/v1/timeboxes/:id/resume
```

Auth: owner/admin/member owner

Request body:

```json
{}
```

Response 200:

```json
{
  "status": true,
  "message": "Timer resumed",
  "data": {
    "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "status": "running",
    "resumed_at": "2026-07-04T10:10:00+07:00",
    "elapsed_seconds": 1800,
    "remaining_seconds": 5400
  },
  "meta": null
}
```

Efek backend:

- Buat time log segment baru dengan `started_at = now`.
- Update Redis timer state ke `running`.

---

### 13.5 Complete Timer

```http
POST /api/v1/timeboxes/:id/complete
```

Auth: owner/admin/member owner

Request body:

```json
{
  "note": "Auth basic sudah selesai. Refresh token perlu test tambahan."
}
```

Response 200:

```json
{
  "status": true,
  "message": "Timebox completed",
  "data": {
    "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "status": "completed",
    "completed_at": "2026-07-04T11:20:00+07:00",
    "planned_minutes": 120,
    "actual_minutes": 100,
    "variance_minutes": -20,
    "streak_updated": true
  },
  "meta": null
}
```

Efek backend:

- Tutup time log segment aktif jika ada.
- Hitung `actual_minutes` dari seluruh time log segment.
- Hapus Redis active timer.
- Update streak harian jika memenuhi syarat.
- Publish `timer.sync`, `timebox.updated`, dan `streak.updated` jika relevan.

---

### 13.6 Skip Timebox

```http
POST /api/v1/timeboxes/:id/skip
```

Auth: owner/admin/member owner

Request body:

```json
{
  "reason": "Prioritas berubah"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Timebox skipped",
  "data": {
    "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "status": "skipped",
    "skipped_at": "2026-07-04T11:25:00+07:00"
  },
  "meta": null
}
```

---

## 14. Time Log Contract

### 14.1 List Time Logs by Timebox

```http
GET /api/v1/timeboxes/:id/logs
```

Auth: workspace member sesuai permission

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "f14142e9-a8af-4041-a50e-2c227e6a087e",
      "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
      "started_at": "2026-07-04T09:30:00+07:00",
      "ended_at": "2026-07-04T10:00:00+07:00",
      "duration_seconds": 1800,
      "source": "timer",
      "note": null,
      "created_by": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1"
    }
  ],
  "meta": null
}
```

---

### 14.2 Create Manual Time Log

```http
POST /api/v1/timeboxes/:id/logs
```

Auth: owner/admin/member owner

Request body:

```json
{
  "started_at": "2026-07-04T13:00:00+07:00",
  "ended_at": "2026-07-04T13:45:00+07:00",
  "source": "manual",
  "note": "Pekerjaan dilanjutkan tanpa timer"
}
```

Response 201:

```json
{
  "status": true,
  "message": "Time log created",
  "data": {
    "id": "5f738651-4904-401c-b356-28f0fc7da8f5",
    "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "started_at": "2026-07-04T13:00:00+07:00",
    "ended_at": "2026-07-04T13:45:00+07:00",
    "duration_seconds": 2700,
    "source": "manual",
    "note": "Pekerjaan dilanjutkan tanpa timer"
  },
  "meta": null
}
```

Business rule:

- Manual entry wajib terkait timebox.
- Perubahan log menghitung ulang actual duration timebox.
- Perubahan manual dicatat di audit trail.

---

### 14.3 Update Time Log

```http
PATCH /api/v1/timebox-logs/:id
```

Auth: owner/admin/member owner

Request body:

```json
{
  "started_at": "2026-07-04T13:05:00+07:00",
  "ended_at": "2026-07-04T13:50:00+07:00",
  "note": "Koreksi waktu manual"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Time log updated",
  "data": {
    "id": "5f738651-4904-401c-b356-28f0fc7da8f5",
    "duration_seconds": 2700,
    "updated_at": "2026-07-04T14:00:00+07:00"
  },
  "meta": null
}
```

---

### 14.4 Delete Time Log

```http
DELETE /api/v1/timebox-logs/:id
```

Auth: owner/admin/member owner

Response 200:

```json
{
  "status": true,
  "message": "Time log deleted",
  "data": null,
  "meta": null
}
```

---

## 15. Recurring Timebox & Template Contract

### 15.1 Create Recurring Timebox Rule

```http
POST /api/v1/workspaces/:wsId/recurrence-rules
```

Auth: owner/admin/member

Request body:

```json
{
  "title": "Olahraga pagi",
  "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
  "owner_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
  "frequency": "weekly",
  "days_of_week": [
    "monday",
    "wednesday",
    "friday"
  ],
  "interval": 1,
  "start_date": "2026-07-06",
  "until_date": "2026-12-31",
  "start_time": "06:00",
  "duration_minutes": 30,
  "rolling_window_days": 14
}
```

Response 201:

```json
{
  "status": true,
  "message": "Recurrence rule created",
  "data": {
    "id": "b01d3b31-8ea6-496d-9160-f5d43d512255",
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "frequency": "weekly",
    "days_of_week": [
      "monday",
      "wednesday",
      "friday"
    ],
    "generated_until": "2026-07-20",
    "generated_count": 6
  },
  "meta": null
}
```

---

### 15.2 Update Recurring Instance

```http
PATCH /api/v1/timeboxes/:id/recurring-instance
```

Auth: owner/admin/member owner

Request body:

```json
{
  "apply_scope": "this_instance",
  "scheduled_start": "2026-07-06T06:30:00+07:00",
  "scheduled_end": "2026-07-06T07:00:00+07:00"
}
```

`apply_scope`:

```txt
this_instance | entire_series
```

Response 200:

```json
{
  "status": true,
  "message": "Recurring timebox updated",
  "data": {
    "timebox_id": "7155f0d1-5fc3-46d2-9e60-2210ac493545",
    "apply_scope": "this_instance",
    "updated_at": "2026-07-04T14:10:00+07:00"
  },
  "meta": null
}
```

---

### 15.3 Create Routine Template

```http
POST /api/v1/workspaces/:wsId/routine-templates
```

Auth: owner/admin/member

Request body:

```json
{
  "name": "Hari Fokus Backend",
  "description": "Template harian untuk development backend.",
  "items": [
    {
      "title": "Deep Work Backend",
      "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
      "start_time": "09:00",
      "duration_minutes": 120
    },
    {
      "title": "Review PR",
      "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
      "start_time": "13:00",
      "duration_minutes": 60
    }
  ]
}
```

Response 201:

```json
{
  "status": true,
  "message": "Routine template created",
  "data": {
    "id": "59a1f919-f66a-46e3-9a5e-b6553ccf8f03",
    "name": "Hari Fokus Backend",
    "items_count": 2
  },
  "meta": null
}
```

---

### 15.4 Apply Routine Template

```http
POST /api/v1/routine-templates/:id/apply
```

Auth: owner/admin/member owner

Request body:

```json
{
  "target_date": "2026-07-08",
  "owner_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1"
}
```

Response 201:

```json
{
  "status": true,
  "message": "Routine template applied",
  "data": {
    "created_timeboxes": [
      {
        "id": "9dbcf417-9ed5-4f6e-8214-aaeb0bbf91b3",
        "title": "Deep Work Backend",
        "scheduled_start": "2026-07-08T09:00:00+07:00",
        "scheduled_end": "2026-07-08T11:00:00+07:00"
      }
    ],
    "warnings": []
  },
  "meta": null
}
```

---

## 16. Category & Tag Contract

### 16.1 List Categories

```http
GET /api/v1/workspaces/:wsId/categories
```

Auth: workspace member

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
      "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
      "name": "Deep Work",
      "color": "#2563EB",
      "is_default": true,
      "created_at": "2026-07-03T10:10:00+07:00"
    }
  ],
  "meta": null
}
```

---

### 16.2 Create Category

```http
POST /api/v1/workspaces/:wsId/categories
```

Auth: owner/admin/member jika diizinkan workspace setting

Request body:

```json
{
  "name": "Meeting",
  "color": "#F97316"
}
```

Response 201:

```json
{
  "status": true,
  "message": "Category created",
  "data": {
    "id": "834f9c3d-ef20-452a-9da8-680989f3e01c",
    "name": "Meeting",
    "color": "#F97316"
  },
  "meta": null
}
```

---

### 16.3 Update Category

```http
PATCH /api/v1/categories/:id
```

Auth: owner/admin/member jika diizinkan

Request body:

```json
{
  "name": "Team Meeting",
  "color": "#EA580C"
}
```

Response 200:

```json
{
  "status": true,
  "message": "Category updated",
  "data": {
    "id": "834f9c3d-ef20-452a-9da8-680989f3e01c",
    "name": "Team Meeting",
    "color": "#EA580C"
  },
  "meta": null
}
```

---

### 16.4 Delete Category

```http
DELETE /api/v1/categories/:id
```

Auth: owner/admin

Response 200:

```json
{
  "status": true,
  "message": "Category deleted or moved to default category",
  "data": {
    "moved_timeboxes_to_category_id": "default-category-id"
  },
  "meta": null
}
```

Business rule:

- Jika kategori masih dipakai, timebox dipindah ke kategori `Lainnya`.

---

### 16.5 Tag CRUD

```http
GET    /api/v1/tags?workspace_id=:wsId&q=backend
POST   /api/v1/tags
PATCH  /api/v1/tags/:id
DELETE /api/v1/tags/:id
```

Create tag request:

```json
{
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "name": "backend"
}
```

Response 201:

```json
{
  "status": true,
  "message": "Tag created",
  "data": {
    "id": "9a801852-8742-4d04-ad15-974b4f252db6",
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "name": "backend"
  },
  "meta": null
}
```

---

## 17. Comment & Discussion Contract

### 17.1 List Comments

```http
GET /api/v1/comments?resource_type=task&resource_id=d337fdbe-264e-4f42-bcf9-ae5dfe2499c9
```

Auth: workspace member dengan akses resource

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| resource_type | string | yes | goal/task/timebox |
| resource_id | uuid | yes | ID resource |
| parent_id | uuid | no | Untuk reply |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "4a0d6acf-301b-4d97-a366-f78ca19c598d",
      "resource_type": "task",
      "resource_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
      "parent_id": null,
      "body": "Jangan lupa add test untuk refresh token.",
      "mentions": [
        {
          "user_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
          "full_name": "Vincent Chandra"
        }
      ],
      "author": {
        "id": "b432ab9f-1330-454b-bf03-0b9186a51b27",
        "full_name": "Reviewer"
      },
      "attachments": [],
      "edited_at": null,
      "created_at": "2026-07-03T11:30:00+07:00"
    }
  ],
  "meta": null
}
```

---

### 17.2 Create Comment

```http
POST /api/v1/comments
```

Auth: workspace member dengan akses resource

Request body:

```json
{
  "resource_type": "task",
  "resource_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
  "parent_id": null,
  "body": "@Vincent jangan lupa add test untuk refresh token.",
  "mention_user_ids": [
    "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1"
  ],
  "attachment_ids": []
}
```

Response 201:

```json
{
  "status": true,
  "message": "Comment created",
  "data": {
    "id": "4a0d6acf-301b-4d97-a366-f78ca19c598d",
    "resource_type": "task",
    "resource_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
    "body": "@Vincent jangan lupa add test untuk refresh token.",
    "created_at": "2026-07-03T11:30:00+07:00"
  },
  "meta": null
}
```

Efek backend:

- User yang di-mention menerima notification `mention`.
- Publish `notification.new` via WebSocket.

---

### 17.3 Update Comment

```http
PATCH /api/v1/comments/:id
```

Auth: author comment atau admin

Request body:

```json
{
  "body": "@Vincent jangan lupa add unit test dan integration test untuk refresh token."
}
```

Response 200:

```json
{
  "status": true,
  "message": "Comment updated",
  "data": {
    "id": "4a0d6acf-301b-4d97-a366-f78ca19c598d",
    "body": "@Vincent jangan lupa add unit test dan integration test untuk refresh token.",
    "edited_at": "2026-07-03T11:35:00+07:00"
  },
  "meta": null
}
```

---

### 17.4 Delete Comment

```http
DELETE /api/v1/comments/:id
```

Auth: author comment atau admin

Response 200:

```json
{
  "status": true,
  "message": "Comment deleted",
  "data": null,
  "meta": null
}
```

---

## 18. Upload & Attachment Contract

### 18.1 Generate Cloudinary Upload Signature

```http
POST /api/v1/uploads/signature
```

Auth: required

Request body:

```json
{
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "resource_type": "task",
  "resource_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
  "file_name": "design-api-contract.png",
  "file_type": "image/png",
  "file_size": 524288
}
```

Response 200:

```json
{
  "status": true,
  "message": "Upload signature generated",
  "data": {
    "cloud_name": "your-cloud-name",
    "api_key": "your-api-key",
    "timestamp": 1783049700,
    "signature": "generated-cloudinary-signature",
    "folder": "timebox-space/development/a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4/task",
    "upload_url": "https://api.cloudinary.com/v1_1/your-cloud-name/auto/upload",
    "public_id_prefix": "task/d337fdbe-264e-4f42-bcf9-ae5dfe2499c9"
  },
  "meta": null
}
```

Validasi:

- API secret Cloudinary tidak pernah dikirim ke client.
- File type harus sesuai whitelist.
- File size tidak boleh melebihi limit workspace/global.

---

### 18.2 Create Attachment Reference

```http
POST /api/v1/attachments
```

Auth: required

Request body setelah client berhasil upload ke Cloudinary:

```json
{
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "resource_type": "task",
  "resource_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
  "cloudinary_public_id": "timebox-space/development/a0376f53/task/design-api-contract",
  "url": "https://res.cloudinary.com/demo/image/upload/v1783049700/timebox-space/development/a0376f53/task/design-api-contract.png",
  "file_name": "design-api-contract.png",
  "file_type": "image/png",
  "file_size": 524288
}
```

Response 201:

```json
{
  "status": true,
  "message": "Attachment created",
  "data": {
    "id": "ec968dbe-8f0b-48f8-87b4-687cd722a3d5",
    "resource_type": "task",
    "resource_id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
    "url": "https://res.cloudinary.com/demo/image/upload/v1783049700/timebox-space/development/a0376f53/task/design-api-contract.png",
    "file_name": "design-api-contract.png",
    "file_type": "image/png",
    "file_size": 524288,
    "created_at": "2026-07-03T11:40:00+07:00"
  },
  "meta": null
}
```

---

### 18.3 List Attachments

```http
GET /api/v1/attachments?resource_type=task&resource_id=d337fdbe-264e-4f42-bcf9-ae5dfe2499c9
```

Auth: workspace member dengan akses resource

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "ec968dbe-8f0b-48f8-87b4-687cd722a3d5",
      "file_name": "design-api-contract.png",
      "file_type": "image/png",
      "file_size": 524288,
      "url": "https://res.cloudinary.com/demo/image/upload/v1783049700/timebox-space/development/a0376f53/task/design-api-contract.png",
      "thumbnail_url": "https://res.cloudinary.com/demo/image/upload/c_thumb,w_300/timebox-space/development/a0376f53/task/design-api-contract.png",
      "uploaded_by": {
        "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
        "full_name": "Vincent Chandra"
      },
      "created_at": "2026-07-03T11:40:00+07:00"
    }
  ],
  "meta": null
}
```

---

### 18.4 Delete Attachment

```http
DELETE /api/v1/attachments/:id
```

Auth: uploader/admin

Response 202:

```json
{
  "status": true,
  "message": "Attachment deletion queued",
  "data": {
    "id": "ec968dbe-8f0b-48f8-87b4-687cd722a3d5",
    "delete_asset_queued": true
  },
  "meta": null
}
```

---

## 19. Notification Contract

### 19.1 List Notifications

```http
GET /api/v1/notifications
```

Auth: required

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| status | string | no | unread/read/all |
| type | string | no | trigger type |
| page | int | no | Pagination |
| limit | int | no | Pagination |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "5eaa9aef-5843-41a8-8613-e42c24354d39",
      "type": "timebox_reminder",
      "title": "Timebox akan dimulai",
      "body": "Buat module authentication Golang dimulai dalam 10 menit.",
      "payload": {
        "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9"
      },
      "read_at": null,
      "created_at": "2026-07-04T08:50:00+07:00"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "totalPages": 1
  }
}
```

---

### 19.2 Mark Notification as Read

```http
PATCH /api/v1/notifications/:id/read
```

Auth: notification owner

Request body:

```json
{
  "read": true
}
```

Response 200:

```json
{
  "status": true,
  "message": "Notification updated",
  "data": {
    "id": "5eaa9aef-5843-41a8-8613-e42c24354d39",
    "read_at": "2026-07-04T08:55:00+07:00"
  },
  "meta": null
}
```

---

### 19.3 Mark All Notifications as Read

```http
PATCH /api/v1/notifications/read-all
```

Auth: required

Request body:

```json
{
  "type": null
}
```

Response 200:

```json
{
  "status": true,
  "message": "All notifications marked as read",
  "data": {
    "updated_count": 12
  },
  "meta": null
}
```

---

### 19.4 Get Notification Preferences

```http
GET /api/v1/notifications/preferences
```

Auth: required

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "trigger_type": "timebox_reminder",
      "channels": {
        "in_app": true,
        "email": false,
        "telegram": true
      },
      "reminder_minutes_before": 10
    },
    {
      "trigger_type": "daily_summary",
      "channels": {
        "in_app": true,
        "email": true,
        "telegram": false
      },
      "send_time": "20:00"
    }
  ],
  "meta": null
}
```

---

### 19.5 Update Notification Preferences

```http
PATCH /api/v1/notifications/preferences
```

Auth: required

Request body:

```json
{
  "preferences": [
    {
      "trigger_type": "timebox_reminder",
      "channels": {
        "in_app": true,
        "email": false,
        "telegram": true
      },
      "reminder_minutes_before": 10
    },
    {
      "trigger_type": "daily_summary",
      "channels": {
        "in_app": true,
        "email": true,
        "telegram": false
      },
      "send_time": "20:00"
    }
  ]
}
```

Response 200:

```json
{
  "status": true,
  "message": "Notification preferences updated",
  "data": {
    "updated_count": 2
  },
  "meta": null
}
```

---

## 20. Telegram Integration Contract

### 20.1 Generate Telegram Link Token

```http
POST /api/v1/integrations/telegram/link
```

Auth: required

Request body:

```json
{}
```

Response 200:

```json
{
  "status": true,
  "message": "Telegram link token generated",
  "data": {
    "token": "tlg_01JZ4XRJCPQHDSHFZQNRBXJFKP",
    "deeplink_url": "https://t.me/TimeboxSpaceBot?start=tlg_01JZ4XRJCPQHDSHFZQNRBXJFKP",
    "expires_at": "2026-07-03T11:50:00+07:00"
  },
  "meta": null
}
```

Redis key:

```txt
timebox-space:<env>:telegram_link_token:<token>
```

TTL default: 10 menit.

---

### 20.2 Telegram Webhook

```http
POST /api/v1/integrations/telegram/webhook
```

Auth: Telegram secret header

Header:

```http
X-Telegram-Bot-Api-Secret-Token: <TELEGRAM_WEBHOOK_SECRET>
```

Request body dari Telegram disimpan sesuai struktur update Telegram. Backend minimal memproses command `/start <token>`.

Contoh simplified request:

```json
{
  "update_id": 123456789,
  "message": {
    "message_id": 10,
    "from": {
      "id": 987654321,
      "is_bot": false,
      "first_name": "Vincent",
      "username": "vincent_dev"
    },
    "chat": {
      "id": 987654321,
      "type": "private"
    },
    "date": 1783049700,
    "text": "/start tlg_01JZ4XRJCPQHDSHFZQNRBXJFKP"
  }
}
```

Response 200:

```json
{
  "status": true,
  "message": "Webhook processed",
  "data": {
    "processed": true
  },
  "meta": null
}
```

Efek backend:

- Validasi secret header.
- Parse token dari command `/start`.
- Ambil `user_id` dari Redis token.
- Simpan/update tabel `telegram_links`.
- Matikan token agar tidak bisa dipakai ulang.
- Kirim pesan konfirmasi ke user via Telegram Bot API.

---

### 20.3 Get Telegram Status

```http
GET /api/v1/integrations/telegram/status
```

Auth: required

Response jika terhubung:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "is_connected": true,
    "telegram_username": "vincent_dev",
    "linked_at": "2026-07-03T11:45:00+07:00"
  },
  "meta": null
}
```

Response jika belum terhubung:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "is_connected": false,
    "telegram_username": null,
    "linked_at": null
  },
  "meta": null
}
```

---

### 20.4 Unlink Telegram

```http
DELETE /api/v1/integrations/telegram/unlink
```

Auth: required

Response 200:

```json
{
  "status": true,
  "message": "Telegram account unlinked",
  "data": {
    "is_connected": false
  },
  "meta": null
}
```

---

## 21. Streak & Achievement Contract

### 21.1 Get My Streak

```http
GET /api/v1/streaks/me?workspace_id=a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4
```

Auth: required

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
    "current_streak": 5,
    "longest_streak": 12,
    "last_completed_date": "2026-07-03",
    "badges": [
      {
        "code": "streak_7_days",
        "title": "7 Hari Konsisten",
        "earned_at": "2026-06-30T20:00:00+07:00"
      }
    ]
  },
  "meta": null
}
```

---

### 21.2 Get Workspace Leaderboard

```http
GET /api/v1/workspaces/:wsId/leaderboard?metric=focus_minutes&period=week
```

Auth: workspace member, jika leaderboard aktif

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| metric | string | no | focus_minutes/streak/completed_timeboxes |
| period | string | no | day/week/month |
| limit | int | no | default 10 |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "metric": "focus_minutes",
    "period": "week",
    "items": [
      {
        "rank": 1,
        "user": {
          "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
          "full_name": "Vincent Chandra",
          "avatar_url": null
        },
        "value": 420
      }
    ]
  },
  "meta": null
}
```

Jika leaderboard nonaktif:

```json
{
  "status": false,
  "message": "Leaderboard is disabled for this workspace",
  "error": "forbidden"
}
```

---

## 22. Dashboard Contract

### 22.1 Personal Dashboard

```http
GET /api/v1/dashboard/personal
```

Auth: required

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| workspace_id | uuid | yes | Workspace konteks |
| date | date | no | Default today sesuai timezone user |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "date": "2026-07-04",
    "timezone": "Asia/Jakarta",
    "active_timer": null,
    "today_timeline": [
      {
        "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
        "title": "Buat module authentication Golang",
        "scheduled_start": "2026-07-04T09:30:00+07:00",
        "scheduled_end": "2026-07-04T11:30:00+07:00",
        "status": "planned"
      }
    ],
    "summary": {
      "planned_minutes": 240,
      "actual_minutes": 100,
      "completed_timeboxes": 1,
      "total_timeboxes": 3,
      "completion_rate": 33.33,
      "current_streak": 5
    },
    "category_distribution": [
      {
        "category_id": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
        "category_name": "Deep Work",
        "planned_minutes": 180,
        "actual_minutes": 100
      }
    ],
    "overrun_or_missed": []
  },
  "meta": null
}
```

---

### 22.2 Workspace Dashboard

```http
GET /api/v1/dashboard/workspace
```

Auth: owner/admin/member terbatas

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| workspace_id | uuid | yes | Workspace konteks |
| date | date | no | Default today |
| team_id | uuid | no | Filter team |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "date": "2026-07-04",
    "summary": {
      "team_focus_minutes_today": 820,
      "team_focus_minutes_week": 3840,
      "completion_rate_today": 68.5,
      "active_members_now": 3,
      "overrun_risk_members": 1
    },
    "team_today": [
      {
        "user": {
          "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
          "full_name": "Vincent Chandra",
          "avatar_url": null
        },
        "current_timebox": {
          "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
          "title": "Buat module authentication Golang",
          "status": "running"
        },
        "planned_minutes": 240,
        "actual_minutes": 100,
        "completion_rate": 50
      }
    ],
    "category_distribution": [],
    "trend": []
  },
  "meta": null
}
```

---

## 23. Report Contract

### 23.1 Time Allocation Report

```http
GET /api/v1/reports/time-allocation
```

Auth: required

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| workspace_id | uuid | yes | Workspace konteks |
| start_date | date | yes | Awal rentang |
| end_date | date | yes | Akhir rentang |
| group_by | string | no | category/goal/tag/user/day |
| user_id | uuid | no | Filter user |
| team_id | uuid | no | Filter team |
| category_id | uuid | no | Filter kategori |
| goal_id | uuid | no | Filter goal |
| format | string | no | json/pdf/xlsx/csv |

Response JSON 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "range": {
      "start_date": "2026-07-01",
      "end_date": "2026-07-07"
    },
    "group_by": "category",
    "summary": {
      "planned_minutes": 1200,
      "actual_minutes": 980,
      "variance_minutes": -220
    },
    "items": [
      {
        "key": "bb6f6b19-6c13-4f90-8f88-90d2a21875e1",
        "label": "Deep Work",
        "planned_minutes": 720,
        "actual_minutes": 650,
        "variance_minutes": -70,
        "percentage": 66.32
      }
    ]
  },
  "meta": null
}
```

Jika `format=pdf|xlsx|csv`, response dapat berupa signed download URL:

```json
{
  "status": true,
  "message": "Report generated",
  "data": {
    "format": "xlsx",
    "download_url": "https://res.cloudinary.com/demo/raw/upload/reports/time-allocation.xlsx",
    "expires_at": "2026-07-04T12:00:00+07:00"
  },
  "meta": null
}
```

---

### 23.2 Productivity Trend Report

```http
GET /api/v1/reports/productivity-trend
```

Auth: required

Query:

```txt
workspace_id=<uuid>&start_date=2026-07-01&end_date=2026-07-31&interval=day
```

`interval`:

```txt
day | week | month
```

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "interval": "day",
    "items": [
      {
        "date": "2026-07-01",
        "planned_minutes": 240,
        "actual_minutes": 210,
        "completed_timeboxes": 3,
        "total_timeboxes": 4,
        "completion_rate": 75
      }
    ]
  },
  "meta": null
}
```

---

### 23.3 Team Workload Report

```http
GET /api/v1/reports/team-workload
```

Auth: owner/admin/member terbatas

Query:

```txt
workspace_id=<uuid>&team_id=<uuid>&start_date=2026-07-01&end_date=2026-07-07
```

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "range": {
      "start_date": "2026-07-01",
      "end_date": "2026-07-07"
    },
    "items": [
      {
        "user": {
          "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
          "full_name": "Vincent Chandra"
        },
        "planned_minutes": 1200,
        "actual_minutes": 980,
        "completed_timeboxes": 18,
        "overrun_count": 2,
        "risk_level": "normal"
      }
    ]
  },
  "meta": null
}
```

---

## 24. Search & Saved View Contract

### 24.1 Global Search

```http
GET /api/v1/search
```

Auth: required

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| workspace_id | uuid | yes | Workspace konteks |
| q | string | yes | Keyword |
| types | string | no | goal,task,timebox,comment,attachment dipisah koma |
| limit | int | no | Default 10 per type |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": {
    "query": "auth",
    "results": {
      "tasks": [
        {
          "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
          "title": "Buat module authentication Golang",
          "snippet": "Auth lengkap JWT access + refresh token."
        }
      ],
      "timeboxes": [],
      "goals": [],
      "comments": [],
      "attachments": []
    }
  },
  "meta": null
}
```

---

### 24.2 Create Saved View

```http
POST /api/v1/saved-views
```

Auth: required

Request body:

```json
{
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "name": "Backend High Priority",
  "resource_type": "task",
  "filter_json": {
    "status": [
      "backlog",
      "in_progress"
    ],
    "priority": [
      "high",
      "urgent"
    ],
    "tag_ids": [
      "9a801852-8742-4d04-ad15-974b4f252db6"
    ]
  },
  "shared": false
}
```

Response 201:

```json
{
  "status": true,
  "message": "Saved view created",
  "data": {
    "id": "35e9bda8-c502-47e2-9952-4353a91f9980",
    "name": "Backend High Priority",
    "resource_type": "task",
    "shared": false
  },
  "meta": null
}
```

CRUD lain:

```http
GET    /api/v1/saved-views?workspace_id=:wsId&resource_type=task
PATCH  /api/v1/saved-views/:id
DELETE /api/v1/saved-views/:id
```

---

## 25. Activity Log & Audit Trail Contract

### 25.1 List Activity Logs

```http
GET /api/v1/activity-logs
```

Auth: owner/admin untuk workspace, super_admin global

Query:

| Parameter | Type | Required | Keterangan |
|---|---|---:|---|
| workspace_id | uuid | no | Required untuk non super admin |
| actor_id | uuid | no | Filter actor |
| action | string | no | Filter action |
| resource_type | string | no | Filter resource |
| resource_id | uuid | no | Filter resource id |
| start_date | date | no | Filter tanggal |
| end_date | date | no | Filter tanggal |

Response 200:

```json
{
  "status": true,
  "message": "data fetched",
  "data": [
    {
      "id": "7b7f47bd-490d-45fe-92bc-f4356dcdfe39",
      "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
      "actor": {
        "id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
        "full_name": "Vincent Chandra"
      },
      "action": "timebox.updated",
      "resource_type": "timebox",
      "resource_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
      "old_value": {
        "scheduled_start": "2026-07-04T09:00:00+07:00"
      },
      "new_value": {
        "scheduled_start": "2026-07-04T09:30:00+07:00"
      },
      "ip_address": "127.0.0.1",
      "user_agent": "Mozilla/5.0",
      "created_at": "2026-07-03T11:20:00+07:00"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "totalPages": 1
  }
}
```

---

## 26. WebSocket Contract

### 26.1 Connect WebSocket

```http
GET /ws?token=<access_token>
```

Auth:

- Access token dikirim via query parameter saat upgrade.
- Backend memvalidasi JWT.
- Setelah valid, koneksi masuk ke hub user.
- Client perlu subscribe workspace.

Client message subscribe:

```json
{
  "event": "workspace.subscribe",
  "payload": {
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4"
  }
}
```

Server response:

```json
{
  "event": "workspace.subscribed",
  "payload": {
    "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4"
  },
  "server_time": "2026-07-03T11:45:00+07:00"
}
```

### 26.2 Event Format

```json
{
  "event": "timebox.updated",
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "payload": {},
  "server_time": "2026-07-03T11:45:00+07:00"
}
```

### 26.3 Event List

| Event | Arah | Keterangan |
|---|---|---|
| `workspace.subscribe` | client → server | Join room workspace |
| `workspace.subscribed` | server → client | Subscribe sukses |
| `timebox.updated` | server → client | Timebox dibuat/diubah/dihapus/status berubah |
| `task.updated` | server → client | Task berubah, termasuk move Kanban |
| `timer.sync` | server → client | State timer aktif berubah |
| `notification.new` | server → client | Notifikasi baru |
| `member.presence` | server → client | Presence member berubah |
| `error` | server → client | Error di koneksi/event |

### 26.4 timebox.updated Payload

```json
{
  "event": "timebox.updated",
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "payload": {
    "action": "updated",
    "timebox": {
      "id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
      "title": "Buat module auth dan refresh token",
      "scheduled_start": "2026-07-04T09:30:00+07:00",
      "scheduled_end": "2026-07-04T11:30:00+07:00",
      "status": "planned"
    }
  },
  "server_time": "2026-07-03T11:45:00+07:00"
}
```

### 26.5 task.updated Payload

```json
{
  "event": "task.updated",
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "payload": {
    "action": "moved",
    "task": {
      "id": "d337fdbe-264e-4f42-bcf9-ae5dfe2499c9",
      "status": "in_progress",
      "position": 2000
    }
  },
  "server_time": "2026-07-03T11:45:00+07:00"
}
```

### 26.6 timer.sync Payload

```json
{
  "event": "timer.sync",
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4",
  "payload": {
    "timebox_id": "2f0d08fb-f1ca-4e9e-a78c-2f63f58c3fb9",
    "status": "running",
    "started_at": "2026-07-04T09:30:00+07:00",
    "elapsed_seconds": 900,
    "remaining_seconds": 6300
  },
  "server_time": "2026-07-04T09:45:00+07:00"
}
```

---

## 27. Health Check Contract

Health endpoint mengikuti struktur default project:

```text
GET /api/v1/health/
GET /api/v1/health/live
GET /api/v1/health/ready
```

`/api/v1/health/live` hanya memastikan proses HTTP hidup. `/api/v1/health/ready` dan `/api/v1/health/` mengecek kesiapan dependency database/Redis sesuai kebutuhan fitur.

### 27.1 Liveness

```http
GET /api/v1/health/live
```

Auth: public/internal

Response 200:

```json
{
  "status": true,
  "message": "alive",
  "data": {
    "service": "timebox-space-api",
    "version": "1.0.0",
    "env": "development"
  }
}
```

---

### 27.2 Readiness

```http
GET /api/v1/health/ready
```

Auth: public/internal

Response 200 jika dependency siap:

```json
{
  "status": true,
  "message": "ready",
  "data": {
    "postgres": "ok",
    "redis": "ok"
  }
}
```

Response 503 jika ada dependency gagal:

```json
{
  "status": false,
  "message": "not ready",
  "error": "dependency unavailable"
}
```

---

## 28. Permission Rules Ringkas

| Resource | Owner | Admin | Member | Viewer |
|---|---:|---:|---:|---:|
| Workspace settings | full | update terbatas | read | read terbatas |
| Member management | full | full kecuali owner | no | no |
| Goal | full | full | own/create | read terbatas |
| Task | full | full | own/assigned | read |
| Kanban move | full | full | own/assigned | no |
| Timebox | full | full/optional edit others | own | read terbatas |
| Timer | own/all if super admin | own only | own only | no |
| Report personal | full | full | own | no |
| Report workspace | full | full | terbatas | terbatas |
| Activity log | full | read | no | no |
| Telegram link | own | own | own | own optional |

Catatan implementasi:

- Permission dicek di service layer, bukan hanya middleware.
- Handler hanya melakukan binding, auth context extraction, dan response mapping.
- Repository tidak boleh memutuskan permission bisnis.

---

## 29. Catatan Implementasi Golang

### 29.1 Struktur Handler Gin

Contoh grouping route:

```go
v1 := router.Group("/api/v1")
{
    auth := v1.Group("/auth")
    auth.POST("/register", authHandler.Register)
    auth.POST("/login", authHandler.Login)
    auth.POST("/refresh", authHandler.Refresh)
    auth.POST("/logout", authMiddleware, authHandler.Logout)

    workspace := v1.Group("/workspaces", authMiddleware)
    workspace.GET("", workspaceHandler.List)
    workspace.POST("", workspaceHandler.Create)
    workspace.GET("/:id", workspaceHandler.Detail)
    workspace.PATCH("/:id", workspaceHandler.Update)

    task := v1.Group("/tasks", authMiddleware)
    task.GET("/:id", taskHandler.Detail)
    task.PATCH("/:id", taskHandler.Update)
    task.DELETE("/:id", taskHandler.Delete)
    task.PATCH("/:id/move", taskHandler.Move)
}
```

### 29.2 DTO Naming Convention

```txt
RegisterRequest
RegisterResponse
CreateWorkspaceRequest
WorkspaceResponse
CreateTaskRequest
TaskResponse
MoveTaskRequest
CreateTimeboxRequest
TimeboxResponse
StartTimerRequest
TimerStateResponse
```

### 29.3 Repository Rule

- Semua query memakai `context.Context`.
- Gunakan `sqlx.GetContext`, `sqlx.SelectContext`, `NamedExecContext`.
- Operasi multi-step memakai transaksi eksplisit.
- Query list wajib support pagination dan filter.
- Query Kanban wajib memakai index `(workspace_id, status, position)`.
- Query planner wajib memakai index `(workspace_id, owner_id, scheduled_start, scheduled_end)`.

### 29.4 Redis Key Convention

```txt
timebox-space:<env>:refresh_token:<token_id>
timebox-space:<env>:timer:<user_id>
timebox-space:<env>:rate_limit:login:<ip>
timebox-space:<env>:telegram_link_token:<token>
timebox-space:<env>:ws:channel:notif:<workspace_id>
timebox-space:<env>:cache:dashboard:<workspace_id>:<user_id>:<date>
timebox-space:<env>:leaderboard:<workspace_id>:<period>
```

### 29.5 Zap Logging Minimum Fields

Setiap request log minimal memiliki:

```json
{
  "level": "info",
  "ts": "2026-07-03T11:45:00+07:00",
  "method": "POST",
  "path": "/api/v1/timeboxes",
  "status": 201,
  "latency_ms": 42,
  "user_id": "2cfa3e8d-84c7-4e3e-90d2-7f8f7616f9e1",
  "workspace_id": "a0376f53-e4d0-46fc-8c4f-6b4a60d3b4d4"
}
```

Error log minimal:

```json
{
  "level": "error",
  "error": "pq: duplicate key value violates unique constraint",
  "code": "CONFLICT",
  "resource": "workspace",
  "operation": "create"
}
```

---

## 30. Checklist Implementasi MVP API

### Phase 1 — Foundation

- Auth register/login/refresh/logout.
- Middleware JWT, request id, recovery, logger Zap.
- Config Viper.
- PostgreSQL connection sqlx + pgx/stdlib.
- Redis connection go-redis.
- Health check.

### Phase 2 — Workspace & Access

- Workspace CRUD.
- Member invite/update.
- Team CRUD.
- Permission service.

### Phase 3 — Core Planning

- Category/tag.
- Goal CRUD.
- Task CRUD.
- Kanban board + move task.
- Timebox CRUD.
- Planner query day/week.

### Phase 4 — Execution

- Timer active/start/pause/resume/complete/skip.
- Time log.
- Redis timer state.
- WebSocket event `timer.sync`, `timebox.updated`, `task.updated`.

### Phase 5 — Collaboration & Notification

- Comments and mentions.
- Cloudinary signed upload + attachment.
- Notification preferences.
- In-app notification.
- Telegram link/webhook/unlink/status.
- Worker notification.

### Phase 6 — Analytics

- Dashboard personal/workspace.
- Streak.
- Reports.
- Saved view.
- Activity log.

---

## 31. Endpoint Summary

```txt
AUTH
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password

PROFILE
GET    /api/v1/me
PATCH  /api/v1/me
PATCH  /api/v1/me/password
GET    /api/v1/me/sessions
DELETE /api/v1/me/sessions/:sessionId

WORKSPACE
GET    /api/v1/workspaces
POST   /api/v1/workspaces
GET    /api/v1/workspaces/:id
PATCH  /api/v1/workspaces/:id
POST   /api/v1/workspaces/:id/invite
GET    /api/v1/workspaces/:id/members
PATCH  /api/v1/workspaces/:id/members/:userId

TEAM
GET    /api/v1/workspaces/:wsId/teams
POST   /api/v1/workspaces/:wsId/teams
PATCH  /api/v1/teams/:id
DELETE /api/v1/teams/:id

GOAL
GET    /api/v1/workspaces/:wsId/goals
POST   /api/v1/workspaces/:wsId/goals
GET    /api/v1/goals/:id
PATCH  /api/v1/goals/:id
DELETE /api/v1/goals/:id

TASK & KANBAN
GET    /api/v1/workspaces/:wsId/tasks
POST   /api/v1/workspaces/:wsId/tasks
GET    /api/v1/tasks/:id
PATCH  /api/v1/tasks/:id
DELETE /api/v1/tasks/:id
PATCH  /api/v1/tasks/:id/move
POST   /api/v1/tasks/:id/convert-to-timebox
GET    /api/v1/workspaces/:wsId/kanban
PATCH  /api/v1/workspaces/:wsId/kanban/settings

TIMEBOX & TIMER
GET    /api/v1/timeboxes
POST   /api/v1/timeboxes
GET    /api/v1/timeboxes/:id
PATCH  /api/v1/timeboxes/:id
DELETE /api/v1/timeboxes/:id
POST   /api/v1/timeboxes/:id/duplicate
POST   /api/v1/timeboxes/:id/move-to-next-day
GET    /api/v1/timer/active
POST   /api/v1/timeboxes/:id/start
POST   /api/v1/timeboxes/:id/pause
POST   /api/v1/timeboxes/:id/resume
POST   /api/v1/timeboxes/:id/complete
POST   /api/v1/timeboxes/:id/skip

TIME LOG
GET    /api/v1/timeboxes/:id/logs
POST   /api/v1/timeboxes/:id/logs
PATCH  /api/v1/timebox-logs/:id
DELETE /api/v1/timebox-logs/:id

RECURRING & TEMPLATE
POST   /api/v1/workspaces/:wsId/recurrence-rules
PATCH  /api/v1/timeboxes/:id/recurring-instance
POST   /api/v1/workspaces/:wsId/routine-templates
POST   /api/v1/routine-templates/:id/apply

CATEGORY & TAG
GET    /api/v1/workspaces/:wsId/categories
POST   /api/v1/workspaces/:wsId/categories
PATCH  /api/v1/categories/:id
DELETE /api/v1/categories/:id
GET    /api/v1/tags
POST   /api/v1/tags
PATCH  /api/v1/tags/:id
DELETE /api/v1/tags/:id

COMMENT & ATTACHMENT
GET    /api/v1/comments
POST   /api/v1/comments
PATCH  /api/v1/comments/:id
DELETE /api/v1/comments/:id
POST   /api/v1/uploads/signature
GET    /api/v1/attachments
POST   /api/v1/attachments
DELETE /api/v1/attachments/:id

NOTIFICATION & TELEGRAM
GET    /api/v1/notifications
PATCH  /api/v1/notifications/:id/read
PATCH  /api/v1/notifications/read-all
GET    /api/v1/notifications/preferences
PATCH  /api/v1/notifications/preferences
POST   /api/v1/integrations/telegram/link
POST   /api/v1/integrations/telegram/webhook
GET    /api/v1/integrations/telegram/status
DELETE /api/v1/integrations/telegram/unlink

DASHBOARD, REPORT, SEARCH
GET    /api/v1/dashboard/personal
GET    /api/v1/dashboard/workspace
GET    /api/v1/reports/time-allocation
GET    /api/v1/reports/productivity-trend
GET    /api/v1/reports/team-workload
GET    /api/v1/search
GET    /api/v1/saved-views
POST   /api/v1/saved-views
PATCH  /api/v1/saved-views/:id
DELETE /api/v1/saved-views/:id
GET    /api/v1/activity-logs

REALTIME & HEALTH
GET    /ws?token=<access_token>
GET    /api/v1/health/live
GET    /api/v1/health/ready
```
