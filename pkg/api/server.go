// pkg/api/server.go
package api

import (
	"encoding/json"
	"fmt"
	"goScadaSur/pkg/config"
	"goScadaSur/pkg/database"
	"goScadaSur/pkg/xmlcreator"
	"goScadaSur/web"
	"log"
	"net/http"
)

// Server representa el servidor de API para la UI
type Server struct {
	AppCfg     *config.AppConfig
	DasipCfg   *config.DasipConfig
	TM         *xmlcreator.TemplateManager
	DBClient   *database.DatabaseClient
	ConfigPath string
}

// NewServer crea una nueva instancia del servidor
func NewServer(appCfg *config.AppConfig, dasipCfg *config.DasipConfig, tm *xmlcreator.TemplateManager, configPath string) *Server {
	return &Server{
		AppCfg:     appCfg,
		DasipCfg:   dasipCfg,
		TM:         tm,
		DBClient:   database.NewDatabaseClient(appCfg),
		ConfigPath: configPath,
	}
}

// Start inicia el servidor HTTP
func (s *Server) Start(port int) error {
	mux := http.NewServeMux()

	// Endpoints de API
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/templates", s.handleTemplates)
	mux.HandleFunc("/api/templates/raw", s.handleRawTemplates)
	mux.HandleFunc("/api/dasip", s.handleDasip)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/query", s.handleQuery)
	mux.HandleFunc("/api/generate", s.handleGenerate)
	
	// Servir archivos estáticos del frontend (React)
	mux.Handle("/", http.FileServer(web.GetFS()))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[INFO] Servidor API iniciado en http://localhost%s", addr)
	
	return http.ListenAndServe(addr, s.corsMiddleware(mux))
}

// corsMiddleware añade cabeceras CORS
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// JSONResponse envía una respuesta JSON
func (s *Server) JSONResponse(w http.ResponseWriter, data any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[ERROR] Error encoding JSON: %v", err)
	}
}

// ErrorResponse envía una respuesta de error JSON
func (s *Server) ErrorResponse(w http.ResponseWriter, message string, code int) {
	s.JSONResponse(w, map[string]string{"error": message}, code)
}

// handleConfig maneja la obtención y actualización de la configuración
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.JSONResponse(w, s.AppCfg, http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		var newCfg config.AppConfig
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		if err := newCfg.Save(s.ConfigPath); err != nil {
			s.ErrorResponse(w, fmt.Sprintf("Error guardando configuración: %v", err), http.StatusInternalServerError)
			return
		}

		// Actualizar en memoria
		*s.AppCfg = newCfg
		s.JSONResponse(w, map[string]string{"message": "Configuración actualizada correctamente"}, http.StatusOK)
		return
	}

	s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
}

// handleTemplates maneja la obtención de estadísticas de plantillas
func (s *Server) handleTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	if s.TM == nil {
		s.JSONResponse(w, map[string]any{
			"stats":    map[string]int{"total": 0, "analog": 0, "discrete": 0, "breaker": 0},
			"warnings": []string{"Plantillas no cargadas"},
		}, http.StatusOK)
		return
	}

	stats := s.TM.GetTemplateStats()
	warnings := s.TM.ValidateTemplates()
	s.JSONResponse(w, map[string]any{"stats": stats, "warnings": warnings}, http.StatusOK)
}

// handleRawTemplates permite leer y guardar el JSON crudo de plantillas
func (s *Server) handleRawTemplates(w http.ResponseWriter, r *http.Request) {
	if s.TM == nil {
		s.ErrorResponse(w, "Gestor de plantillas no inicializado", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		raw, err := s.TM.GetRawJSON()
		if err != nil {
			s.ErrorResponse(w, fmt.Sprintf("Error obteniendo JSON: %v", err), http.StatusInternalServerError)
			return
		}
		s.JSONResponse(w, map[string]string{"raw": raw}, http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Raw string `json:"raw"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		if err := s.TM.SaveRawJSON(req.Raw); err != nil {
			s.ErrorResponse(w, fmt.Sprintf("Error guardando plantillas: %v", err), http.StatusInternalServerError)
			return
		}

		s.JSONResponse(w, map[string]string{"message": "Plantillas actualizadas correctamente"}, http.StatusOK)
		return
	}

	s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
}

// handleDasip permite gestionar el mapeo DASIP
func (s *Server) handleDasip(w http.ResponseWriter, r *http.Request) {
	if s.DasipCfg == nil {
		s.ErrorResponse(w, "Configuración DASIP no cargada", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		s.JSONResponse(w, s.DasipCfg, http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		var newDasip config.DasipConfig
		if err := json.NewDecoder(r.Body).Decode(&newDasip); err != nil {
			s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		if newDasip.DefaultPath == "" {
			newDasip.DefaultPath = "SCADA/RTU"
		}

		path := s.AppCfg.GetDasipConfigPath()
		if err := newDasip.Save(path); err != nil {
			s.ErrorResponse(w, fmt.Sprintf("Error guardando DASIP: %v", err), http.StatusInternalServerError)
			return
		}

		s.DasipCfg.DefaultPath = newDasip.DefaultPath
		s.DasipCfg.DasipMapping = newDasip.DasipMapping

		s.JSONResponse(w, map[string]string{"message": "Configuración DASIP actualizada correctamente"}, http.StatusOK)
		return
	}

	s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
}

// handleSearch maneja la búsqueda de estación
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Host     string `json:"host"`
		User     string `json:"user"`
		Password string `json:"password"`
		Path     string `json:"path"`
		Aor      string `json:"aor"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	empresa, region, _, _, b3, err := database.ParsePath(req.Path)
	if err != nil {
		s.ErrorResponse(w, fmt.Sprintf("Path inválido: %v", err), http.StatusBadRequest)
		return
	}

	input := database.CSharpInput{
		Mode:     "station_search",
		User:     req.User,
		Password: req.Password,
		Host:     req.Host,
		B3:       b3,
	}

	result, err := s.DBClient.ExecuteCommand(input)
	if err != nil {
		s.ErrorResponse(w, fmt.Sprintf("Error en DB: %v", err), http.StatusInternalServerError)
		return
	}

	s.JSONResponse(w, map[string]any{
		"result":  result.PayloadJSON,
		"empresa": empresa,
		"region":  region,
		"aor":     req.Aor,
		"b3":      b3,
	}, http.StatusOK)
}

// handleQuery maneja una consulta SQL directa
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Host     string `json:"host"`
		User     string `json:"user"`
		Password string `json:"password"`
		Query    string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	input := database.CSharpInput{
		Mode:     "direct_query",
		User:     req.User,
		Password: req.Password,
		Host:     req.Host,
		Query:    req.Query,
	}

	result, err := s.DBClient.ExecuteCommand(input)
	if err != nil {
		s.ErrorResponse(w, fmt.Sprintf("Error en DB: %v", err), http.StatusInternalServerError)
		return
	}

	s.JSONResponse(w, result.PayloadJSON, http.StatusOK)
}

// handleGenerate maneja la generación de XML desde el payload JSON
func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	s.JSONResponse(w, map[string]string{"status": "to_be_integrated_with_filesystem"}, http.StatusOK)
}
