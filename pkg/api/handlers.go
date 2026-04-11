// pkg/api/handlers.go
package api

import (
	"encoding/json"
	"fmt"
	"goScadaSur/pkg/database"
	"goScadaSur/pkg/xmlcreator"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ============ AUTH ============

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", 405)
		return
	}
	var req struct{ Username, Password, Email, FullName string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.ErrorResponse(w, "JSON inválido", 400)
		return
	}
	if s.PGClient == nil {
		s.ErrorResponse(w, "Postgres no configurado", 500)
		return
	}
	hp, err := database.HashPassword(req.Password)
	if err != nil {
		s.ErrorResponse(w, "Error hash", 500)
		return
	}
	u := database.User{Username: req.Username, Password: hp, Email: req.Email, FullName: req.FullName}
	if err := s.PGClient.DB.Create(&u).Error; err != nil {
		s.ErrorResponse(w, "Usuario ya existe", 409)
		return
	}
	s.JSONResponse(w, map[string]string{"message": "Usuario registrado"}, 201)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", 405)
		return
	}
	var req struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.ErrorResponse(w, "JSON inválido", 400)
		return
	}
	if s.PGClient == nil {
		s.ErrorResponse(w, "Postgres no configurado", 500)
		return
	}
	var u database.User
	if err := s.PGClient.DB.Where("username = ?", req.Username).First(&u).Error; err != nil {
		s.ErrorResponse(w, "Credenciales inválidas", 401)
		return
	}
	if !database.CheckPasswordHash(req.Password, u.Password) {
		s.ErrorResponse(w, "Credenciales inválidas", 401)
		return
	}
	tok, err := GenerateToken(u.ID, u.Username, u.Role)
	if err != nil {
		s.ErrorResponse(w, "Error token", 500)
		return
	}
	s.JSONResponse(w, map[string]string{"token": tok, "username": u.Username, "role": u.Role}, 200)
}

// ============ JOBS ============

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	if s.PGClient == nil {
		s.ErrorResponse(w, "Postgres no configurado", 500)
		return
	}
	switch r.Method {
	case http.MethodGet:
		var jobs []database.Job
		s.PGClient.DB.Preload("CreatedBy").Preload("Comments").Find(&jobs)
		s.JSONResponse(w, jobs, 200)
	case http.MethodPost:
		var job database.Job
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			s.ErrorResponse(w, "JSON inválido", 400)
			return
		}
		if err := s.PGClient.DB.Create(&job).Error; err != nil {
			s.ErrorResponse(w, "Error creando Job", 500)
			return
		}
		s.PGClient.DB.Create(&database.JobHistory{
			JobID: job.ID, UserID: job.CreatedByID,
			NewStatus: database.StatusPending, Action: "Job Created", CreatedAt: time.Now(),
		})
		s.JSONResponse(w, job, 201)
	case http.MethodPatch:
		var req struct {
			ID     uint               `json:"id"`
			Status database.JobStatus `json:"status"`
			UserID uint               `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.ErrorResponse(w, "JSON inválido", 400)
			return
		}
		var job database.Job
		if err := s.PGClient.DB.First(&job, req.ID).Error; err != nil {
			s.ErrorResponse(w, "Job no encontrado", 404)
			return
		}
		old := job.Status
		job.Status = req.Status
		if err := s.PGClient.DB.Save(&job).Error; err != nil {
			s.ErrorResponse(w, "Error actualizando", 500)
			return
		}
		s.PGClient.DB.Create(&database.JobHistory{
			JobID: job.ID, UserID: req.UserID,
			OldStatus: old, NewStatus: req.Status,
			Action: "Status Updated", CreatedAt: time.Now(),
		})
		s.JSONResponse(w, job, 200)
	default:
		s.ErrorResponse(w, "Método no permitido", 405)
	}
}

func (s *Server) handleJobHistory(w http.ResponseWriter, r *http.Request) {
	if s.PGClient == nil {
		s.ErrorResponse(w, "Postgres no configurado", 500)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		s.ErrorResponse(w, "ID requerido", 400)
		return
	}
	var h []database.JobHistory
	s.PGClient.DB.Where("job_id = ?", id).Order("created_at desc").Find(&h)
	s.JSONResponse(w, h, 200)
}

func (s *Server) handleComments(w http.ResponseWriter, r *http.Request) {
	if s.PGClient == nil {
		s.ErrorResponse(w, "Postgres no configurado", 500)
		return
	}
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", 405)
		return
	}
	var c database.Comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		s.ErrorResponse(w, "JSON inválido", 400)
		return
	}
	if err := s.PGClient.DB.Create(&c).Error; err != nil {
		s.ErrorResponse(w, "Error guardando", 500)
		return
	}
	s.JSONResponse(w, c, 201)
}

// ============ UPLOAD + OUTPUT ============

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		s.ErrorResponse(w, "Multipart inválido", http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		s.ErrorResponse(w, "Archivo no recibido", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("[UPLOAD] %s (%d bytes)", handler.Filename, handler.Size)

	tmp, err := os.CreateTemp("", "goscada-upload-*"+filepath.Ext(handler.Filename))
	if err != nil {
		s.ErrorResponse(w, "Error temp", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		s.ErrorResponse(w, "Error escribiendo temp", http.StatusInternalServerError)
		return
	}
	tmp.Close()

	before := snapshotOutput(s.AppCfg.Files.OutputDir)

	bl := &BufferLogger{}
	procErr := xmlcreator.CreateXMLFromFileWithLogger(tmp.Name(), s.AppCfg, s.DasipCfg, s.TM, bl)

	after := snapshotOutput(s.AppCfg.Files.OutputDir)
	generated := diffOutput(before, after)

	info, warn, errs := bl.Counts()
	resp := map[string]any{
		"file":            handler.Filename,
		"success":         procErr == nil,
		"logs":            bl.Entries,
		"summary":         map[string]int{"info": info, "warn": warn, "error": errs},
		"files_generated": generated,
	}
	if procErr != nil {
		resp["error"] = procErr.Error()
		s.JSONResponse(w, resp, http.StatusInternalServerError)
		return
	}
	s.JSONResponse(w, resp, http.StatusOK)
}

func (s *Server) handleListOutput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	dir := s.AppCfg.Files.OutputDir
	entries, err := os.ReadDir(dir)
	if err != nil {
		s.ErrorResponse(w, fmt.Sprintf("Error leyendo output: %v", err), http.StatusInternalServerError)
		return
	}
	type item struct {
		Name     string    `json:"name"`
		Size     int64     `json:"size"`
		Modified time.Time `json:"modified"`
	}
	out := make([]item, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".xml") {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, item{Name: e.Name(), Size: fi.Size(), Modified: fi.ModTime()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Modified.After(out[j].Modified) })
	s.JSONResponse(w, out, http.StatusOK)
}

func (s *Server) handleGetOutputFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	name := filepath.Base(r.URL.Query().Get("name"))
	if name == "" || name == "." || !strings.HasSuffix(strings.ToLower(name), ".xml") {
		s.ErrorResponse(w, "Nombre inválido", http.StatusBadRequest)
		return
	}
	full := filepath.Join(s.AppCfg.Files.OutputDir, name)
	data, err := os.ReadFile(full)
	if err != nil {
		s.ErrorResponse(w, "No encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func snapshotOutput(dir string) map[string]time.Time {
	out := make(map[string]time.Time)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".xml") {
			continue
		}
		if fi, err := e.Info(); err == nil {
			out[e.Name()] = fi.ModTime()
		}
	}
	return out
}

func diffOutput(before, after map[string]time.Time) []string {
	var changed []string
	for name, t := range after {
		if old, ok := before[name]; !ok || t.After(old) {
			changed = append(changed, name)
		}
	}
	sort.Strings(changed)
	return changed
}
