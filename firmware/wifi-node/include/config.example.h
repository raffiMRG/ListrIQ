#pragma once

// Salin file ini menjadi config.h, lalu isi dengan konfigurasi lokal.
// Jangan commit config.h karena berisi kredensial dan alamat jaringan.
#define WIFI_SSID "ganti_ssid"
#define WIFI_PASS "ganti_password"

#define MQTT_HOST "192.168.1.123" // IP atau hostname server MQTT lokal
#define MQTT_PORT 1883
#define MQTT_USER nullptr // isi "user" jika Mosquitto memakai autentikasi
#define MQTT_PASS nullptr // isi "password" jika Mosquitto memakai autentikasi

#define DEVICE_ID "main"
#define TOPIC_DATA "powermon/" DEVICE_ID "/data"
#define TOPIC_STATUS "powermon/" DEVICE_ID "/status"

// PZEM-004T v3 di UART2: TX PZEM -> GPIO16, RX PZEM -> GPIO17
#define PZEM_RX_PIN 16
#define PZEM_TX_PIN 17

// Interval sampling dan publish ke MQTT.
#define SAMPLE_INTERVAL_MS 2000

// Samakan dengan OFFLINE_AFTER pada deploy/docker-compose.yml.
#define DEVICE_OFFLINE_AFTER_MS 6000
