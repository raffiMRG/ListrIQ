# 5. Database (PostgreSQL + TimescaleDB)

Menggunakan image Docker `timescale/timescaledb` (PostgreSQL penuh + ekstensi TimescaleDB).

## 5.1 Skema Tabel

### `sensor_readings` (hypertable)

```sql
CREATE TABLE sensor_readings (
    time            TIMESTAMPTZ NOT NULL,
    device_id       TEXT NOT NULL DEFAULT 'main',
    voltage         DOUBLE PRECISION,
    current         DOUBLE PRECISION,
    power           DOUBLE PRECISION,
    energy          DOUBLE PRECISION,
    frequency       DOUBLE PRECISION,
    power_factor    DOUBLE PRECISION,
    alarm           SMALLINT DEFAULT 0
);

SELECT create_hypertable('sensor_readings', 'time');

CREATE INDEX ON sensor_readings (device_id, time DESC);
```

### `device_status`

```sql
CREATE TABLE device_status (
    device_id   TEXT PRIMARY KEY,
    status      TEXT NOT NULL DEFAULT 'offline',
    last_seen   TIMESTAMPTZ
);
```

## 5.2 Continuous Aggregate

### Per jam

```sql
CREATE MATERIALIZED VIEW sensor_readings_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    device_id,
    avg(voltage) AS avg_voltage,
    avg(current) AS avg_current,
    avg(power) AS avg_power,
    max(power) AS max_power,
    max(energy) - min(energy) AS energy_delta,
    avg(frequency) AS avg_frequency,
    avg(power_factor) AS avg_power_factor
FROM sensor_readings
GROUP BY bucket, device_id;

SELECT add_continuous_aggregate_policy('sensor_readings_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');
```

### Per hari

```sql
CREATE MATERIALIZED VIEW sensor_readings_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    device_id,
    avg(voltage) AS avg_voltage,
    avg(current) AS avg_current,
    avg(power) AS avg_power,
    max(power) AS max_power,
    max(energy) - min(energy) AS energy_delta,
    avg(frequency) AS avg_frequency,
    avg(power_factor) AS avg_power_factor
FROM sensor_readings
GROUP BY bucket, device_id;

SELECT add_continuous_aggregate_policy('sensor_readings_daily',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');
```

Catatan: `energy_delta` dihitung dari selisih max-min nilai `energy` kumulatif dalam bucket waktu tersebut, karena register energy dari PZEM-004T bersifat akumulatif (bukan reset tiap pembacaan).

## 5.3 Retention Policy

Agar storage raw data tidak terus membesar, raw data lama bisa di-drop sementara continuous aggregate (hourly/daily) tetap disimpan permanen:

```sql
SELECT add_retention_policy('sensor_readings', INTERVAL '90 days');
```

Sesuaikan durasi retensi (90 hari adalah contoh) sesuai kebutuhan dan kapasitas storage server.

## 5.4 Query Histori untuk Frontend

- Rentang pendek (jam/hari terakhir) -> query langsung ke `sensor_readings` (raw)
- Rentang menengah (minggu/bulan) -> query ke `sensor_readings_hourly`
- Rentang panjang (bulan/tahun) -> query ke `sensor_readings_daily`

Backend memilih tabel sumber berdasarkan parameter `resolution` pada endpoint `/api/readings/history`.
</content>
