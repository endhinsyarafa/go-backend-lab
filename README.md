# Meeting Room Booking API

REST API untuk manajemen ruangan meeting dan pemesanan jadwal berbasis Go (Gin) dan PostgreSQL.

## Menjalankan Aplikasi dengan Docker
Pastikan Docker Desktop sudah aktif, lalu jalankan:

```bash
docker-compose up -d --build
```

Aplikasi akan berjalan di `http://localhost:8080` dan database PostgreSQL akan otomatis terinisialisasi menggunakan `schema.sql` beserta 3 akun awal (`admin@test.com`, `user1@test.com`, `user2@test.com` dengan password `password123`).

Untuk menghentikan dan menghapus volume database:

```bash
docker-compose down -v
```

## Menjalankan Automated Test
Pastikan database aktif (baik lokal maupun Docker) dan konfigurasi environment sudah sesuai, lalu jalankan:

```bash
go test -v ./...
```

## Dokumentasi Endpoint
* **Autentikasi:** Semua endpoint terproteksi memerlukan header `Authorization: Bearer <jwt_token>`.
* **Format Waktu:** Semua parameter query dan field waktu menggunakan format **ISO-8601 UTC** (contoh: `2026-12-01T10:00:00Z`).

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
