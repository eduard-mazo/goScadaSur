// pkg/api/handlers.go
package api

import (
	"encoding/json"
	"goScadaSur/pkg/database"
	"net/http"
	"time"
)

// handleRegister maneja el registro de nuevos usuarios
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if s.PGClient == nil {
		s.ErrorResponse(w, "Base de datos Postgres no configurada", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := database.HashPassword(req.Password)
	if err != nil {
		s.ErrorResponse(w, "Error procesando contraseña", http.StatusInternalServerError)
		return
	}

	user := database.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		FullName: req.FullName,
	}

	if err := s.PGClient.DB.Create(&user).Error; err != nil {
		s.ErrorResponse(w, "Error creando usuario (puede que ya exista)", http.StatusConflict)
		return
	}

	s.JSONResponse(w, map[string]string{"message": "Usuario registrado correctamente"}, http.StatusCreated)
}

// handleLogin maneja la autenticación de usuarios
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if s.PGClient == nil {
		s.ErrorResponse(w, "Base de datos Postgres no configurada", http.StatusInternalServerError)
		return
	}

	var user database.User
	if err := s.PGClient.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		s.ErrorResponse(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		return
	}

	if !database.CheckPasswordHash(req.Password, user.Password) {
		s.ErrorResponse(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		return
	}

	token, err := GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		s.ErrorResponse(w, "Error generando token", http.StatusInternalServerError)
		return
	}

	s.JSONResponse(w, map[string]string{
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
	}, http.StatusOK)
}

// handleJobs gestiona la creación y listado de trabajos
func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	if s.PGClient == nil {
		s.ErrorResponse(w, "Base de datos Postgres no configurada", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var jobs []database.Job
		s.PGClient.DB.Preload("CreatedBy").Preload("Comments").Find(&jobs)
		s.JSONResponse(w, jobs, http.StatusOK)

	case http.MethodPost:
		var job database.Job
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// En una implementación real, extraeríamos el UserID del token JWT (inyectado en el contexto)
		// Por ahora, asumimos un ID de prueba o dejamos que el frontend lo envíe (menos seguro)
		if err := s.PGClient.DB.Create(&job).Error; err != nil {
			s.ErrorResponse(w, "Error creando Job", http.StatusInternalServerError)
			return
		}

		// Registrar historia inicial
		s.PGClient.DB.Create(&database.JobHistory{
			JobID:     job.ID,
			UserID:    job.CreatedByID,
			NewStatus: database.StatusPending,
			Action:    "Job Created",
			CreatedAt: time.Now(),
		})

		s.JSONResponse(w, job, http.StatusCreated)

	case http.MethodPatch:
		var req struct {
			ID     uint               `json:"id"`
			Status database.JobStatus `json:"status"`
			UserID uint               `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		var job database.Job
		if err := s.PGClient.DB.First(&job, req.ID).Error; err != nil {
			s.ErrorResponse(w, "Job no encontrado", http.StatusNotFound)
			return
		}

		oldStatus := job.Status
		job.Status = req.Status

		if err := s.PGClient.DB.Save(&job).Error; err != nil {
			s.ErrorResponse(w, "Error actualizando Job", http.StatusInternalServerError)
			return
		}

		// Registrar cambio en historia
		s.PGClient.DB.Create(&database.JobHistory{
			JobID:     job.ID,
			UserID:    req.UserID,
			OldStatus: oldStatus,
			NewStatus: req.Status,
			Action:    "Status Updated",
			CreatedAt: time.Now(),
		})

		s.JSONResponse(w, job, http.StatusOK)

	default:
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// handleJobHistory retorna la historia de un trabajo específico
func (s *Server) handleJobHistory(w http.ResponseWriter, r *http.Request) {
	if s.PGClient == nil {
		s.ErrorResponse(w, "Base de datos Postgres no configurada", http.StatusInternalServerError)
		return
	}

	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		s.ErrorResponse(w, "ID de Job requerido", http.StatusBadRequest)
		return
	}

	var history []database.JobHistory
	s.PGClient.DB.Where("job_id = ?", jobID).Order("created_at desc").Find(&history)
	s.JSONResponse(w, history, http.StatusOK)
}

// handleComments permite añadir comentarios a un Job
func (s *Server) handleComments(w http.ResponseWriter, r *http.Request) {
	if s.PGClient == nil {
		s.ErrorResponse(w, "Base de datos Postgres no configurada", http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var comment database.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := s.PGClient.DB.Create(&comment).Error; err != nil {
		s.ErrorResponse(w, "Error guardando comentario", http.StatusInternalServerError)
		return
	}

	s.JSONResponse(w, comment, http.StatusCreated)
}
