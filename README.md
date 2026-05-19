# ZyCrypt — Ziya Encryptor License Manager

Sistem manajemen lisensi berbasis Go untuk produk **Laravel + Vue + Inertia.js**. Beroperasi sebagai single binary CLI yang ringan — tidak butuh runtime tambahan.

---

## Daftar Isi

1. [Persyaratan](#1-persyaratan)
2. [Instalasi](#2-instalasi)
3. [Konfigurasi](#3-konfigurasi)
4. [Setup Database](#4-setup-database)
5. [Manajemen Plan](#5-manajemen-plan)
6. [Registrasi Lisensi](#6-registrasi-lisensi)
7. [Manajemen Lisensi](#7-manajemen-lisensi)
8. [Manajemen Domain](#8-manajemen-domain)
9. [Utilitas Key](#9-utilitas-key)
10. [Menjalankan Server API](#10-menjalankan-server-api)
11. [API Endpoint](#11-api-endpoint)
12. [Referensi Lengkap CLI](#12-referensi-lengkap-cli)

---

## 1. Persyaratan

| Kebutuhan | Versi Minimum |
|---|---|
| Go | 1.22+ |
| PostgreSQL | 15+ |

---

## 2. Instalasi

### Build dari source

```bash
git clone https://github.com/ziaulkamal/zycrypt.git
cd zycrypt
go mod tidy
go build -o zycrypt ./main.go
```

### Windows

```powershell
go build -o zycrypt.exe ./main.go
```

Binary siap digunakan di direktori yang sama.

---

## 3. Konfigurasi

Buat file `zycrypt.yaml` di direktori yang sama dengan binary (atau di `/etc/zycrypt/zycrypt.yaml` untuk production):

```yaml
server:
  host: 0.0.0.0
  port: 8743
  base_url: https://zycrypt.yourdomain.com

database:
  dsn: postgres://user:password@localhost:5432/zycrypt?sslmode=disable

security:
  shared_secret: "isi-dengan-string-random-minimal-32-karakter"
  token_ttl_minutes: 10
  grace_period_hours: 24
  min_pkg_version: "1.0.0"

storage:
  themes_path: /var/zycrypt/themes

rate_limit:
  requests_per_minute: 60
```

### Override via Environment Variable

Setiap nilai di `zycrypt.yaml` bisa di-override menggunakan env variable dengan prefix `ZYCRYPT_`:

```bash
ZYCRYPT_SERVER_PORT=9000
ZYCRYPT_DATABASE_DSN=postgres://user:pass@localhost:5432/zycrypt
ZYCRYPT_SECURITY_SHARED_SECRET=rahasia-baru
```

### Gunakan config file kustom

```bash
./zycrypt --config /path/ke/config.yaml serve
```

---

## 4. Setup Database

### Buat database PostgreSQL

```bash
# Via psql
createdb zycrypt

# Atau via Docker
docker exec nama_container psql -U user -d postgres -c "CREATE DATABASE zycrypt;"
```

### Jalankan migrasi

```bash
./zycrypt db migrate
```

Output:
```
Running migrations...
  ✓ Applied: 001_create_plans.sql
  ✓ Applied: 002_create_licenses.sql
  ✓ Applied: 003_create_domains.sql
  ✓ Applied: 004_create_audit_logs.sql
✓ Migrations complete
```

Migrasi bersifat idempotent — aman dijalankan berkali-kali.

---

## 5. Manajemen Plan

Plan adalah template lisensi yang mendefinisikan jumlah domain, durasi, dan harga.

### Buat plan baru

```bash
./zycrypt plan create \
  --name "Single Site 1 Tahun" \
  --slug single_site_1y \
  --limit 1 \
  --duration 1y \
  --price 500000
```

**Opsi flag:**

| Flag | Keterangan | Nilai |
|---|---|---|
| `--name` | Nama tampilan plan | String (wajib) |
| `--slug` | Identifier unik | Huruf kecil + underscore (wajib) |
| `--limit` | Batas jumlah domain | `1` = single, `5` = multi, `-1` = unlimited |
| `--duration` | Durasi lisensi | `1y`, `3y`, `lifetime` |
| `--price` | Harga dalam Rupiah | Angka desimal |

**Contoh semua variasi plan:**

```bash
# Single Site 1 Tahun
./zycrypt plan create --name "Single Site 1 Tahun" --slug single_site_1y --limit 1 --duration 1y --price 500000

# Single Site 3 Tahun
./zycrypt plan create --name "Single Site 3 Tahun" --slug single_site_3y --limit 1 --duration 3y --price 1200000

# Multi Site 1 Tahun (5 domain)
./zycrypt plan create --name "Multi Site 1 Tahun" --slug multi_site_1y --limit 5 --duration 1y --price 1500000

# Lifetime Unlimited
./zycrypt plan create --name "Lifetime Unlimited" --slug lifetime_unlimited --limit -1 --duration lifetime --price 5000000
```

### Lihat semua plan

```bash
./zycrypt plan list
```

```
SLUG            NAME                 LIMIT  DURATION  PRICE      ACTIVE
single_site_1y  Single Site 1 Tahun  1      1y        500000.00  yes
single_site_3y  Single Site 3 Tahun  1      3y        1200000.00 yes
multi_site_1y   Multi Site 1 Tahun   5      1y        1500000.00 yes
```

### Update plan

```bash
./zycrypt plan update single_site_1y --price 600000
./zycrypt plan update multi_site_1y --name "Multi Site Pro" --limit 10
```

### Nonaktifkan plan

```bash
./zycrypt plan disable single_site_1y
```

> Lisensi yang sudah ada tidak terpengaruh. Plan dinonaktifkan hanya mencegah pembuatan lisensi baru.

---

## 6. Registrasi Lisensi

Ada dua cara membuat lisensi: **interaktif** (direkomendasikan) atau **via flag langsung**.

### Cara 1 — Interaktif (Register Wizard)

Panduan langkah demi langkah, cocok untuk digunakan sehari-hari:

```bash
./zycrypt register
```

```
  ╔══════════════════════════════════════╗
  ║    ZyCrypt — Registrasi Lisensi      ║
  ╚══════════════════════════════════════╝

  ── Step 1: Masukkan nama domain ──
   Format: tanpa http:// atau https://
   Contoh: mtq.abdya.go.id
   Domain: portal.kotabekasi.go.id

  ── Step 2: Pilih Paket ──
   1. Single Site  — 1 domain
   2. Multi Site   — hingga 5 domain
   Pilihan [1/2]: 1

  ── Step 3: Pilih Durasi ──
   1. 1 Tahun
   2. 3 Tahun
   3. Lifetime (tanpa batas)
   Pilihan [1/2/3]: 1

  ── Step 4: Identifikasi Pelanggan ──
   Nama pelanggan: Dinas Kominfo Kota Bekasi
   Email pelanggan: it@kotabekasi.go.id

  Memproses...

  ──────────────────────────────────────
  ✓  Registrasi Berhasil!
  ──────────────────────────────────────
  Domain   : portal.kotabekasi.go.id
  Key      : ZYC-S1GH-MSN9-2XBM-BQBD
  Paket    : Single Site
  Durasi   : 1 Tahun
  Expired  : 19 May 2027
  Status   : active
  Dibuat   : 19 May 2026, 23:33
  ──────────────────────────────────────

  Simpan key di file .env project client:
  ZYCRYPT_LICENSE_KEY=ZYC-S1GH-MSN9-2XBM-BQBD
  ZYCRYPT_SERVER_URL=https://zycrypt.yourdomain.com
```

> Domain otomatis terdaftar langsung. Plan yang cocok dicari otomatis — jika belum ada, dibuat baru.

### Cara 2 — Via Flag

```bash
./zycrypt license create \
  --name "Pemerintah Aceh Barat Daya" \
  --email "it@abdya.go.id" \
  --plan single_site_1y
```

```
✓ License created successfully
  Key     : ZYC-S1AB-2C3D-4E5F-6G7H
  Plan    : single_site_1y
  Expired : 2027-05-19
  Status  : active
```

---

## 7. Manajemen Lisensi

### Lihat semua lisensi

```bash
./zycrypt license list

# Filter berdasarkan status
./zycrypt license list --status active
./zycrypt license list --status banned
./zycrypt license list --status expired
```

```
KEY                      CUSTOMER                    PLAN            STATUS  EXPIRES
ZYC-S1GH-MSN9-2XBM-BQBD  Dinas Kominfo Kota Bekasi  single_site_1y  active  2027-05-19
ZYC-S147-DUVH-349E-2HRL  Pemerintah Aceh Barat Daya  single_site_1y  active  2027-05-19
```

### Detail lisensi

```bash
./zycrypt license show ZYC-S1GH-MSN9-2XBM-BQBD
```

```
Key           : ZYC-S1GH-MSN9-2XBM-BQBD
Customer      : Dinas Kominfo Kota Bekasi <it@kotabekasi.go.id>
Plan          : single_site_1y (limit: 1)
Status        : active
Purchase date : 2026-05-19
Expires       : 2027-05-19

Domains:
  - portal.kotabekasi.go.id [primary] (since 2026-05-19)

Recent events:
  [2026-05-19 23:33:20] license_created
  [2026-05-19 23:45:01] validate_success
```

### Ban lisensi

```bash
./zycrypt license ban ZYC-S1GH-MSN9-2XBM-BQBD --reason "Pelanggaran ketentuan penggunaan"
```

> Validasi berikutnya (maks. 10 menit) akan langsung ditolak dengan kode `license_banned`. Grace period tidak berlaku untuk status banned.

### Unban lisensi

```bash
./zycrypt license unban ZYC-S1GH-MSN9-2XBM-BQBD
```

### Perpanjang masa aktif

```bash
# Perpanjang 1 tahun (default)
./zycrypt license extend ZYC-S1GH-MSN9-2XBM-BQBD

# Perpanjang 3 tahun
./zycrypt license extend ZYC-S1GH-MSN9-2XBM-BQBD --years 3
```

> Jika lisensi sudah expired, perpanjangan dihitung dari hari ini. Jika masih aktif, ditambah dari tanggal expired saat ini.

### Hapus lisensi permanen

```bash
./zycrypt license revoke ZYC-S1GH-MSN9-2XBM-BQBD
```

CLI akan meminta konfirmasi dengan mengetik ulang key sebelum data dihapus permanen beserta semua domain dan log terkait.

---

## 8. Manajemen Domain

### Tambah domain ke lisensi

```bash
./zycrypt domain add ZYC-S1GH-MSN9-2XBM-BQBD --domain mtq.kotabekasi.go.id
```

> Untuk paket Single Site, hanya boleh 1 domain. Untuk Multi Site, sesuai batas plan.

### Hapus domain dari lisensi

```bash
./zycrypt domain remove ZYC-S1GH-MSN9-2XBM-BQBD --domain mtq.kotabekasi.go.id
```

Gunakan ini ketika client pindah server atau ganti domain.

### Lihat daftar domain

```bash
./zycrypt domain list ZYC-S1GH-MSN9-2XBM-BQBD
```

```
DOMAIN                      PRIMARY  REGISTERED
portal.kotabekasi.go.id     yes      2026-05-19
mtq.kotabekasi.go.id        no       2026-05-20
```

---

## 9. Utilitas Key

### Generate key (preview tanpa simpan ke DB)

```bash
./zycrypt key generate --plan single_site_1y --years 1
./zycrypt key generate --plan lifetime_unlimited --years 0
```

```
Preview key (not saved to DB):
  Key  : ZYC-S1AB-2C3D-4E5F-6G7H
  Type : S
  Years: 1 (0=lifetime)
```

### Verifikasi key

```bash
./zycrypt key verify ZYC-S1AB-2C3D-4E5F-6G7H
```

```
✓ Key is valid
  Type     : S
  Duration : 1 years (0=lifetime)
```

---

## 10. Menjalankan Server API

```bash
# Jalankan di port default (8743)
./zycrypt serve

# Override port
./zycrypt serve --port 9000
```

```
ZyCrypt server starting on 0.0.0.0:8743
```

### Setup sebagai systemd service (Linux/VPS)

Buat file `/etc/systemd/system/zycrypt.service`:

```ini
[Unit]
Description=ZyCrypt License Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/zycrypt
ExecStart=/usr/local/bin/zycrypt serve
Restart=always
RestartSec=5
Environment=ZYCRYPT_DATABASE_DSN=postgres://user:password@localhost:5432/zycrypt

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable zycrypt
systemctl start zycrypt
systemctl status zycrypt
```

### Setup Nginx sebagai reverse proxy

```nginx
server {
    listen 443 ssl;
    server_name zycrypt.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8743;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

---

## 11. API Endpoint

### GET `/api/v1/ping`

Cek status server.

```bash
curl https://zycrypt.yourdomain.com/api/v1/ping
```

```json
{
  "status": "ok",
  "version": "1.0.0",
  "ts": 1747612800
}
```

### POST `/api/v1/validate`

Validasi lisensi dari package client. Dipanggil otomatis oleh `zycrypt-vue` dan `zycrypt-laravel`.

**Request:**
```json
{
  "license_key": "ZYC-S1AB-2C3D-4E5F-6G7H",
  "domain":      "portal.kotabekasi.go.id",
  "pkg_version": "1.0.0",
  "timestamp":   1747612800,
  "signature":   "hmac_sha256_hex"
}
```

Signature dibuat dengan: `HMAC-SHA256(shared_secret, license_key + ":" + domain + ":" + timestamp)`

**Response sukses (200):**
```json
{
  "data": "base64_aes256_encrypted_payload"
}
```

Payload setelah didekripsi:
```json
{
  "valid": true,
  "plan": "single_site_1y",
  "site_limit": 1,
  "expires_at": "2027-05-19T00:00:00Z",
  "is_lifetime": false,
  "ts": 1747612800
}
```

**Response gagal (403):**
```json
{
  "valid": false,
  "reason": "license_banned",
  "detail": "Pelanggaran ketentuan penggunaan"
}
```

**Kode reason yang mungkin:**

| Kode | Penyebab |
|---|---|
| `license_not_found` | Key tidak ditemukan di database |
| `license_banned` | Lisensi dibanned oleh admin |
| `license_inactive` | Lisensi dinonaktifkan |
| `license_expired` | Masa aktif habis |
| `domain_mismatch` | Domain tidak terdaftar & limit penuh |
| `invalid_signature` | HMAC signature tidak valid |
| `token_expired` | Timestamp request lebih dari 10 menit |

### Grace Period

Jika server tidak dapat dijangkau, package client mengizinkan aplikasi tetap berjalan selama **24 jam** sejak validasi terakhir yang berhasil. Grace period **tidak berlaku** untuk status `banned` dan `expired`.

---

## 12. Referensi Lengkap CLI

```
zycrypt
├── register                           Registrasi lisensi interaktif (wizard)
│
├── serve                              Jalankan HTTP license server
│   ├── --port   int                   Override port
│   └── --config string                Path ke config file
│
├── db
│   └── migrate                        Jalankan migrasi database
│
├── plan
│   ├── create                         Buat plan baru
│   │   ├── --name     string*         Nama tampilan
│   │   ├── --slug     string*         Slug unik
│   │   ├── --limit    int             1 | 5 | -1
│   │   ├── --duration string          1y | 3y | lifetime
│   │   └── --price    float           Harga Rupiah
│   ├── list                           Tampilkan semua plan
│   ├── update [slug]                  Update atribut plan
│   │   ├── --name     string
│   │   ├── --price    float
│   │   └── --limit    int
│   └── disable [slug]                 Nonaktifkan plan
│
├── license
│   ├── create                         Buat lisensi baru
│   │   ├── --name     string*         Nama pelanggan
│   │   ├── --email    string*         Email pelanggan
│   │   └── --plan     string*         Slug plan
│   ├── list                           Tampilkan semua lisensi
│   │   └── --status   string          active|inactive|banned|expired
│   ├── show [key]                     Detail lisensi + domain + 10 log terakhir
│   ├── ban [key]                      Set status banned
│   │   └── --reason   string*         Alasan ban (wajib)
│   ├── unban [key]                    Restore ke status active
│   ├── extend [key]                   Perpanjang masa aktif
│   │   └── --years    int             Jumlah tahun (default: 1)
│   └── revoke [key]                   Hapus permanen (dengan konfirmasi)
│
├── domain
│   ├── add [key]                      Tambah domain ke lisensi
│   │   └── --domain   string*         Hostname tanpa https://
│   ├── remove [key]                   Hapus domain dari lisensi
│   │   └── --domain   string*         Hostname yang dihapus
│   └── list [key]                     Tampilkan domain terdaftar
│
└── key
    ├── generate                       Preview key tanpa simpan ke DB
    │   ├── --plan     string          Slug plan
    │   └── --years    int             0 = lifetime
    └── verify [key]                   Cek format + checksum
```

---

*ZyCrypt — Ziya Encryptor. Satu binary. Kendali penuh. Karya terlindungi.*
