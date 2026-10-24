package main

import (
	"strings"
	"testing"
	"time"
)

func TestPasswordHash(t *testing.T) {
	hash := hashPassword("rahasia-admin")
	if !verifyPassword("rahasia-admin", hash) {
		t.Fatal("password yang benar harus diterima")
	}
	if verifyPassword("salah", hash) {
		t.Fatal("password yang salah harus ditolak")
	}
	if verifyPassword("rahasia-admin", "bukan-hash") {
		t.Fatal("hash invalid harus ditolak")
	}
}

func TestHistorySQL(t *testing.T) {
	day := 24 * time.Hour
	for _, c := range []struct {
		res   string
		span  time.Duration
		table string
	}{
		{"", day, "FROM sensor_readings WHERE"},
		{"auto", 7 * day, "sensor_readings_hourly"},
		{"", 365 * day, "sensor_readings_daily"},
		{"raw", 365 * day, "FROM sensor_readings WHERE"},
	} {
		sql, err := historySQL(c.res, c.span)
		if err != nil || !strings.Contains(sql, c.table) {
			t.Errorf("historySQL(%q, %v) = %q, %v; want %s", c.res, c.span, sql, err, c.table)
		}
	}
	if _, err := historySQL("x; DROP TABLE", day); err == nil {
		t.Error("resolution invalid harus ditolak")
	}
}

func TestReportRange(t *testing.T) {
	loc, _ = time.LoadLocation("Asia/Jakarta")
	from, to, _ := reportRange("2006-01-02", "2026-10-22", time.Now())
	if from.UTC().Format(time.RFC3339) != "2026-10-21T17:00:00Z" || to.Sub(from) != 24*time.Hour {
		t.Errorf("daily: %v - %v", from, to)
	}
	from, to, _ = reportRange("2006-01", "2026-12", time.Now())
	if to.Format("2006-01-02") != "2027-01-01" || from.Format("2006-01-02") != "2026-12-01" {
		t.Errorf("monthly: %v - %v", from, to)
	}
	// default = hari ini waktu WIB: 2026-10-22 20:00 UTC sudah 2026-10-23 di WIB
	from, _, _ = reportRange("2006-01-02", "", time.Date(2026, 10, 22, 20, 0, 0, 0, time.UTC))
	if from.Format("2006-01-02") != "2026-10-23" {
		t.Errorf("default date: %v", from)
	}
	if _, _, err := reportRange("2006-01", "bukan-bulan", time.Now()); err == nil {
		t.Error("tanggal invalid harus error")
	}
}
