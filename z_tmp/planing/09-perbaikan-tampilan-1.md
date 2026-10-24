# 9. Perbaikan Tampilan: Pemilih Waktu dan Interaksi Grafik

## 9.1 Tujuan

Meningkatkan pengalaman pemantauan tanpa mengubah alur perangkat IoT atau deployment staging:

- Mengganti input native `datetime-local` pada halaman Histori dengan pemilih tanggal/waktu modern.
- Menambahkan interaksi grafik seperti Grafana: zoom ke rentang tertentu, pan, dan reset zoom.
- Mengganti teks statis `15 menit terakhir` pada Dashboard dengan preset rentang `1 menit`, `10 menit`, `30 menit`, dan `1 jam`.
- Memastikan Dashboard tetap menerima pembacaan realtime melalui WebSocket setelah pengguna memilih rentang waktu.

## 9.2 Kondisi Saat Ini

| Area | Implementasi saat ini | Dampak |
|---|---|---|
| Date picker | `input type="datetime-local"` di `frontend/src/views/History.vue` | Tampilan bergantung browser dan kurang konsisten/profesional. |
| Chart | `frontend/src/AppChart.vue` memakai Chart.js dengan label kategori/string | Tooltip dan grafik sudah baik, tetapi belum mendukung zoom, pan, atau reset zoom. |
| Data Dashboard | `frontend/src/live.js` hanya mengambil raw data 15 menit ketika aplikasi dimulai dan menyimpan maksimum 450 titik | Tidak bisa menampilkan rentang 1 jam dan teks rentang bersifat statis. |
| Realtime | Pembacaan baru masuk melalui WebSocket dan ditambahkan ke `live.points` | Harus dipertahankan pada seluruh pilihan timeframe. |

## 9.3 Pilihan Library

| Kebutuhan | Library | Alasan |
|---|---|---|
| Date/time range picker | `@vuepic/vue-datepicker` | Komponen Vue 3, mendukung range, waktu, format, locale, dark mode, dan CSS bawaan; cocok menggantikan dua input native. |
| Zoom dan pan Chart.js | `chartjs-plugin-zoom` | Plugin resmi ekosistem Chart.js; mendukung drag-to-zoom, scroll wheel/pinch zoom, pan, dan reset tanpa mengganti komponen chart yang sudah ada. |

Kedua dependency ditambahkan ke `frontend/package.json` dan `frontend/package-lock.json` menggunakan `npm install`. Tidak ada perubahan dependency backend, database, firmware, atau MQTT.

## 9.4 Desain Fungsional

### A. Halaman Histori

File utama: `frontend/src/views/History.vue`.

1. Ganti dua kontrol `Dari` dan `Sampai` dengan satu `VueDatePicker` mode `range` dan `enable-time-picker`.
2. Gunakan locale `id-ID`, format tampilan lokal yang jelas, tombol apply/cancel, serta label aksesibel.
3. State komponen menyimpan dua objek `Date` dalam range. Saat memanggil API, masing-masing dikonversi ke RFC3339 menggunakan `toISOString()` sebagaimana kontrak backend saat ini.
4. Tetap pertahankan pilihan resolution, metric, tombol `Tampilkan`, pesan error, dan kondisi tanpa data.
5. Tambahkan validasi UI: range wajib lengkap dan waktu awal harus sebelum waktu akhir; jangan kirim request bila tidak valid.

### B. Interaksi Zoom Grafik

File utama: `frontend/src/AppChart.vue`.

1. Register `zoomPlugin` secara eksplisit bersama Chart.js.
2. Aktifkan zoom drag pada sumbu X dengan visual selection box, agar pengguna dapat memilih area waktu pada grafik dengan mouse.
3. Aktifkan wheel zoom dan pinch zoom hanya pada sumbu X; tooltip dan mode interaksi indeks tetap dipertahankan.
4. Aktifkan pan horizontal setelah chart dizoom.
5. Tambahkan kontrol kecil di bawah/atas chart: `Reset zoom`. Kontrol hanya tampil atau aktif ketika chart berada pada state zoom.
6. Gunakan rentang label kategori Chart.js yang ada agar perubahan tetap minimal. Nilai dan urutan data tidak diubah.
7. Pastikan instance Chart dihancurkan ketika component unmount dan state zoom tidak bocor saat ganti halaman/chart.

Catatan: zoom chart hanya mengubah viewport data yang sudah dimuat. Ia tidak mengunduh data tambahan secara otomatis. Pengambilan cakupan data dilakukan oleh preset Dashboard atau filter Histori.

### C. Preset Timeframe Dashboard

File utama: `frontend/src/views/Dashboard.vue` dan `frontend/src/live.js`.

1. Tambahkan dropdown dengan opsi berikut:

| Nilai | Label |
|---:|---|
| 1 menit | `1 menit terakhir` |
| 10 menit | `10 menit terakhir` |
| 30 menit | `30 menit terakhir` |
| 1 jam | `1 jam terakhir` |

2. Gunakan default `10 menit` agar jumlah titik awal sekitar 300 pada interval perangkat 2 detik, tetap ringan sekaligus cukup informatif. Pilihan terakhir disimpan selama sesi halaman aktif; persistence browser tidak diperlukan pada tahap ini.
3. Ketika preset berubah, Dashboard meminta ulang `GET /api/readings/history?resolution=raw&from=<RFC3339>&to=<RFC3339>` untuk rentang tersebut, lalu mengganti points chart dengan hasil API.
4. WebSocket tetap menambahkan pembacaan terbaru ke kumpulan points. Setelah setiap pembacaan baru, buang data yang timestamp-nya lebih tua daripada rentang preset aktif. Dengan demikian chart bergerak secara realtime dan tidak melebihi timeframe yang dipilih.
5. Untuk mencegah hasil request lama menimpa pilihan terbaru saat pengguna cepat mengganti dropdown, gunakan `AbortController` atau sequence request di `live.js`.
6. Ubah `MAX_POINTS` yang saat ini 450 agar mengikuti pilihan terbesar: 1 jam pada interval 2 detik membutuhkan sekitar 1.800 titik. Terapkan batas aman 1.800 atau hitung berdasarkan timeframe aktif. Ini hanya array browser; API/database tetap menjadi sumber data.
7. Ganti teks statis `15 menit terakhir` dengan dropdown dan label yang mengikuti opsi aktif.
8. Tambahkan state loading ringkas di area grafik saat data rentang baru sedang diminta. Data terakhir tetap tampil hingga respons baru berhasil, sehingga chart tidak berkedip kosong.

## 9.5 Perubahan API dan Data

Tidak diperlukan endpoint atau migrasi database baru.

Endpoint `GET /api/readings/history` sudah menerima `from`, `to`, dan `resolution=raw`, serta memilih data raw untuk rentang hingga dua hari. Semua preset Dashboard maksimum satu jam, sehingga query raw yang berisi data tiap dua detik tetap sesuai dan ringan.

## 9.6 Pengujian

### Data pengujian

Gunakan database lokal, bukan server staging dan bukan dummy data, sebagai sumber utama pengujian.

Hasil pemeriksaan lokal saat rencana ini dibuat:

- Tersedia `6.584` pembacaan pada `sensor_readings`.
- Rentang data dari `2026-10-23 13:43:17 UTC` sampai `2026-10-23 17:26:10 UTC` (sekitar 3 jam 43 menit).
- Ada `5.999` pembacaan dengan `sensor_ok = true`.

Data ini cukup untuk memverifikasi preset 1 menit, 10 menit, 30 menit, dan 1 jam, histori range picker, serta zoom/pan. Tidak perlu menulis data dummy ke database dan tidak perlu mengarahkan perangkat IoT dari staging ke lokal.

### Skenario verifikasi

1. Jalankan `npm run build` untuk memverifikasi build frontend dan import CSS library date picker.
2. Jalankan backend dan frontend lokal atau rebuild Docker lokal, lalu buka Dashboard.
3. Uji semua preset Dashboard. Pastikan request API mengirim `from`/`to` yang benar dan jumlah titik meningkat sesuai rentang.
4. Pastikan pembacaan WebSocket baru muncul tanpa refresh, serta data yang lebih tua dari rentang pilihan tidak tetap ditampilkan.
5. Pada Dashboard dan Histori, drag pada chart untuk zoom, pan horizontal setelah zoom, coba wheel/pinch, lalu tekan `Reset zoom` dan pastikan rentang penuh kembali.
6. Pada Histori, pilih range waktu lengkap, ubah resolution dan metric, kemudian pastikan query/API dan label grafik benar.
7. Uji range tidak lengkap atau terbalik untuk memastikan tidak ada request invalid.
8. Uji layout desktop dan mobile: dropdown tidak menabrak tombol metrik, date picker tidak terpotong, dan canvas tetap responsif.
9. Setelah lolos lokal, build image dan deploy ke staging sesuai prosedur Docker yang sudah digunakan. Verifikasi akhir dilakukan di `http://192.168.1.123:1077/#dashboard` tanpa mengubah konfigurasi perangkat IoT.

## 9.7 File yang Akan Diubah

| File | Perubahan |
|---|---|
| `frontend/package.json` | Tambah `@vuepic/vue-datepicker` dan `chartjs-plugin-zoom`. |
| `frontend/package-lock.json` | Diperbarui oleh npm sesuai dependency baru. |
| `frontend/src/AppChart.vue` | Register plugin, konfigurasi zoom/pan, dan kontrol reset zoom. |
| `frontend/src/live.js` | Fungsi pemuatan points berdasarkan timeframe, batas points dinamis, dan perlindungan request race. |
| `frontend/src/views/Dashboard.vue` | Dropdown timeframe, loading state, dan integrasi reload chart. |
| `frontend/src/views/History.vue` | Ganti input native menjadi Vue date/time range picker beserta validasinya. |
| `frontend/src/App.vue` | Penyesuaian kecil style global jika dibutuhkan oleh kontrol baru. |

## 9.8 Di Luar Cakupan

- Tidak mengganti Chart.js dengan library chart baru.
- Tidak mengubah struktur tabel, retensi data, firmware ESP32, MQTT, atau backend API.
- Tidak menambah autentikasi, multi-device selector, export data, atau penyimpanan preferensi timeframe permanen.
