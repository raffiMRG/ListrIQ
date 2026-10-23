#include <Arduino.h>
#include <SPI.h>
#include <RF24.h>
#include <PZEM004Tv30.h>
#include "config.h"
#include "link.h"

PZEM004Tv30 pzem(Serial2, PZEM_RX_PIN, PZEM_TX_PIN);
RF24 radio(RF_CE_PIN, RF_CSN_PIN);
uint32_t seq = 0;

void setup() {
  Serial.begin(115200);
  if (!radio.begin()) Serial.println("NRF24 tidak terdeteksi, cek wiring/power 3.3V");
  radio.setChannel(RF_CHANNEL);
  radio.setDataRate(RF_DATA_RATE);
  radio.setPALevel(RF_PA_LEVEL);
  radio.setPayloadSize(sizeof(SensorData));
  radio.setRetries(5, 15);  // 1.5ms x 15 retry, cukup untuk 250kbps
  radio.openWritingPipe(RF_ADDRESS);
  radio.stopListening();
}

void loop() {
  static uint32_t last = 0;
  if (millis() - last < SAMPLE_INTERVAL_MS) return;
  last = millis();

  SensorData d;
  d.voltage = pzem.voltage();
  if (isnan(d.voltage)) {  // PZEM tidak menjawab (wiring UART / sisi AC tidak bertegangan)
    Serial.println("Gagal baca PZEM");
    return;
  }
  d.current = pzem.current();
  d.power = pzem.power();
  d.energy = pzem.energy();
  d.frequency = pzem.frequency();
  d.powerFactor = pzem.pf();
  d.alarm = pzem.getPowerAlarm();
  d.seq = seq++;

  bool ok = radio.write(&d, sizeof(d));
  Serial.printf("#%u %.1fV %.2fA %.1fW %.3fkWh %.1fHz pf%.2f -> %s\n", d.seq, d.voltage, d.current,
                d.power, d.energy, d.frequency, d.powerFactor, ok ? "OK" : "GAGAL (tidak ada ACK)");
}
