# Mini Task Management System — indonesian-cloud

Aplikasi manajemen tugas sederhana berbasis web: **CRUD Task** (Create, Read, Update, Delete) dengan filter status (`all` / `pending` / `completed`), dibangun dengan arsitektur fullstack terpisah — REST API di Go dan SPA di React.

## Tech Stack

**Backend** (`indo-cloud-be/`) — REST API
| Komponen | Teknologi |
|---|---|
| Bahasa | Go 1.27 |
| Router | [chi v5](https://github.com/go-chi/chi) |
| CORS | go-chi/cors |
| Logging | go-chi/httplog v3 (ECS JSON) + `log/slog` |
| Database | PostgreSQL |
| DB Driver | [pgx v5](https://github.com/jackc/pgx) (connection pool) |
| DB Migration | [golang-migrate v4](https://github.com/golang-migrate/migrate) (auto-migrate saat start) |
| Konfigurasi | joho/godotenv (`.env`) |
| ID | google/uuid (UUIDv7 primary key) |

**Frontend** (`indo-cloud-fe/`) — SPA
| Komponen | Teknologi |
|---|---|
| Framework | React 19 + TypeScript |
| Build Tool | Vite |
| Styling | Tailwind CSS v4 |
| UI Kit | shadcn/ui (Button, Card, Input) + lucide-react |
| Font | Inter Variable |
| Kualitas kode | ESLint, Prettier, `tsc --noEmit` (typecheck) |

## Struktur Proyek

```
indonesian-cloud/
├── README.md
├── indo-cloud-be/                     # Backend (Go REST API)
│   ├── main.go                        # Entry point + handler CRUD
│   ├── config.go                      # Load konfigurasi dari .env
│   ├── database.go                    # Koneksi PostgreSQL (pgx pool)
│   ├── go.mod
│   ├── .env                           # Konfigurasi lokal (tidak di-commit)
│   ├── .env.example
│   └── database/migration/
│       ├── 000001_initial_schema.up.sql
│       └── 000001_initial_schema.down.sql
└── indo-cloud-fe/                     # Frontend (React SPA)
    ├── package.json
    └── src/
        ├── App.tsx                    # UI CRUD task + filter status
        ├── components/ui/             # shadcn/ui components
        └── index.css                  # Tailwind v4 theme tokens
```

## API Documentation

Base URL: `http://localhost:8080`

### Task Model
```json
{
  "id": "019b2c3d-4e5f-7a8b-9c0d-1e2f3a4b5c6d",
  "title": "Belajar Go",
  "status": "pending",
  "created_at": "2026-01-15T08:30:00Z"
}
```
- `id` — UUIDv7, dibuat otomatis oleh database
- `title` — wajib, maksimal 255 karakter
- `status` — `pending` (default) atau `completed`

### Endpoints

#### 1. GET /api/tasks — Daftar Task (dengan filter)
```
GET /api/tasks?status=all          # semua task
GET /api/tasks?status=pending      # hanya task pending
GET /api/tasks?status=completed    # hanya task completed
GET /api/tasks                     # kosong = all
```
Response `200 OK` — array JSON, diurutkan `created_at DESC`:
```json
[
  { "id": "019b...", "title": "Belajar Go", "status": "pending", "created_at": "2026-01-15T08:30:00Z" }
]
```
Query `status` tidak valid → `400 Bad Request`.

#### 2. POST /api/tasks — Buat Task
```json
{ "title": "Belajar Go", "status": "pending" }
```
- `title` wajib, maksimal 255 karakter; kosong → `400`
- `status` opsional; kosong → `pending`; selain pending/completed → `400`

Response `201 Created`.

#### 3. PATCH /api/tasks/{id} — Update Task (dinamis)
Hanya field yang dikirim yang di-update (`COALESCE`); id harus UUID valid → selain itu `400`.
```json
{ "status": "completed" }
```
```json
{ "title": "Judul baru" }
```
Response `200 OK` — `{"message": "Task updated successfully"}`.

#### 4. DELETE /api/tasks/{id} — Hapus Task
Response `200 OK` — `{"message": "Task deleted successfully"}`.

### Contoh Penggunaan (curl)
```bash
# Buat task
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Belajar Go", "status": "pending"}'

# Daftar semua task
curl http://localhost:8080/api/tasks?status=all

# Tandai selesai
curl -X PATCH http://localhost:8080/api/tasks/019b2c3d-4e5f-7a8b-9c0d-1e2f3a4b5c6d \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'

# Hapus task
curl -X DELETE http://localhost:8080/api/tasks/019b2c3d-4e5f-7a8b-9c0d-1e2f3a4b5c6d
```

## Setup & Menjalankan

### Prasyarat
- Go ≥ 1.27
- Node.js ≥ 20 + npm
- PostgreSQL ≥ 16 (mendukung `uuidv7()`)

### 1. Database
Buat database — skema tabel **tidak perlu dibuat manual**, migrasi dijalankan otomatis oleh backend:
```sql
CREATE DATABASE indocloud;
```

### 2. Backend
Backend menggunakan **golang-migrate** untuk migrasi skema database. Saat `DB_MIGRATE=true`, migrasi dijalankan otomatis setiap kali server start (idempotent — skema yang sudah terbaru tidak akan diubah ulang).

File migrasi berada di `database/migration/` mengikuti konvensi golang-migrate:
```
database/migration/
├── 000001_initial_schema.up.sql      # skema tabel tasks
└── 000001_initial_schema.down.sql    # rollback skema tabel tasks
```
```bash
cd indo-cloud-be
cp .env.example .env          # sesuaikan DB_NAME/USER/PASSWORD/HOST/PORT
go mod download
go run .                      # migrasi otomatis + server jalan di :8080
```

Konfigurasi migrasi di `.env`:
```env
DB_MIGRATE=true                        # jalankan migrasi otomatis saat start
DB_MIGRATIONS_PATH=database/migration  # direktori file migrasi (opsional, default "database/migration")
```

> `DB_MIGRATE` bersifat opsional; jika dihilangkan, backend start tanpa menjalankan migrasi — gunakan ini di production jika migrasi dikelola lewat CI/CD atau tool terpisah.


### 3. Frontend
```bash
cd indo-cloud-fe
npm install
npm run dev                   # SPA jalan di :5173 (dev)
```

## Fitur

- ✅ **CRUD Task** lengkap (buat, baca, update dinamis, hapus)
- ✅ **Filter status** task: `all` / `pending` / `completed`
- ✅ **Validasi input**: title wajib (maks 255 karakter), status enum terbatas
- ✅ **UUIDv7** sebagai primary key (sortable by time)
- ✅ **Database migration otomatis** via golang-migrate (`DB_MIGRATE=true`, idempotent)
- ✅ **Update dinamis** via PATCH dengan `COALESCE` (hanya field yang dikirim yang berubah)
- ✅ **Request logging** ECS JSON + CORS di backend
- ✅ **UI responsif** dengan Tailwind v4 + shadcn/ui (layout mobile-first, segmented control filter, status badge, empty state)
