// MVP: satu ESP32 baca PZEM-004T lalu publish langsung ke MQTT via WiFi (tanpa NRF24).
#include <Arduino.h>
#include <WiFi.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>
#include <PZEM004Tv30.h>
#include "config.h"

PZEM004Tv30 pzem(Serial2, PZEM_RX_PIN, PZEM_TX_PIN);
WiFiClient net;
PubSubClient mqtt(net);
uint32_t seq = 0;

bool mqttConnect() {
  String id = "powermon-" + String((uint32_t)ESP.getEfuseMac(), HEX);
  // LWT: broker publish "offline" (retained) jika node putus tidak normal
  if (!mqtt.connect(id.c_str(), MQTT_USER, MQTT_PASS, TOPIC_STATUS, 1, true, "offline")) return false;
  mqtt.publish(TOPIC_STATUS, "online", true);
  return true;
}

void setup() {
  Serial.begin(115200);
  WiFi.mode(WIFI_STA);
  WiFi.setAutoReconnect(true);
  WiFi.begin(WIFI_SSID, WIFI_PASS);
  mqtt.setServer(MQTT_HOST, MQTT_PORT);
  mqtt.setBufferSize(512);
}

void loop() {
  static uint32_t lastTry = 0, lastSample = 0;
  if (WiFi.status() == WL_CONNECTED && !mqtt.connected() && millis() - lastTry > 5000) {
    lastTry = millis();
    Serial.printf("WiFi OK (RSSI %d dBm), MQTT %s\n", WiFi.RSSI(),
                  mqttConnect() ? "terhubung" : "gagal, coba lagi 5 detik");
  }
  mqtt.loop();

  if (millis() - lastSample < SAMPLE_INTERVAL_MS) return;
  lastSample = millis();

  float voltage = pzem.voltage();
  bool sensorOk = !isnan(voltage);
  if (!sensorOk) Serial.println("Gagal baca PZEM");  // wiring UART / sisi AC tidak bertegangan

  // Selalu publish tiap interval (sensor OK atau tidak) agar backend punya heartbeat
  // untuk status perangkat, terpisah dari status pembacaan sensor PZEM.
  JsonDocument doc;
  doc["sensor_ok"] = sensorOk;
  doc["voltage"] = sensorOk ? voltage : (float)NAN;
  doc["current"] = sensorOk ? pzem.current() : (float)NAN;
  doc["power"] = sensorOk ? pzem.power() : (float)NAN;
  doc["energy"] = sensorOk ? pzem.energy() : (float)NAN;
  doc["frequency"] = sensorOk ? pzem.frequency() : (float)NAN;
  doc["power_factor"] = sensorOk ? pzem.pf() : (float)NAN;
  doc["alarm"] = sensorOk ? pzem.getPowerAlarm() : 0;
  doc["seq"] = seq++;
  char buf[256];
  size_t n = serializeJson(doc, buf);
  Serial.println(buf);
  if (mqtt.connected()) mqtt.publish(TOPIC_DATA, (const uint8_t*)buf, n, false);  // ponytail: data saat offline dibuang, tambah buffer jika celah data jadi masalah
}
