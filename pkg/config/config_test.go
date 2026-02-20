package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Crear archivo de configuración temporal
	configContent := `
app:
  name: "TestApp"
  version: "1.0.0"
files:
  templates: "templates.json"
  dasip_mapping: "dasip_config.yaml"
  output_dir: "output"
xml:
  lang: "es"
  version: "1.0"
`
	tmpFile, err := os.CreateTemp("", "config_*.yaml")
	if err != nil {
		t.Fatalf("Error creando archivo temporal: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Error escribiendo archivo temporal: %v", err)
	}
	tmpFile.Close()

	// Probar carga
	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load falló: %v", err)
	}

	if cfg.App.Name != "TestApp" {
		t.Errorf("Esperado Name='TestApp', obtenido '%s'", cfg.App.Name)
	}
}

func TestLoadDasipConfig(t *testing.T) {
	configContent := `
default_path: "DEFAULT/PATH"
dasip_mapping:
  "10.0.0.1": "PATH/ONE"
`
	tmpFile, err := os.CreateTemp("", "dasip_*.yaml")
	if err != nil {
		t.Fatalf("Error creando archivo temporal: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Error escribiendo archivo temporal: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadDasipConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadDasipConfig falló: %v", err)
	}

	if cfg.DefaultPath != "DEFAULT/PATH" {
		t.Errorf("Esperado DefaultPath='DEFAULT/PATH', obtenido '%s'", cfg.DefaultPath)
	}

	if val, ok := cfg.DasipMapping["10.0.0.1"]; !ok || val != "PATH/ONE" {
		t.Errorf("Mapeo incorrecto para 10.0.0.1")
	}
}
