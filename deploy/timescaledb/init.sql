-- Dijalankan otomatis sekali saat volume TimescaleDB pertama kali dibuat.
CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE sensor_readings (
    time            TIMESTAMPTZ NOT NULL,
    device_id       TEXT NOT NULL DEFAULT 'main',
    sensor_ok       BOOLEAN DEFAULT true, -- false jika PZEM gagal dibaca saat sampling ini
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

CREATE TABLE device_status (
    device_id   TEXT PRIMARY KEY,
    status      TEXT NOT NULL DEFAULT 'offline',
    last_seen   TIMESTAMPTZ
);

-- energy_delta = max-min karena register energy PZEM akumulatif.
-- ponytail: reset counter PZEM di tengah bucket membuat delta bucket itu salah (terlalu besar);
-- reset itu manual & jarang, ganti ke sum-of-positive-diffs jika ternyata sering terjadi.
-- materialized_only=false: data terbaru yang belum di-refresh tetap ikut (real-time aggregate).
CREATE MATERIALIZED VIEW sensor_readings_hourly
WITH (timescaledb.continuous, timescaledb.materialized_only = false) AS
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

-- Bucket harian mengikuti hari lokal WIB, bukan UTC (UTC 00:00 = 07:00 WIB).
CREATE MATERIALIZED VIEW sensor_readings_daily
WITH (timescaledb.continuous, timescaledb.materialized_only = false) AS
SELECT
    time_bucket('1 day', time, 'Asia/Jakarta') AS bucket,
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

-- Raw data dibuang setelah 90 hari; aggregate hourly/daily tetap permanen.
SELECT add_retention_policy('sensor_readings', INTERVAL '90 days');
