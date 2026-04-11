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

type Server struct {
	AppCfg     *config.AppConfig
	DasipCfg   *config.DasipConfig
	TM         *xmlcreator.TemplateManager
	DBClient   *database.DatabaseClient
	PGClient   *database.PostgresClient
	ConfigPath string
}

func NewServer(appCfg *config.AppConfig, dasipCfg *config.DasipConfig, tm *xmlcreator.TemplateManager, configPath string) *Server {
	pgClient, err := database.NewPostgresClient(appCfg)
	if err != nil {
		log.Printf("[WARN] No se pudo conectar a Postgres: %v", err)
	}
	return &Server{
		AppCfg: appCfg, DasipCfg: dasipCfg, TM: tm,
		DBClient: database.NewDatabaseClient(appCfg),
		PGClient: pgClient, ConfigPath: configPath,
	}
}

func (s *Server) Start(port int) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/templates", s.handleTemplates)
	mux.HandleFunc("/api/templates/raw", s.handleRawTemplates)
	mux.HandleFunc("/api/dasip", s.handleDasip)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/query", s.handleQuery)
	mux.HandleFunc("/api/generate", s.handleGenerate)
	mux.HandleFunc("/api/upload", s.handleUpload)

	// Endpoints de output (XMLs generados)
	mux.HandleFunc("/api/output", s.handleListOutput)
	mux.HandleFunc("/api/output/file", s.handleGetOutputFile)

	mux.HandleFunc("/api/auth/register", s.handleRegister)
	mux.HandleFunc("/api/auth/login", s.handleLogin)

	mux.HandleFunc("/api/jobs", s.handleJobs)
	mux.HandleFunc("/api/jobs/history", s.handleJobHistory)
	mux.HandleFunc("/api/jobs/comments", s.handleComments)

	mux.Handle("/", http.FileServer(web.GetFS()))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[INFO] Servidor API iniciado en http://localhost%s", addr)

	handler := s.corsMiddleware(s.AuthMiddleware(mux))
	return http.ListenAndServe(addr, handler)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) JSONResponse(w http.ResponseWriter, data any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[ERROR] encoding JSON: %v", err)
	}
}

func (s *Server) ErrorResponse(w http.ResponseWriter, message string, code int) {
	s.JSONResponse(w, map[string]string{"error": message}, code)
}

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
			s.ErrorResponse(w, fmt.Sprintf("Error guardando: %v", err), http.StatusInternalServerError)
			return
		}
		*s.AppCfg = newCfg
		s.JSONResponse(w, map[string]string{"message": "Configuración actualizada"}, http.StatusOK)
		return
	}
	s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
}

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
	s.JSONResponse(w, map[string]any{"stats": s.TM.GetTemplateStats(), "warnings": s.TM.ValidateTemplates()}, http.StatusOK)
}

func (s *Server) handleRawTemplates(w http.ResponseWriter, r *http.Request) {
	if s.TM == nil {
		s.ErrorResponse(w, "Gestor de plantillas no inicializado", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodGet {
		raw, err := s.TM.GetRawJSON()
		if err != nil {
			s.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
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
			s.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.JSONResponse(w, map[string]string{"message": "Plantillas actualizadas"}, http.StatusOK)
		return
	}
	s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
}

func (s *Server) handleDasip(w http.ResponseWriter, r *http.Request) {
	if s.DasipCfg == nil {
		s.ErrorResponse(w, "DASIP no cargado", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodGet {
		s.JSONResponse(w, s.DasipCfg, http.StatusOK)
		return
	}
	if r.Method == http.MethodPost {
		var nd config.DasipConfig
		if err := json.NewDecoder(r.Body).Decode(&nd); err != nil {
			s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		if nd.DefaultPath == "" {
			nd.DefaultPath = "SCADA/RTU"
		}
		if err := nd.Save(s.AppCfg.GetDasipConfigPath()); err != nil {
			s.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.DasipCfg.DefaultPath = nd.DefaultPath
		s.DasipCfg.DasipMapping = nd.DasipMapping
		s.JSONResponse(w, map[string]string{"message": "DASIP actualizado"}, http.StatusOK)
		return
	}
	s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Host, User, Password, Path, Aor string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	empresa, region, _, _, b3, err := database.ParsePath(req.Path)
	if err != nil {
		s.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := s.DBClient.ExecuteCommand(database.CSharpInput{
		Mode: "station_search", User: req.User, Password: req.Password, Host: req.Host, B3: b3,
	})
	if err != nil {
		s.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.JSONResponse(w, map[string]any{
		"result": result.PayloadJSON, "empresa": empresa, "region": region, "aor": req.Aor, "b3": b3,
	}, http.StatusOK)
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.ErrorResponse(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct{ Host, User, Password, Query string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.ErrorResponse(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	result, err := s.DBClient.ExecuteCommand(database.CSharpInput{
		Mode: "direct_query", User: req.User, Password: req.Password, Host: req.Host, Query: req.Query,
	})
	if err != nil {
		s.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.JSONResponse(w, result.PayloadJSON, http.StatusOK)
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	s.JSONResponse(w, map[string]string{"status": "use /api/upload"}, http.StatusOK)
}
