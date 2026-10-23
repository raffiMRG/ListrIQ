package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	db  *pgxpool.Pool
	loc *time.Location
	hub = &Hub{conns: map[*websocket.Conn]bool{}}
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	var err error
	if loc, err = time.LoadLocation(env("TZ", "Asia/Jakarta")); err != nil {
		log.Fatal(err)
	}
	// Watchdog lapis kedua: device dianggap offline jika tidak ada data selama ini (>= 3x interval sampling).
	offlineAfter, err := time.ParseDuration(env("OFFLINE_AFTER", "20s"))
	if err != nil {
		log.Fatal(err)
	}
	if db, err = pgxpool.New(context.Background(), env("DB_DSN", "postgres://powermon:powermon@localhost:5432/powermon")); err != nil {
		log.Fatal(err)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(env("MQTT_BROKER", "tcp://localhost:1883")).
		SetClientID("powermon-backend").
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Println("mqtt connected")
			c.Subscribe("powermon/+/data", 1, onData)
			c.Subscribe("powermon/+/status", 1, onStatus)
		})
	mqtt.NewClient(opts).Connect()

	go watchdog(offlineAfter)

	mux := http.NewServeMux()
	routes(mux)
	mux.HandleFunc("GET /ws", hub.serve)
	mux.Handle("/", http.FileServer(http.Dir(env("STATIC_DIR", "web"))))
	addr := env("HTTP_ADDR", ":8080")
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// deviceID ambil "main" dari "powermon/main/data".
func deviceID(topic string) string { return strings.Split(topic, "/")[1] }

type Reading struct {
	DeviceID    string    `json:"device_id"`
	Time        time.Time `json:"ts"`
	SensorOK    bool      `json:"sensor_ok"` // false jika PZEM gagal dibaca; terpisah dari status perangkat (MQTT heartbeat)
	Voltage     *float64  `json:"voltage"`   // pointer: NaN dari PZEM dikirim sebagai null
	Current     *float64  `json:"current"`
	Power       *float64  `json:"power"`
	Energy      *float64  `json:"energy"`
	Frequency   *float64  `json:"frequency"`
	PowerFactor *float64  `json:"power_factor"`
	Alarm       int16     `json:"alarm"`
}

func onData(_ mqtt.Client, m mqtt.Message) {
	var r Reading
	if err := json.Unmarshal(m.Payload(), &r); err != nil {
		log.Printf("bad payload on %s: %v", m.Topic(), err)
		return
	}
	// Gateway tidak punya jam (RTC/NTP), jadi timestamp diberi saat diterima server.
	r.DeviceID, r.Time = deviceID(m.Topic()), time.Now()
	ctx := context.Background()
	_, err := db.Exec(ctx, `INSERT INTO sensor_readings (time, device_id, sensor_ok, voltage, current, power, energy, frequency, power_factor, alarm)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		r.Time, r.DeviceID, r.SensorOK, r.Voltage, r.Current, r.Power, r.Energy, r.Frequency, r.PowerFactor, r.Alarm)
	if err == nil {
		// Status perangkat (online/offline) murni dari heartbeat MQTT; terlepas dari sensor_ok.
		_, err = db.Exec(ctx, `INSERT INTO device_status (device_id, status, last_seen) VALUES ($1, 'online', $2)
			ON CONFLICT (device_id) DO UPDATE SET status = 'online', last_seen = $2`, r.DeviceID, r.Time)
	}
	if err != nil {
		log.Println("db:", err)
	}
	hub.broadcast("reading", r)
}

type Status struct {
	DeviceID string     `json:"device_id"`
	Status   string     `json:"status"`
	LastSeen *time.Time `json:"last_seen"`
}

// onStatus hanya menangani "offline" (LWT gateway). "online" dari gateway tidak berarti sensor
// node mengirim data, jadi status online hanya ditetapkan oleh data yang benar-benar masuk.
func onStatus(_ mqtt.Client, m mqtt.Message) {
	if string(m.Payload()) != "offline" {
		return
	}
	s := Status{DeviceID: deviceID(m.Topic()), Status: "offline"}
	err := db.QueryRow(context.Background(), `INSERT INTO device_status (device_id, status) VALUES ($1, 'offline')
		ON CONFLICT (device_id) DO UPDATE SET status = 'offline' RETURNING last_seen`, s.DeviceID).Scan(&s.LastSeen)
	if err != nil {
		log.Println("db:", err)
		return
	}
	hub.broadcast("status", s)
}

func watchdog(offlineAfter time.Duration) {
	for range time.Tick(1 * time.Second) {
		rows, err := db.Query(context.Background(), `UPDATE device_status SET status = 'offline'
			WHERE status = 'online' AND last_seen < now() - make_interval(secs => $1)
			RETURNING device_id, last_seen`, offlineAfter.Seconds())
		if err != nil {
			log.Println("watchdog:", err)
			continue
		}
		for rows.Next() {
			s := Status{Status: "offline"}
			if rows.Scan(&s.DeviceID, &s.LastSeen) == nil {
				hub.broadcast("status", s)
			}
		}
		rows.Close()
	}
}

// Hub menyimpan koneksi WebSocket aktif dan mengirim pesan ke semuanya.
// ponytail: satu lock untuk semua tulisan, client lambat menahan yang lain sampai write deadline (2s); cukup untuk LAN rumah.
type Hub struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]bool
}

var upgrader websocket.Upgrader

func (h *Hub) serve(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.mu.Lock()
	h.conns[c] = true
	h.mu.Unlock()
	for { // pesan dari client diabaikan; loop ini hanya untuk mendeteksi koneksi tutup
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
	h.mu.Lock()
	delete(h.conns, c)
	h.mu.Unlock()
	c.Close()
}

func (h *Hub) broadcast(typ string, data any) {
	msg, _ := json.Marshal(map[string]any{"type": typ, "data": data})
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.conns {
		c.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if c.WriteMessage(websocket.TextMessage, msg) != nil {
			c.Close() // read loop di serve() akan menghapusnya dari map
		}
	}
}
