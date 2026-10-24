# 10. Login Admin, Settings, Backup, dan Restore

## Tujuan

Menambahkan area Pengaturan yang hanya dapat diakses admin untuk mengunduh backup database, memulihkan backup, dan melihat penggunaan penyimpanan. Dashboard, Histori, dan Laporan tidak berubah dan tetap dapat dibuka tanpa login.

## Kondisi Saat Ini

- Backend Go belum mempunyai autentikasi, session, atau role.
- Frontend Vue memakai navigasi hash pada `frontend/src/App.vue`.
- TimescaleDB PostgreSQL 16 telah menyediakan `pg_dump` dan `pg_restore` di container `timescaledb`.
- Container aplikasi memakai image distroless; tidak dapat menjalankan shell, `pg_dump`, atau `pg_restore`.
- Log aplikasi hanya berada di stdout Docker dan belum memiliki volume log persistent.
- Saat pemeriksaan, database lokal berukuran sekitar 11 MB.

## Keputusan yang Sudah Disetujui

| Area | Keputusan | Dampak |
|---|---|---|
| Cakupan login | Dashboard, Histori, dan Laporan publik; hanya Pengaturan memerlukan login. | Middleware auth hanya melindungi `/api/admin/*` dan operasi sesi admin. |
| HTTPS | Belum dipasang; admin hanya memakai LAN tepercaya. | Login HTTP tetap dapat dipakai pada LAN privat, tetapi password/cookie tidak terenkripsi. HTTPS wajib ditambahkan sebelum akses dari jaringan yang tidak tepercaya atau internet. |
| Restore | Restore penuh disetujui. | Seluruh database monitoring diganti, termasuk schema, raw data, aggregate, status, indeks, dan policy. Ada konfirmasi dua tahap serta pre-restore backup otomatis. Data MQTT yang masuk saat maintenance tidak disimpan. |
| Retensi pre-restore backup | Tidak dihapus otomatis. | Settings menampilkan daftar dan ukuran backup agar admin dapat menghapusnya manual ketika disk mulai penuh. |
| Disk usage log | Tahap awal hanya database dan backup; log Docker dikelola host. | Tidak menambah persistent application log sekarang. UI menjelaskan bahwa ukuran Docker log diperiksa dari host. |
| Akun admin | Satu akun dari environment variable. | Tidak ada registrasi, tabel users, role, atau manajemen user pada versi pertama. |

## Autentikasi Admin

### Konfigurasi

Tambahkan variabel berikut ke `deploy/.env`, dengan contoh placeholder saja di `deploy/.env.example`:

```dotenv
ADMIN_USERNAME=admin
ADMIN_PASSWORD_HASH=password123
ADMIN_SESSION_TTL=8h
ADMIN_RESTORE_MAX_BYTES=2147483648
```

- `ADMIN_PASSWORD_HASH` berisi password mentah sesuai kebutuhan deployment. Backend langsung mengubahnya menjadi hash Argon2id di memori saat startup; nilai mentah tidak disimpan di database atau dikirim ke browser.
- Karena `.env` tetap berisi password plaintext, file ini harus diabaikan Git dan hanya dapat dibaca admin server.
- File `.env` tetap diabaikan Git.

### Sesi dan proteksi

- `POST /api/auth/login` menerima username/password dan menerapkan rate limit per IP.
- Jika valid, backend membuat session acak in-memory dengan batas 8 jam.
- Session disimpan browser melalui cookie `HttpOnly`, `SameSite=Strict`, `Path=/`.
- Atribut `Secure` diaktifkan nanti setelah HTTPS tersedia. Cookie `Secure` tidak akan berfungsi pada HTTP LAN saat ini.
- `POST /api/auth/logout` menghapus sesi dan cookie.
- `GET /api/auth/me` memberi status login untuk frontend tanpa mengirim data sensitif.
- Semua endpoint admin memeriksa session dan Origin/CSRF untuk request mutasi.
- Restart container aplikasi menghapus semua sesi, yang aman untuk satu akun admin.

## Backup Database

### Format dan proses

- Gunakan `pg_dump --format=custom` agar hasil `.dump` memuat schema TimescaleDB, hypertable, continuous aggregate, policy, indeks, dan data.
- Beri nama download `powermon-YYYYMMDD-HHMMSS.dump`.
- Backup dijalankan sebagai job asynchronous agar browser tidak timeout ketika database tumbuh besar.
- Setelah job selesai, browser mengunduh file melalui endpoint admin yang memeriksa session.
- File download dihapus setelah TTL singkat atau setelah selesai diunduh.

### Endpoint yang direncanakan

| Endpoint | Fungsi |
|---|---|
| `POST /api/admin/backup` | Membuat job backup. |
| `GET /api/admin/jobs/{id}` | Membaca status/progress backup atau restore. |
| `GET /api/admin/backup/{id}/download` | Mengunduh hasil backup yang selesai dibuat. |

## Restore Database

### Aturan keamanan

- Hanya menerima custom dump PostgreSQL `.dump` yang dihasilkan fitur Backup.
- Ukuran upload dibatasi `ADMIN_RESTORE_MAX_BYTES`.
- Upload disimpan sementara dalam volume, bukan seluruhnya di RAM.
- Validasi file memakai `pg_restore --list`; ekstensi nama file saja tidak dipercaya.
- UI menampilkan nama/ukuran file, warning dampak restore, checkbox persetujuan, dan konfirmasi kedua dengan mengetik `RESTORE`.

### Alur restore

1. Admin mengunggah `.dump`.
2. Backend membuat job dan worker memvalidasi dump.
3. Admin melihat hasil validasi dan menekan konfirmasi kedua.
4. Aplikasi masuk maintenance mode; `onData` MQTT berhenti menulis pembacaan baru ke database sementara proses berlangsung.
5. Worker membuat pre-restore backup otomatis ke volume backup persistent.
6. Worker menjalankan `pg_restore --clean --if-exists --no-owner --no-privileges` ke database `powermon`.
7. Backend menjalankan health check: koneksi database, extension TimescaleDB, tabel utama, dan jumlah data.
8. Maintenance mode berakhir; hasil sukses/gagal dan lokasi pre-restore backup ditampilkan ke admin.

Jika restore gagal, pre-restore backup tetap disimpan untuk pemulihan manual. Karena backup pengaman tidak dihapus otomatis, admin harus memantau kapasitas disk lewat Settings.

## Arsitektur Worker Backup/Restore

Backend publik tidak boleh diberi akses Docker socket. Mount `/var/run/docker.sock` pada aplikasi web berisiko memberi root access ke host.

Tambahkan service helper `db-admin` pada `deploy/docker-compose.yml`:

- Service internal, tanpa port publik.
- Memakai image yang memiliki PostgreSQL CLI (`pg_dump`, `pg_restore`).
- Berbagi network internal dengan `timescaledb`.
- Memakai volume `admin_tmp` untuk upload/job sementara dan `admin_backups` untuk pre-restore backup.
- Menerima perintah backup/restore dari backend melalui kanal internal berautentikasi shared secret atau Unix socket.
- Tidak menerima akses langsung dari browser.

Alternatif mengubah app distroless agar memuat PostgreSQL CLI tidak dipilih karena memperbesar hak akses aplikasi publik.

## Penggunaan Disk

### Yang ditampilkan pada Settings

- Total ukuran database: `pg_database_size(current_database())`.
- Ukuran raw readings beserta indeks: `pg_total_relation_size('sensor_readings')`.
- Ukuran tabel/materialized view penting bila diperlukan.
- Total ukuran volume backup persistent dan jumlah file backup.
- Ukuran pekerjaan/upload sementara bila ada.

### Log

Pada versi ini UI menampilkan catatan: log aplikasi berasal dari stdout Docker dan dikelola server/host. Tidak ada ukuran log Docker yang dihitung backend karena file log berada di host dan tidak diekspos aman ke container app.

Peningkatan terpisah bila diperlukan nanti:

- Konfigurasi Docker log rotation di Compose.
- File log aplikasi persistent dengan rotasi ukuran dan retensi.

## Antarmuka Frontend

### Navigasi

- Tambah menu `Pengaturan` dengan ikon settings ke `frontend/src/App.vue`.
- Membuka `#settings` tanpa sesi menampilkan form login.
- Setelah login, halaman menampilkan tombol Logout.

### Card Pengaturan

1. **Penyimpanan**: ukuran database, `sensor_readings`, backup persistent, file sementara, waktu refresh, tombol Muat ulang.
2. **Backup Database**: deskripsi format `.dump`, tombol Buat dan Unduh Backup, status/progress/error job.
3. **Restore Database**: input upload `.dump`, metadata file, warning operasi destruktif, checkbox persetujuan, input konfirmasi `RESTORE`, dan status job.
4. **Backup Pengaman**: daftar nama, tanggal, dan ukuran pre-restore backup agar admin bisa memantau/bersih-bersih manual.

## File yang Akan Diubah

| File | Perubahan |
|---|---|
| `backend/main.go` | Konfigurasi admin, lifecycle worker, maintenance mode. |
| `backend/api.go` | Route publik, auth, dan admin. |
| `backend/auth.go` (baru) | Hash verification, session, cookie, CSRF/origin, rate limit. |
| `backend/admin.go` (baru) | Storage statistics, backup/restore jobs, audit event. |
| `backend/admin_worker.go` (baru) | Komunikasi internal dengan service `db-admin`. |
| `backend/*_test.go` | Test auth, endpoint terlindungi, job, dan validasi. |
| `frontend/src/App.vue` | Menu Settings dan state session. |
| `frontend/src/views/Settings.vue` (baru) | Login, storage, backup, restore, status job. |
| `deploy/docker-compose.yml` | Service `db-admin`, environment, volume temporary dan backup. |
| `deploy/.env.example` | Placeholder semua konfigurasi admin. |
| `Dockerfile`/Dockerfile worker | Build service helper yang membutuhkan PostgreSQL CLI. |
| `.gitignore` | Mengabaikan backup lokal bila directory host dipakai. |

## Pengujian

### Backend

1. Password hash benar/salah, session expiry, logout, endpoint admin tanpa sesi (`401`).
2. Rate limit login dan Origin/CSRF request mutasi.
3. Query disk usage database dan backup.
4. Validasi file kosong, dump invalid/corrupt, dan dump valid.
5. Job backup/restore serta pembersihan upload temporary ketika gagal/berhasil.

### End-to-end lokal

1. Login admin, refresh halaman, dan logout.
2. Buat backup dari database lokal, unduh, lalu validasi melalui `pg_restore --list`.
3. Catat jumlah rows, tambahkan fixture lokal, restore backup, dan pastikan jumlah rows kembali ke keadaan backup.
4. Kirim publish MQTT test saat restore untuk memastikan maintenance mode tidak mencampur data baru.
5. Pastikan pre-restore backup muncul di Settings dan tidak dihapus otomatis.
6. Uji upload limit, file salah, pembatalan, serta browser desktop/mobile.

### Staging

1. Deploy tanpa restore terlebih dahulu.
2. Konfigurasikan akun admin pada `.env` staging.
3. Login dari LAN tepercaya, lakukan backup, unduh file, dan validasi dump offline.
4. Jadwalkan maintenance window sebelum restore pertama. Jangan menimpa staging memakai backup lokal tanpa persetujuan eksplisit.

## Catatan Risiko

- HTTP LAN tidak mengenkripsi password maupun cookie; akses admin dibatasi LAN tepercaya sampai HTTPS dipasang.
- Backup pre-restore tidak dihapus otomatis, sehingga disk dapat penuh bila restore sering dilakukan. Disk usage dan daftar backup harus dipantau.
- Pembacaan IoT selama maintenance restore sengaja dibuang agar database hasil restore konsisten.
- Restore penuh lintas versi schema perlu diuji dengan backup nyata sebelum dipakai pada staging.
