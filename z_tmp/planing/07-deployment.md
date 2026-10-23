# 7. Deployment (Docker Compose)

Target: home server/NAS, tanpa domain publik, diakses via IP LAN.

## 7.1 Struktur Project (usulan)

```
monitoring-listrik/
  firmware/
    sensor-node/            - sketch ESP32 Sensor Node
    gateway-node/           - sketch ESP32 Gateway Node
  backend/                  - source Golang
    Dockerfile
  frontend/                 - source Vue
    Dockerfile
  deploy/
    docker-compose.yml
    mosquitto/
      mosquitto.conf
    timescaledb/
      init.sql              - script CREATE TABLE, hypertable, continuous aggregate
    .env.example
  z_tmp/
    device.md
    planing/                - dokumen ini
```

## 7.2 docker-compose.yml (kerangka)

```yaml
version: "3.9"

services:
  mosquitto:
    image: eclipse-mosquitto:2
    restart: unless-stopped
    ports:
      - "1883:1883"
    volumes:
      - ./mosquitto/mosquitto.conf:/mosquitto/config/mosquitto.conf
      - mosquitto_data:/mosquitto/data

  timescaledb:
    image: timescale/timescaledb:latest-pg16
    restart: unless-stopped
    environment:
      POSTGRES_DB: powermon
      POSTGRES_USER: powermon
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - timescale_data:/var/lib/postgresql/data
      - ./timescaledb/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"

  backend:
    build: ../backend
    restart: unless-stopped
    environment:
      MQTT_BROKER: tcp://mosquitto:1883
      DB_DSN: postgres://powermon:${DB_PASSWORD}@timescaledb:5432/powermon
    depends_on:
      - mosquitto
      - timescaledb
    ports:
      - "8080:8080"

  frontend:
    build: ../frontend
    restart: unless-stopped
    ports:
      - "8081:80"
    depends_on:
      - backend

volumes:
  mosquitto_data:
  timescale_data:
```

## 7.3 Konfigurasi Mosquitto (dasar)

```
listener 1883
allow_anonymous true
persistence true
persistence_location /mosquitto/data/
```

Catatan: `allow_anonymous true` cukup untuk home server di LAN tertutup. Jika jaringan rumah tidak sepenuhnya dipercaya (banyak perangkat lain/tamu), pertimbangkan menambah username/password MQTT di konfigurasi ini nanti.

## 7.4 Environment Variables (.env.example)

```
DB_PASSWORD=ganti_dengan_password_aman
```

## 7.5 Langkah Deploy

1. `git clone`/copy project ke home server
2. Isi `.env` berdasarkan `.env.example`
3. `docker compose up -d --build` dari folder `deploy/`
4. Backend otomatis menjalankan `init.sql` saat TimescaleDB pertama kali start (buat tabel, hypertable, continuous aggregate)
5. Akses frontend via `http://<ip-server>:8081`
6. Pastikan Gateway Node ESP32 dikonfigurasi dengan IP/hostname server dan port MQTT (1883) yang sesuai di jaringan LAN
</content>
