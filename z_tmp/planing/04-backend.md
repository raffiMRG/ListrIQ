# 4. Backend (Golang)

## 4.1 Tanggung Jawab

- Subscribe ke MQTT broker (topic data & status)
- Simpan data sensor ke TimescaleDB
- Kelola status device (online/offline) termasuk watchdog tambahan di level aplikasi
- Sajikan WebSocket server untuk push data realtime ke frontend Vue
- Sajikan REST API untuk histori data, laporan, dan status

## 4.2 Library yang disarankan

| Kebutuhan | Library |
|---|---|
| MQTT client | `eclipse/paho.mqtt.golang` |
| WebSocket | `gorilla/websocket` |
| HTTP router | `gin-gonic/gin` atau `go-chi/chi` (sesuaikan preferensi) |
| Database driver | `jackc/pgx` (driver Postgres performa baik, cocok dipakai bareng TimescaleDB) |
| Query builder/ORM (opsional) | `sqlc` (generate kode dari SQL) atau raw SQL dengan `pgx` langsung |

## 4.3 Modul Internal

```
backend/
  cmd/server/main.go          - entry point
  internal/
    mqtt/                     - MQTT client, handler subscribe topic data & status
    ws/                       - WebSocket hub (register client, broadcast)
    api/                      - HTTP handler REST (histori, laporan, status)
    storage/                  - layer akses TimescaleDB (insert reading, query aggregate)
    device/                   - logic status device (online/offline, watchdog last_seen)
    config/                   - load config (env var: MQTT broker addr, DB DSN, dll)
```

## 4.4 Alur Proses Utama

1. **MQTT handler** `powermon/main/data`:
   - Unmarshal JSON payload
   - Insert ke tabel `sensor_readings` (TimescaleDB hypertable)
   - Update `last_seen` di tabel `device_status`
   - Broadcast data ke semua WebSocket client yang terkoneksi

2. **MQTT handler** `powermon/main/status`:
   - Terima `online`/`offline` dari LWT/manual publish
   - Update kolom `status` di tabel `device_status`
   - Broadcast perubahan status ke WebSocket client

3. **Watchdog tambahan** (lapisan kedua selain MQTT LWT):
   - Goroutine berjalan tiap beberapa detik, cek `last_seen` vs waktu sekarang
   - Jika selisih > threshold (mis. 15-20 detik, lebih dari 3x interval sampling), tandai device sebagai `offline` meski status MQTT masih `online`
   - Berguna untuk kasus Gateway Node masih konek MQTT tapi RF24 link ke Sensor Node putus (LWT MQTT tidak akan trigger karena WiFi/MQTT gateway tetap hidup)

## 4.5 REST API Endpoint (contoh)

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/api/status` | Status device saat ini (online/offline, last_seen) |
| GET | `/api/readings/latest` | Data reading terakhir |
| GET | `/api/readings/history?from=&to=&resolution=raw\|hourly\|daily` | Data histori untuk grafik, pilih resolusi dari raw/aggregate |
| GET | `/api/reports/daily?date=` | Ringkasan konsumsi energi harian |
| GET | `/api/reports/monthly?month=` | Ringkasan konsumsi energi bulanan |
| GET | `/ws` | Upgrade koneksi ke WebSocket untuk data realtime |

## 4.6 WebSocket Message Format (contoh)

```json
{
  "type": "reading",
  "data": {
    "voltage": 221.3,
    "current": 2.15,
    "power": 475.6,
    "energy": 12.345,
    "frequency": 50.0,
    "power_factor": 0.98,
    "alarm": 0,
    "ts": "2026-10-22T10:00:00Z"
  }
}
```
```json
{
  "type": "status",
  "data": {
    "device_id": "main",
    "status": "offline",
    "last_seen": "2026-10-22T09:59:40Z"
  }
}
```
</content>
