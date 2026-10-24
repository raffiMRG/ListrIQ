package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

// Semua endpoint mengembalikan array JSON baris hasil query.
func routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", auth.login)
	mux.HandleFunc("POST /api/auth/logout", auth.logout)
	mux.HandleFunc("GET /api/auth/me", auth.me)
	mux.HandleFunc("GET /api/admin/storage", auth.require(admin.storage))
	mux.HandleFunc("POST /api/admin/backup", auth.require(admin.backup))
	mux.HandleFunc("GET /api/admin/backup/{id}/download", auth.require(admin.download))
	mux.HandleFunc("POST /api/admin/restore", auth.require(admin.uploadRestore))
	mux.HandleFunc("POST /api/admin/restore/{id}/confirm", auth.require(admin.confirmRestore))
	mux.HandleFunc("GET /api/admin/jobs/{id}", auth.require(admin.status))
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		queryJSON(w, r, `SELECT device_id, status, last_seen FROM device_status ORDER BY device_id`)
	})

	mux.HandleFunc("GET /api/readings/latest", func(w http.ResponseWriter, r *http.Request) {
		queryJSON(w, r, `SELECT time AS ts, voltage, current, power, energy, frequency, power_factor, alarm
			FROM sensor_readings WHERE device_id = $1 ORDER BY time DESC LIMIT 1`, device(r))
	})

	mux.HandleFunc("GET /api/readings/history", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		to, from := time.Now(), time.Now().Add(-24*time.Hour)
		var err error
		if s := q.Get("to"); s != "" {
			if to, err = time.Parse(time.RFC3339, s); err != nil {
				http.Error(w, "to: format RFC3339", 400)
				return
			}
		}
		if s := q.Get("from"); s != "" {
			if from, err = time.Parse(time.RFC3339, s); err != nil {
				http.Error(w, "from: format RFC3339", 400)
				return
			}
		}
		sql, err := historySQL(q.Get("resolution"), to.Sub(from))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		queryJSON(w, r, sql, device(r), from, to)
	})

	mux.HandleFunc("GET /api/reports/daily", func(w http.ResponseWriter, r *http.Request) {
		report(w, r, "2006-01-02", r.URL.Query().Get("date"), "sensor_readings_hourly")
	})
	mux.HandleFunc("GET /api/reports/monthly", func(w http.ResponseWriter, r *http.Request) {
		report(w, r, "2006-01", r.URL.Query().Get("month"), "sensor_readings_daily")
	})
}

func device(r *http.Request) string {
	if d := r.URL.Query().Get("device"); d != "" {
		return d
	}
	return "main"
}

// historySQL memilih tabel sumber dari resolution; kosong/"auto" = pilih berdasar panjang rentang.
func historySQL(res string, span time.Duration) (string, error) {
	if res == "" || res == "auto" {
		switch {
		case span <= 2*24*time.Hour:
			res = "raw"
		case span <= 60*24*time.Hour:
			res = "hourly"
		default:
			res = "daily"
		}
	}
	switch res {
	case "raw":
		return `SELECT time AS ts, voltage, current, power, energy, frequency, power_factor
			FROM sensor_readings WHERE device_id = $1 AND time >= $2 AND time < $3 ORDER BY time`, nil
	case "hourly", "daily":
		return `SELECT bucket AS ts, avg_voltage AS voltage, avg_current AS current, avg_power AS power,
			max_power, energy_delta, avg_frequency AS frequency, avg_power_factor AS power_factor
			FROM sensor_readings_` + res + ` WHERE device_id = $1 AND bucket >= $2 AND bucket < $3 ORDER BY bucket`, nil
	}
	return "", fmt.Errorf("resolution harus raw, hourly, daily, atau auto")
}

// reportRange: "2026-10-22" -> [00:00, besok 00:00) WIB; "2026-10" -> [tgl 1, bulan depan).
func reportRange(layout, s string, now time.Time) (from, to time.Time, err error) {
	if s == "" {
		s = now.In(loc).Format(layout)
	}
	if from, err = time.ParseInLocation(layout, s, loc); err != nil {
		return
	}
	if layout == "2006-01" {
		return from, from.AddDate(0, 1, 0), nil
	}
	return from, from.AddDate(0, 0, 1), nil
}

func report(w http.ResponseWriter, r *http.Request, layout, s, table string) {
	from, to, err := reportRange(layout, s, time.Now())
	if err != nil {
		http.Error(w, "format tanggal harus "+layout, 400)
		return
	}
	queryJSON(w, r, `SELECT bucket AS ts, energy_delta AS energy, avg_power, max_power
		FROM `+table+` WHERE device_id = $1 AND bucket >= $2 AND bucket < $3 ORDER BY bucket`, device(r), from, to)
}

func queryJSON(w http.ResponseWriter, r *http.Request, sql string, args ...any) {
	rows, err := db.Query(r.Context(), sql, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if out == nil {
		out = []map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
