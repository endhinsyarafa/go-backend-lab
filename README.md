# Meeting Room Booking API

REST API untuk manajemen ruangan meeting dan pemesanan jadwal berbasis **Go (Gin)** dan **PostgreSQL** dengan proteksi *double booking* menggunakan transaksi database dan *row-level locking* (`SELECT ... FOR UPDATE`).

## Fitur Utama & Tech Stack
* **Bahasa & Framework:** Go, Gin Web Framework
* **Database:** PostgreSQL 16 (`lib/pq`)
* **Autentikasi & Otorisasi:** JWT (Bearer Token), Bcrypt password hashing, Role-Based Access Control (`ADMIN` & `USER`)
* **Concurrency Control:** Database Transaction + Row-Level Locking untuk mencegah bentrok jadwal pada request paralel
* **Dokumentasi API:** OpenAPI 3.0 (`swagger.yaml`) & Swagger UI terintegrasi
* **Containerization:** Docker & Docker Compose

## Struktur Proyek
```text
.
├── cmd/api/main.go               # Entry point, dependency injection, & routing
├── internal/
│   ├── config/                   # Loader environment variable & DSN database
│   ├── handler/                  # HTTP handler & integration/concurrency test
│   ├── middleware/               # Verifikasi JWT & validasi role (ADMIN/USER)
│   ├── model/                    # Struct domain & payload request/response
│   ├── repository/               # Query SQL PostgreSQL & row-level locking
│   └── service/                  # Logika bisnis & validasi aturan booking
├── queries.sql                   # Query SQL analitik (rekap jam & ketersediaan)
├── schema.sql                    # DDL tabel, index, & seeder 3 akun awal
├── swagger.yaml                  # Spesifikasi OpenAPI 3.0
├── Dockerfile                    # Multi-stage build image Go
└── docker-compose.yml            # Orkestrasi container API & PostgreSQL
```

---

## 1. Menjalankan Aplikasi dengan Docker
Pastikan Docker Desktop sudah aktif dan port `8080` serta `5432` tidak sedang digunakan, lalu jalankan:

```bash
docker-compose up -d --build
```

Aplikasi akan berjalan di `http://localhost:8080` dan database PostgreSQL akan otomatis terinisialisasi menggunakan `schema.sql`.

### Akun Seeder Awal
| Role | Email | Password |
| :--- | :--- | :--- |
| `ADMIN` | `admin@test.com` | `password123` |
| `USER` | `user1@test.com` | `password123` |
| `USER` | `user2@test.com` | `password123` |

Untuk menghentikan container dan menghapus volume database:
```bash
docker-compose down -v
```

---

## 2. Dokumentasi Interaktif (Swagger UI)
Saat aplikasi berjalan, dokumentasi interaktif OpenAPI 3.0 dapat diakses langsung melalui browser:
* **Swagger UI:** `http://localhost:8080/swagger/index.html`
* **Spesifikasi YAML:** `http://localhost:8080/swagger.yaml`

> **Tips Pengujian di Swagger UI:** Lakukan login di endpoint `/api/v1/auth/login`, salin nilai `token`, lalu klik tombol **Authorize** dan masukkan **hanya string token-nya saja** (tanpa awalan `Bearer `).

---

## 3. Menjalankan Automated Test
Proyek ini dilengkapi dengan **unit test** (`config`, `middleware`, `service`) serta **integration & concurrency test** (`handler` dengan simulasi 10 request paralel pada slot waktu yang sama).

Pastikan database aktif, lalu jalankan:
```bash
go test -v -count=1 ./...
```
*(Flag `-count=1` memastikan seluruh test dieksekusi ulang tanpa menggunakan cache).*

---

## 4. Query SQL Analitik (`queries.sql`)
File `queries.sql` di root repositori berisi 2 query siap eksekusi di PostgreSQL:
1. **Rekap jam penggunaan per ruangan** untuk bulan berjalan (hanya menghitung status `CONFIRMED`).
2. **Pencarian ruangan tersedia** pada rentang waktu tertentu (ruangan aktif tanpa irisan jadwal `CONFIRMED`).

---

## 5. Dokumentasi Endpoint (`curl`)
* **Autentikasi:** Semua endpoint terproteksi memerlukan header `Authorization: Bearer <jwt_token>`.
* **Format Waktu:** Semua parameter query dan field waktu menggunakan format **ISO-8601 UTC** (contoh: `2026-12-01T10:00:00Z`).

### Health Check (`GET /health`)
```bash
curl http://localhost:8080/health
```
* **Sukses (`200 OK`):** `{"status":"ok"}`

### 1. Login (`POST /api/v1/auth/login`)
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"password123"}'
```
* **Sukses (`200 OK`):**
  ```json
  {"token":"eyJhbGciOiJIUzI1NiIs..."}
  ```
* **Gagal (`401 Unauthorized`):**
  ```json
  {"error":"invalid credentials"}
  ```

### 2. Create Room (`POST /api/v1/rooms`) — Khusus ADMIN
```bash
curl -X POST http://localhost:8080/api/v1/rooms \
  -H "Authorization: Bearer <TOKEN_ADMIN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Ruang Meeting A","capacity":10}'
```
* **Sukses (`201 Created`):**
  ```json
  {"id":1,"name":"Ruang Meeting A","capacity":10,"is_active":true}
  ```
* **Gagal (`403 Forbidden`):**
  ```json
  {"error":"forbidden"}
  ```

### 3. Room Availability (`GET /api/v1/rooms/available`)
```bash
curl "http://localhost:8080/api/v1/rooms/available?start=2026-12-01T10:00:00Z&end=2026-12-01T12:00:00Z" \
  -H "Authorization: Bearer <TOKEN>"
```
* **Sukses (`200 OK`):**
  ```json
  {"data":[{"id":1,"name":"Ruang Meeting A","capacity":10}]}
  ```
* **Gagal (`400 Bad Request`):**
  ```json
  {"error":"end must be after start"}
  ```

### 4. Create Booking (`POST /api/v1/bookings`)
```bash
curl -X POST http://localhost:8080/api/v1/bookings \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"room_id":1,"title":"Sprint Planning","start_time":"2026-12-01T10:00:00Z","end_time":"2026-12-01T12:00:00Z"}'
```
* **Sukses (`201 Created`):**
  ```json
  {"id":1,"room_id":1,"user_id":1,"title":"Sprint Planning","start_time":"2026-12-01T10:00:00Z","end_time":"2026-12-01T12:00:00Z","status":"CONFIRMED"}
  ```
* **Bentrok (`409 Conflict`):**
  ```json
  {"error_code":"ROOM_CONFLICT","message":"Ruangan sudah dipesan pada rentang waktu tersebut"}
  ```
* **Waktu Tidak Valid (`400 Bad Request`):**
  ```json
  {"error_code":"INVALID_TIME","message":"Waktu mulai harus di masa depan dan waktu selesai harus setelah waktu mulai"}
  ```

### 5. Cancel Booking (`PATCH /api/v1/bookings/{id}/cancel`)
```bash
curl -X PATCH http://localhost:8080/api/v1/bookings/1/cancel \
  -H "Authorization: Bearer <TOKEN>"
```
* **Sukses (`200 OK`):**
  ```json
  {"id":1,"status":"CANCELLED"}
  ```
* **Tidak Ditemukan (`404 Not Found`):**
  ```json
  {"error":"booking not found"}
  ```
