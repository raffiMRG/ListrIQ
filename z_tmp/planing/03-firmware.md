# 3. Firmware (ESP32)

Ditulis dengan Arduino framework, menggunakan **PlatformIO** (via VS Code extension atau CLI) sebagai build tool utama — bukan Arduino IDE. PlatformIO dipilih karena: dependency library dikunci per-project (`platformio.ini`), mudah dikelola dengan version control (cocok ditaruh dalam folder `firmware/` di repo yang sama), serta mendukung 2 target (Sensor Node & Gateway Node) sebagai environment terpisah dalam satu konfigurasi.

## 3.1 Struktur Project PlatformIO

```
firmware/
  sensor-node/
    platformio.ini
    src/
      main.cpp
    include/
      config.h          - konstanta pin, alamat RF24, interval sampling
  gateway-node/
    platformio.ini
    src/
      main.cpp
    include/
      config.h          - konstanta pin, WiFi SSID/password, MQTT broker, alamat RF24
```

Dua folder terpisah (bukan 1 project dengan `src_filter`) dipakai agar board config, library, dan source code untuk Sensor Node vs Gateway Node tetap jelas terpisah dan independen saat build/upload.

## 3.2 Contoh `platformio.ini` — Sensor Node

```ini
[env:esp32dev]
platform = espressif32
board = esp32dev
framework = arduino
monitor_speed = 115200
lib_deps =
    olehs/PZEM004Tv30
    tmrh20/RF24
```

## 3.3 Contoh `platformio.ini` — Gateway Node

```ini
[env:esp32dev]
platform = espressif32
board = esp32dev
framework = arduino
monitor_speed = 115200
lib_deps =
    tmrh20/RF24
    knolleary/PubSubClient
    bblanchon/ArduinoJson
```

Catatan: board `esp32dev` adalah profil generik ESP32 DevKitC yang umum dipakai di PlatformIO Registry dan kompatibel dengan ESP32 DevKitC V4 WROOM-32U.

## 3.4 Build & Upload

```bash
# dari dalam folder firmware/sensor-node atau firmware/gateway-node
pio run                 # compile
pio run -t upload       # compile + upload ke board via USB
pio device monitor      # buka serial monitor untuk debugging
```

## 3.5 Library yang digunakan

| Kebutuhan | Library |
|---|---|
| Baca data PZEM-004T v3 | `PZEM004Tv30` (pastikan versi v3.0, bukan library PZEM versi lama) |
| Komunikasi NRF24L01 | `RF24` (by TMRh20) |
| MQTT client (di Gateway Node) | `PubSubClient` (Nick O'Leary) — ingat `setBufferSize()` jika payload JSON > 256 byte |
| JSON | `ArduinoJson` |

## 3.6 Sensor Node

Alur program:
1. Setup: inisialisasi UART2 (PZEM), inisialisasi SPI + RF24 sebagai TX, set address pipe tujuan (alamat Gateway Node)
2. Loop setiap 2-5 detik:
   - Baca semua parameter dari PZEM-004T (voltage, current, power, energy, frequency, power_factor, alarm status)
   - Jika pembacaan valid, kemas ke dalam struct (lebih efisien dari JSON untuk RF24 karena payload RF24 terbatas ~32 byte default)
   - Kirim struct via `radio.write()` ke Gateway Node
   - (Opsional) retry singkat jika `write()` gagal (tidak ada ACK)

Contoh struct payload (binary, ringkas untuk RF24):
```cpp
struct SensorData {
  float voltage;
  float current;
  float power;
  float energy;
  float frequency;
  float powerFactor;
  uint8_t alarm;
  uint32_t seq; // sequence number untuk deteksi packet loss
};
```

## 3.7 Gateway Node

Alur program:
1. Setup: inisialisasi SPI + RF24 sebagai RX (listen pada address yang sama dengan tujuan Sensor Node), koneksi WiFi, koneksi MQTT (set LWT sebelum connect)
2. Setup LWT sebelum `mqttClient.connect()`:
   - Topic: `powermon/main/status`, payload: `offline`, retain: `true`
3. Setelah berhasil connect ke broker:
   - Publish manual: topic `powermon/main/status`, payload `online`, retain `true`
4. Loop:
   - Jika ada data masuk dari RF24 (`radio.available()`), baca struct `SensorData`
   - Konversi ke JSON menggunakan `ArduinoJson`
   - Publish ke topic `powermon/main/data`
   - `mqttClient.loop()` dipanggil tiap iterasi untuk menjaga koneksi MQTT & keepalive
   - Reconnect WiFi/MQTT otomatis jika terputus

Contoh payload JSON yang dipublish ke `powermon/main/data`:
```json
{
  "voltage": 221.3,
  "current": 2.15,
  "power": 475.6,
  "energy": 12.345,
  "frequency": 50.0,
  "power_factor": 0.98,
  "alarm": 0,
  "seq": 1024,
  "ts": 1732000000
}
```

## 3.8 Pengujian Awal RF24 (disarankan sebelum pasang permanen)

Sebelum instalasi final di panel listrik, uji dulu jarak & reliability NRF24L01+PA/LNA pada jarak 50-100m aktual di lokasi:
- Gunakan contoh sketch `GettingStarted` dari library RF24 untuk tes ping-pong sederhana
- Catat packet loss rate pada jarak target
- Jika packet loss tinggi, pertimbangkan: menurunkan data rate RF24 ke 250kbps (lebih tahan jarak jauh dibanding 1Mbps/2Mbps), menyesuaikan posisi antena, atau menambah PA level
</content>
