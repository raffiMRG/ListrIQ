# 8. Risiko & Catatan Tambahan

## 8.1 Jarak RF24L01 (50-100m)

- Modul yang dipakai sudah varian PA/LNA (antena eksternal), namun jarak 50-100m tetap merupakan batas atas kemampuan NRF24L01 bahkan dengan PA/LNA, terutama jika ada penghalang (tembok, pohon, dll).
- **Mitigasi:** Lakukan uji coba RF24 di lokasi sebelum instalasi permanen (lihat bagian 3.4). Jika packet loss tinggi:
  - Turunkan data rate RF24 ke 250kbps (lebih tahan jarak dibanding default 1-2Mbps)
  - Pastikan line-of-sight sebaik mungkin antara Sensor Node dan Gateway Node
  - Sebagai fallback, modul ESP32-CAM dan 1 NRF24L01 cadangan yang Anda miliki bisa dipakai sebagai **repeater/relay node** di tengah jalur jika jarak langsung tidak reliable

## 8.2 Deteksi Offline Dua Lapis

- MQTT LWT hanya mendeteksi jika koneksi **Gateway Node ke broker** (WiFi) terputus
- Jika link RF24 antara Sensor Node dan Gateway Node yang terputus (misal baterai/power Sensor Node mati), Gateway Node tetap terkoneksi MQTT normal sehingga LWT tidak trigger
- **Mitigasi:** Backend memiliki watchdog tambahan berbasis `last_seen` pada tabel `device_status` (lihat 4.4) sebagai lapisan kedua deteksi offline

## 8.3 Keselamatan Kelistrikan

- Pemasangan PZEM-004T melibatkan koneksi langsung ke jalur AC 220V
- Selalu matikan MCB/pemutus utama sebelum memasang/melepas kabel pada sisi AC PZEM-004T
- Pastikan casing/enclosure yang memadai untuk melindungi modul dari kontak langsung dan kondisi lingkungan (panel listrik biasanya lembap/berdebu)

## 8.4 Energy Register Bersifat Akumulatif

- Register energy pada PZEM-004T adalah akumulatif (terus bertambah, bukan reset otomatis)
- Perhitungan konsumsi per periode (jam/hari) di continuous aggregate menggunakan `max(energy) - min(energy)` dalam bucket waktu tersebut (lihat 5.2), bukan `sum()`
- Jika PZEM di-reset manual (via command reset energy) atau device restart dengan kondisi tertentu, nilai energy bisa turun ke 0 — perlu penanganan edge case ini di level backend (deteksi counter reset) agar perhitungan delta tidak menghasilkan nilai negatif

## 8.5 Pengembangan Lanjutan (di luar scope saat ini, dicatat untuk referensi)

- Notifikasi aktif (Telegram/Email) bisa ditambahkan nanti tanpa mengubah arsitektur inti (cukup tambah consumer baru di backend yang subscribe ke event status/alarm)
- Multi-titik pengukuran bisa ditambahkan dengan menambah Sensor Node baru (device_id berbeda) tanpa mengubah skema database (sudah mendukung `device_id`)
- Estimasi biaya listrik (Rupiah) bisa ditambahkan ke halaman laporan dengan mengalikan `energy_delta` dengan tarif per kWh yang dikonfigurasi
</content>
