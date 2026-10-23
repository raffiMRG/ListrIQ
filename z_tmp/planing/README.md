# Planning: Sistem Monitoring Listrik Rumah

Dokumen ini berisi rencana lengkap pembangunan sistem monitoring listrik, mulai dari wiring hardware hingga aplikasi yang berjalan di atas Docker.

## Daftar Isi

1. [Ringkasan & Arsitektur](./01-arsitektur.md)
2. [Wiring Hardware](./02-wiring.md)
3. [Firmware (ESP32)](./03-firmware.md)
4. [Backend (Golang)](./04-backend.md)
5. [Database (PostgreSQL + TimescaleDB)](./05-database.md)
6. [Frontend (Vue)](./06-frontend.md)
7. [Deployment (Docker Compose)](./07-deployment.md)
8. [Risiko & Catatan Tambahan](./08-risiko.md)

## Ringkasan Keputusan

| Area | Keputusan |
|---|---|
| Topologi | 1 titik ukur: Sensor Node (ESP32+PZEM+NRF24L01+PA/LNA) -> Gateway Node (ESP32+NRF24L01+PA/LNA) -> WiFi -> Server |
| Build tool firmware | PlatformIO (bukan Arduino IDE), 2 environment terpisah (sensor-node, gateway-node) |
| Protokol device->server | MQTT (Mosquitto broker), LWT untuk status online/offline |
| Realtime frontend | WebSocket dari backend Go ke Vue |
| Parameter data | Voltage, Current, Power, Energy, Frequency, Power Factor, Alarm |
| Interval sampling | 2-5 detik |
| Database | PostgreSQL + TimescaleDB extension (hypertable + continuous aggregate) |
| Notifikasi | Cukup indikator visual di dashboard (tidak ada Telegram/Email) |
| Fitur tambahan | Riwayat & laporan historis (harian/bulanan) |
| Deployment | Docker Compose, home server, akses via IP LAN (tanpa domain publik/SSL) |
| Modul NRF24L01 | Varian PA+LNA (antena eksternal) di kedua sisi |
</content>
