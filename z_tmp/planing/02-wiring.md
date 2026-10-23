# 2. Wiring Hardware

## 2.1 Sensor Node (di panel listrik utama)

Terdiri dari: ESP32 DevKitC V4 + PZEM-004T v3 + NRF24L01+PA/LNA

### PZEM-004T v3 -> ESP32

PZEM-004T berkomunikasi via UART (Modbus-RTU, 9600 baud, 8N1, alamat default 0xF8).

| Pin PZEM-004T | Pin ESP32 | Keterangan |
|---|---|---|
| 5V | 5V (VIN) | Power logic modul (bukan 3.3V) |
| GND | GND | |
| TX | GPIO16 (RX2) | TX PZEM -> RX ESP32 (silang) |
| RX | GPIO17 (TX2) | RX PZEM -> TX ESP32 (silang) |

Sisi AC PZEM-004T (terpisah dari jalur komunikasi di atas):
- Konektor AC (L, N) PZEM disambung ke jalur listrik utama (PLN) yang ingin diukur
- CT (current transformer) split-core dijepitkan ke kabel **Line** (bukan Netral), dengan arah panah pada CT mengarah ke sisi beban (load)

**PERHATIAN KESELAMATAN:** Bagian AC PZEM-004T terhubung langsung ke tegangan listrik PLN (220V). Pastikan daya listrik dimatikan (MCB off) saat melakukan pemasangan/pengkabelan, gunakan isolasi yang baik, dan jika tidak yakin, minta bantuan orang yang berpengalaman dengan instalasi listrik.

### NRF24L01+PA/LNA -> ESP32 (Sensor Node)

| Pin NRF24L01 | Pin ESP32 | Keterangan |
|---|---|---|
| VCC | 3.3V | **Wajib 3.3V**, jangan 5V (modul bisa rusak) |
| GND | GND | |
| CE | GPIO4 | Chip Enable |
| CSN | GPIO5 | Chip Select (SPI SS) |
| SCK | GPIO18 | SPI Clock (VSPI) |
| MOSI | GPIO23 | SPI MOSI (VSPI) |
| MISO | GPIO19 | SPI MISO (VSPI) |
| IRQ | - | Tidak dipakai (opsional) |

**Catatan penting:** Tambahkan kapasitor 10-100µF antara VCC dan GND, dipasang sedekat mungkin ke pin modul NRF24L01. Modul (terutama varian PA/LNA) rawan brownout saat transmit karena lonjakan arus sesaat (~115mA+) yang bisa membuat tegangan 3.3V ESP32 drop dan menyebabkan koneksi SPI tidak stabil.

## 2.2 Gateway Node (dekat router/modem)

Terdiri dari: ESP32 DevKitC V4 + NRF24L01+PA/LNA (tanpa PZEM)

Wiring NRF24L01 -> ESP32 **sama persis** dengan tabel di atas (2.1). Tidak ada sambungan UART karena tidak ada sensor PZEM di node ini.

ESP32 Gateway Node cukup disuplai power via USB/adaptor 5V dan ditempatkan di lokasi dengan sinyal WiFi baik, idealnya juga punya line-of-sight yang relatif bersih ke Sensor Node untuk jarak 50-100m.

## 2.3 Diagram Pin Summary (kedua node)

```
ESP32 DevKitC V4
  GPIO4  ---- CE    (NRF24L01)
  GPIO5  ---- CSN   (NRF24L01)
  GPIO18 ---- SCK   (NRF24L01)
  GPIO19 ---- MISO  (NRF24L01)
  GPIO23 ---- MOSI  (NRF24L01)
  3.3V   ---- VCC   (NRF24L01, + kapasitor 10-100uF)
  GND    ---- GND   (NRF24L01)

  [Hanya di Sensor Node:]
  GPIO16 (RX2) ---- TX (PZEM-004T)
  GPIO17 (TX2) ---- RX (PZEM-004T)
  5V           ---- 5V (PZEM-004T)
  GND          ---- GND (PZEM-004T)
```
</content>
