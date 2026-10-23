# 1. Ringkasan & Arsitektur

## Hardware yang digunakan

- PZEM-004T v3 + CT eksternal (sensor tegangan/arus/daya)
- ESP32 DevKitC V4 (WROOM-32U, antena eksternal U.FL) x2 (Sensor Node + Gateway Node)
- NRF24L01+PA+LNA (antena eksternal) x2

## Diagram Arsitektur

```
[Panel Listrik Utama]                         [Dekat Router/Modem]
  ESP32 + PZEM-004T                              ESP32 (Gateway Node)
  + NRF24L01+PA/LNA (TX)   --RF 2.4GHz-->         + NRF24L01+PA/LNA (RX)
  "Sensor Node"             (50-100m)                     |
                                                           | WiFi
                                                           v
                                              +-------------------------+
                                              |   MQTT Broker           |
                                              |   (Mosquitto)           |
                                              +-------------------------+
                                                           |
                                              +-------------------------+
                                              |   Backend (Golang)      |
                                              |   - MQTT subscriber     |
                                              |   - WebSocket server    |
                                              |   - REST API            |
                                              +-------------------------+
                                                 |                   |
                                                 v                   v
                                   +----------------------+  +----------------+
                                   | PostgreSQL+TimescaleDB|  |  Frontend Vue  |
                                   | - sensor_readings      |  | - dashboard   |
                                   |   (hypertable)          |  | - grafik WS   |
                                   | - continuous aggregate  |  | - status      |
                                   | - device_status         |  | - histori     |
                                   +----------------------+  +----------------+

Semua service (mosquitto, timescaledb, backend, frontend) berjalan dalam
satu docker-compose.yml di home server, diakses via IP LAN.
```

## Alur Data

1. Sensor Node membaca data PZEM-004T tiap 2-5 detik via UART (Modbus-RTU)
2. Sensor Node mengirim data terbaca ke Gateway Node via NRF24L01 (RF 2.4GHz)
3. Gateway Node menerima data, mengemas jadi JSON, publish ke MQTT broker via WiFi
4. Backend Go subscribe topic data & status MQTT, simpan ke TimescaleDB, broadcast ke semua client Vue via WebSocket
5. Frontend Vue menampilkan data realtime (grafik, kartu status) dan histori (query REST API ke backend)

## Topik MQTT

| Topic | Payload | Keterangan |
|---|---|---|
| `powermon/main/data` | JSON: voltage, current, power, energy, frequency, power_factor, alarm, timestamp | Publish tiap kali gateway terima data baru dari sensor node |
| `powermon/main/status` | `online` / `offline` | Retained + LWT. `online` dipublish manual setelah connect, `offline` otomatis dipublish broker (LWT) saat gateway disconnect tidak normal |
</content>
