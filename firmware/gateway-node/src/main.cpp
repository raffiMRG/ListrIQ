#include <Arduino.h>
#include <SPI.h>
#include <WiFi.h>
#include <RF24.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>
#include "config.h"
#include "link.h"

RF24 radio(RF_CE_PIN, RF_CSN_PIN);
WiFiClient net;
PubSubClient mqtt(net);

bool mqttConnect() {
  String id = "powermon-gw-" + String((uint32_t)ESP.getEfuseMac(), HEX);
  // LWT: broker publish "offline" (retained) jika gateway putus tidak normal
  if (!mqtt.connect(id.c_str(), MQTT_USER, MQTT_PASS, TOPIC_STATUS, 1, true, "offline")) return false;
  mqtt.publish(TOPIC_STATUS, "online", true);
  return true;
}

void setup() {
  Serial.begin(115200);
  if (!radio.begin()) Serial.println("NRF24 tidak terdeteksi, cek wiring/power 3.3V");
  radio.setChannel(RF_CHANNEL);
  radio.setDataRate(RF_DATA_RATE);
  radio.setPALevel(RF_PA_LEVEL);
  radio.setPayloadSize(sizeof(SensorData));
  radio.openReadingPipe(1, RF_ADDRESS);
  radio.startListening();

  WiFi.mode(WIFI_STA);
  WiFi.setAutoReconnect(true);
  WiFi.begin(WIFI_SSID, WIFI_PASS);

  mqtt.setServer(MQTT_HOST, MQTT_PORT);
  mqtt.setBufferSize(512);
}

void loop() {
  static uint32_t lastTry = 0;
  if (WiFi.status() == WL_CONNECTED && !mqtt.connected() && millis() - lastTry > 5000) {
    lastTry = millis();
    Serial.println(mqttConnect() ? "MQTT terhubung" : "MQTT gagal, coba lagi 5 detik");
  }
  mqtt.loop();

  while (radio.available()) {
    SensorData d;
    radio.read(&d, sizeof(d));
    if (!mqtt.connected()) continue;  // tidak ada buffer; data saat offline dibuang

    JsonDocument doc;
    doc["voltage"] = d.voltage;
    doc["current"] = d.current;
    doc["power"] = d.power;
    doc["energy"] = d.energy;
    doc["frequency"] = d.frequency;
    doc["power_factor"] = d.powerFactor;
    doc["alarm"] = d.alarm;
    doc["seq"] = d.seq;
    char buf[256];
    size_t n = serializeJson(doc, buf);
    mqtt.publish(TOPIC_DATA, (const uint8_t*)buf, n, false);
    Serial.println(buf);
  }
}
