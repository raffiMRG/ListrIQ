package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

type session struct{ expires time.Time }

type authService struct {
	username string
	hash     string
	ttl      time.Duration
	mu       sync.Mutex
	sessions map[string]session
	failed   map[string][]time.Time
}

func newAuth() (*authService, error) {
	ttl, err := time.ParseDuration(env("ADMIN_SESSION_TTL", "8h"))
	if err != nil {
		return nil, err
	}
	return &authService{username: osEnv("ADMIN_USERNAME"), hash: osEnv("ADMIN_PASSWORD_HASH"), ttl: ttl, sessions: map[string]session{}, failed: map[string][]time.Time{}}, nil
}

func osEnv(key string) string { return strings.TrimSpace(env(key, "")) }

func (a *authService) enabled() bool { return a.username != "" && a.hash != "" }

func (a *authService) login(w http.ResponseWriter, r *http.Request) {
	if !a.enabled() {
		http.Error(w, "admin belum dikonfigurasi", http.StatusServiceUnavailable)
		return
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	a.mu.Lock()
	now := time.Now()
	valid := a.failed[ip][:0]
	for _, t := range a.failed[ip] {
		if now.Sub(t) < 15*time.Minute {
			valid = append(valid, t)
		}
	}
	a.failed[ip] = valid
	if len(valid) >= 5 {
		a.mu.Unlock()
		http.Error(w, "terlalu banyak percobaan login", http.StatusTooManyRequests)
		return
	}
	a.mu.Unlock()
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "payload login tidak valid", 400)
		return
	}
	if body.Username != a.username || !verifyPassword(body.Password, a.hash) {
		a.mu.Lock()
		a.failed[ip] = append(a.failed[ip], now)
		a.mu.Unlock()
		http.Error(w, "username atau password salah", http.StatusUnauthorized)
		return
	}
	key := randomToken()
	a.mu.Lock()
	a.sessions[key] = session{expires: now.Add(a.ttl)}
	delete(a.failed, ip)
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "powermon_admin", Value: key, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: int(a.ttl.Seconds())})
	writeJSON(w, map[string]any{"authenticated": true, "username": a.username})
}

func (a *authService) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("powermon_admin"); err == nil {
		a.mu.Lock()
		delete(a.sessions, c.Value)
		a.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: "powermon_admin", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	writeJSON(w, map[string]bool{"authenticated": false})
}

func (a *authService) me(w http.ResponseWriter, r *http.Request) {
	if a.authorized(r) {
		writeJSON(w, map[string]any{"authenticated": true, "username": a.username})
		return
	}
	writeJSON(w, map[string]bool{"authenticated": false})
}

func (a *authService) authorized(r *http.Request) bool {
	c, err := r.Cookie("powermon_admin")
	if err != nil {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	s, ok := a.sessions[c.Value]
	if !ok || time.Now().After(s.expires) {
		delete(a.sessions, c.Value)
		return false
	}
	return true
}

func (a *authService) require(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.authorized(r) {
			http.Error(w, "login admin diperlukan", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host && origin != "https://"+r.Host {
				http.Error(w, "origin ditolak", http.StatusForbidden)
				return
			}
		}
		next(w, r)
	}
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[4])
	expected, err2 := base64.RawStdEncoding.DecodeString(parts[5])
	if err1 != nil || err2 != nil {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func hashPassword(password string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	const memory = 65536
	const iterations = 3
	const parallelism = 2
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, 32)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memory, iterations, parallelism, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
}
