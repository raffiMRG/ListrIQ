// Dipakai bersama oleh sensor-node & gateway-node. Ubah di sini saja agar kedua sisi selalu cocok.
#pragma once
#include <stdint.h>
#include <RF24.h>

// Wiring NRF24L01 (sama di kedua node, lihat planing/02-wiring.md)
#define RF_CE_PIN  4
#define RF_CSN_PIN 5

// Knob tuning link RF. 250KBPS paling tahan jarak; channel 108 (2.508 GHz) di atas kanal WiFi.
// Kalau packet loss tinggi: coba channel lain (mis. 100-125) dan cek posisi antena.
#define RF_CHANNEL   108
#define RF_DATA_RATE RF24_250KBPS
#define RF_PA_LEVEL  RF24_PA_MAX  // turunkan ke RF24_PA_HIGH jika ESP32 brownout/reset saat kirim
static const uint8_t RF_ADDRESS[6] = "PWR01";

struct __attribute__((packed)) SensorData {
  float voltage;
  float current;
  float power;
  float energy;  // kWh, akumulatif
  float frequency;
  float powerFactor;
  uint8_t alarm;
  uint32_t seq;  // naik tiap kirim; lompatan = paket hilang
};
static_assert(sizeof(SensorData) <= 32, "payload NRF24 maksimal 32 byte");
