# 6. Frontend (Vue)

## 6.1 Tanggung Jawab

- Dashboard realtime: menampilkan voltage, current, power, energy, frequency, power factor terkini
- Indikator status device (online/offline) yang update otomatis via WebSocket
- Grafik live untuk parameter utama (power/current/voltage)
- Halaman histori: pilih rentang waktu, tampilkan grafik dari data raw/aggregate
- Halaman laporan: ringkasan konsumsi energi harian/bulanan

## 6.2 Struktur Halaman (usulan)

```
src/
  views/
    DashboardView.vue      - kartu realtime + grafik live + status device
    HistoryView.vue        - filter rentang waktu + grafik histori
    ReportView.vue         - ringkasan harian/bulanan (tabel + chart)
  components/
    StatusBadge.vue        - indikator online/offline
    RealtimeChart.vue      - line chart realtime (WebSocket)
    HistoryChart.vue       - line/bar chart histori (REST API)
    MetricCard.vue         - kartu angka (voltage, current, dll)
  composables/
    useWebSocket.ts         - koneksi & reconnect WebSocket, expose data reaktif
    useApi.ts                - wrapper fetch/axios ke REST API backend
  stores/ (Pinia, jika dipakai)
    deviceStore.ts           - state status device, data realtime terakhir
```

## 6.3 Komunikasi dengan Backend

- **WebSocket** (`/ws`): subscribe begitu halaman Dashboard dimuat, update reaktif ke state (voltage/current/power dst) dan status device, auto-reconnect jika koneksi putus
- **REST API**: dipakai di halaman History & Report untuk mengambil data rentang waktu tertentu (`/api/readings/history`, `/api/reports/daily`, `/api/reports/monthly`)

## 6.4 Library yang disarankan

| Kebutuhan | Library |
|---|---|
| Charting | Chart.js (via `vue-chartjs`) atau ApexCharts (via `vue3-apexcharts`) — pilih salah satu, keduanya mendukung update data realtime |
| State management | Pinia (standar Vue 3) |
| HTTP client | Axios atau `fetch` native |
| Routing | Vue Router |

## 6.5 Indikator Status Device

- Badge hijau "Online" / merah "Offline" di header dashboard, bersumber dari state `deviceStore.status` yang diupdate oleh pesan WebSocket bertipe `status`
- Tampilkan juga `last_seen` (format relatif, mis. "2 menit lalu") agar user tahu kapan terakhir device mengirim data walau sedang offline

## 6.6 Build & Serve

Build Vue sebagai static file (`npm run build`), disajikan melalui Nginx di dalam container Docker, atau di-proxy oleh backend Go langsung (opsional, tergantung preferensi saat implementasi).
</content>
