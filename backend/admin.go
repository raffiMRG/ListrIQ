package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

type adminJob struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	File      string    `json:"-"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
}
type adminService struct {
	dir, tmp    string
	maxUpload   int64
	mu          sync.Mutex
	jobs        map[string]*adminJob
	maintenance bool
}

func newAdmin() (*adminService, error) {
	max, err := strconv.ParseInt(env("ADMIN_RESTORE_MAX_BYTES", "2147483648"), 10, 64)
	if err != nil {
		return nil, err
	}
	a := &adminService{dir: env("ADMIN_BACKUP_DIR", "/data/backups"), tmp: env("ADMIN_TMP_DIR", "/data/tmp"), maxUpload: max, jobs: map[string]*adminJob{}}
	if err := os.MkdirAll(a.dir, 0700); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(a.tmp, 0700); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *adminService) isMaintenance() bool { a.mu.Lock(); defer a.mu.Unlock(); return a.maintenance }
func (a *adminService) newJob(typ string) *adminJob {
	j := &adminJob{ID: randomToken()[:12], Type: typ, Status: "queued", CreatedAt: time.Now()}
	a.mu.Lock()
	a.jobs[j.ID] = j
	a.mu.Unlock()
	return j
}
func (a *adminService) update(j *adminJob, status, message string) {
	a.mu.Lock()
	j.Status = status
	if status == "failed" {
		j.Error = message
	}
	a.mu.Unlock()
}

func (a *adminService) storage(w http.ResponseWriter, r *http.Request) {
	var database, readings int64
	if err := db.QueryRow(r.Context(), "SELECT pg_database_size(current_database()), pg_total_relation_size('sensor_readings')").Scan(&database, &readings); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"database_bytes": database, "readings_bytes": readings, "backup_bytes": dirSize(a.dir), "backup_files": files(a.dir), "maintenance": a.isMaintenance()})
}

func (a *adminService) backup(w http.ResponseWriter, r *http.Request) {
	j := a.newJob("backup")
	go func() {
		a.update(j, "running", "")
		j.File = filepath.Join(a.tmp, "powermon-"+time.Now().Format("20060102-150405")+".dump")
		cmd := exec.Command("pg_dump", "--format=custom", "--file="+j.File, env("DB_DSN", ""))
		if out, err := cmd.CombinedOutput(); err != nil {
			a.update(j, "failed", string(out))
			return
		}
		info, _ := os.Stat(j.File)
		if info != nil {
			j.Size = info.Size()
		}
		a.update(j, "ready", "")
	}()
	writeJSON(w, j)
}

func (a *adminService) download(w http.ResponseWriter, r *http.Request) {
	j := a.job(r.PathValue("id"))
	if j == nil || j.Type != "backup" || j.Status != "ready" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(j.File))
	http.ServeFile(w, r, j.File)
}

func (a *adminService) uploadRestore(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxUpload)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "file terlalu besar atau form tidak valid", 400)
		return
	}
	f, _, err := r.FormFile("backup")
	if err != nil {
		http.Error(w, "field backup diperlukan", 400)
		return
	}
	defer f.Close()
	j := a.newJob("restore")
	j.Status = "awaiting_confirmation"
	j.File = filepath.Join(a.tmp, "restore-"+j.ID+".dump")
	out, err := os.Create(j.File)
	if err == nil {
		_, err = io.Copy(out, f)
		out.Close()
	}
	if err != nil {
		os.Remove(j.File)
		a.update(j, "failed", err.Error())
		http.Error(w, "gagal menyimpan upload", 500)
		return
	}
	if output, err := exec.Command("pg_restore", "--list", j.File).CombinedOutput(); err != nil || len(output) == 0 {
		os.Remove(j.File)
		a.update(j, "failed", "backup tidak valid")
		http.Error(w, "backup PostgreSQL tidak valid", 400)
		return
	}
	info, _ := os.Stat(j.File)
	if info != nil {
		j.Size = info.Size()
	}
	writeJSON(w, j)
}

func (a *adminService) confirmRestore(w http.ResponseWriter, r *http.Request) {
	j := a.job(r.PathValue("id"))
	if j == nil || j.Type != "restore" || j.Status != "awaiting_confirmation" {
		http.Error(w, "job restore tidak tersedia", 409)
		return
	}
	var body struct {
		Confirmation string `json:"confirmation"`
	}
	if decodeJSON(r, &body) != nil || body.Confirmation != "RESTORE" {
		http.Error(w, "konfirmasi harus RESTORE", 400)
		return
	}
	go func() {
		a.mu.Lock()
		if a.maintenance {
			a.mu.Unlock()
			a.update(j, "failed", "restore lain sedang berjalan")
			return
		}
		a.maintenance = true
		a.mu.Unlock()
		defer func() { a.mu.Lock(); a.maintenance = false; a.mu.Unlock() }()
		a.update(j, "running", "")
		pre := filepath.Join(a.dir, "pre-restore-"+time.Now().Format("20060102-150405")+".dump")
		if out, err := exec.Command("pg_dump", "--format=custom", "--file="+pre, env("DB_DSN", "")).CombinedOutput(); err != nil {
			a.update(j, "failed", "pre-restore backup: "+string(out))
			return
		}
		if out, err := exec.Command("pg_restore", "--clean", "--if-exists", "--no-owner", "--no-privileges", "--dbname="+env("DB_DSN", ""), j.File).CombinedOutput(); err != nil {
			a.update(j, "failed", string(out))
			return
		}
		var n int
		if err := db.QueryRow(context.Background(), "SELECT count(*) FROM sensor_readings").Scan(&n); err != nil {
			a.update(j, "failed", "health check: "+err.Error())
			return
		}
		a.update(j, "completed", "")
	}()
	writeJSON(w, j)
}

func (a *adminService) status(w http.ResponseWriter, r *http.Request) {
	j := a.job(r.PathValue("id"))
	if j == nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, j)
}
func (a *adminService) job(id string) *adminJob { a.mu.Lock(); defer a.mu.Unlock(); return a.jobs[id] }
func dirSize(dir string) int64 {
	var total int64
	_ = filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}
func files(dir string) []map[string]any {
	entries, _ := os.ReadDir(dir)
	out := make([]map[string]any, 0)
	for _, e := range entries {
		if i, err := e.Info(); err == nil && !i.IsDir() {
			out = append(out, map[string]any{"name": e.Name(), "size": i.Size(), "modified": i.ModTime()})
		}
	}
	sort.Slice(out, func(i, j int) bool { return fmt.Sprint(out[i]["modified"]) > fmt.Sprint(out[j]["modified"]) })
	return out
}
func decodeJSON(r *http.Request, v any) error { return json.NewDecoder(r.Body).Decode(v) }
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
