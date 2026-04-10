package api

import (
	"bytes"
	"encoding/json"
	"goScadaSur/pkg/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleConfig(t *testing.T) {
	appCfg := &config.AppConfig{
		App: config.AppInfo{Name: "TestScada"},
	}
	server := NewServer(appCfg, nil, nil, "")

	// Test GET /api/config
	req, _ := http.NewRequest("GET", "/api/config", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleConfig)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler regresó código incorrecto: obtenido %v esperado %v", status, http.StatusOK)
	}

	var response config.AppConfig
	json.NewDecoder(rr.Body).Decode(&response)
	if response.App.Name != "TestScada" {
		t.Errorf("Respuesta incorrecta: esperado TestScada, obtenido %v", response.App.Name)
	}
}

func TestHandleLogin_NoDB(t *testing.T) {
	server := NewServer(&config.AppConfig{}, nil, nil, "")
	server.PGClient = nil // Forzar modo sin DB

	loginData := map[string]string{
		"username": "admin",
		"password": "password",
	}
	body, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleLogin)

	handler.ServeHTTP(rr, req)

	// Debería fallar con 500 porque PGClient es nil pero estamos intentando loguear
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler regresó código incorrecto: obtenido %v esperado %v", status, http.StatusInternalServerError)
	}
}
