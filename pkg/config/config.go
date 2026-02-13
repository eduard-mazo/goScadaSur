// pkg/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// AppConfig representa la configuración completa de la aplicación
type AppConfig struct {
	App        AppInfo           `yaml:"app"`
	Files      FilesConfig       `yaml:"files"`
	XML        XMLConfig         `yaml:"xml"`
	Logging    LoggingConfig     `yaml:"logging"`
	Database   DatabaseConfig    `yaml:"database"`
	Output     OutputConfig      `yaml:"output"`
	Validation ValidationConfig  `yaml:"validation"`
	Processing ProcessingConfig  `yaml:"processing"`
}

type AppInfo struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type FilesConfig struct {
	Templates              string   `yaml:"templates"`
	DasipMapping           string   `yaml:"dasip_mapping"`
	OutputDir              string   `yaml:"output_dir"`
	SupportedInputFormats  []string `yaml:"supported_input_formats"`
}

type XMLConfig struct {
	Lang    string `yaml:"lang"`
	Version string `yaml:"version"`
	Indent  string `yaml:"indent"`
}

type LoggingConfig struct {
	Level           string `yaml:"level"`
	TimestampFormat string `yaml:"timestamp_format"`
}

type DatabaseConfig struct {
	ConnectionTimeout int    `yaml:"connection_timeout"`
	CSharpExecutable  string `yaml:"csharp_executable"`
}

type OutputConfig struct {
	TimestampFormat string            `yaml:"timestamp_format"`
	Suffixes        map[string]string `yaml:"suffixes"`
}

type ValidationConfig struct {
	RequiredColumns []string `yaml:"required_columns"`
	OptionalColumns []string `yaml:"optional_columns"`
}

type ProcessingConfig struct {
	ParallelEnabled bool `yaml:"parallel_enabled"`
	MaxWorkers      int  `yaml:"max_workers"`
	BufferSize      int  `yaml:"buffer_size"`
}

// DasipConfig contiene la configuración del mapeo DASIP
type DasipConfig struct {
	DefaultPath  string            `yaml:"default_path"`
	DasipMapping map[string]string `yaml:"dasip_mapping"`
}

var (
	// Global representa la configuración global de la aplicación
	Global *AppConfig
	
	// Dasip representa la configuración de mapeo DASIP
	Dasip *DasipConfig
)

// Load carga la configuración principal desde un archivo YAML
func Load(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo de configuración '%s': %w", configPath, err)
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parseando configuración YAML: %w", err)
	}

	// Aplicar valores por defecto si no están configurados
	applyDefaults(&cfg)
	
	// Validar configuración
	if err := validate(&cfg); err != nil {
		return fmt.Errorf("configuración inválida: %w", err)
	}

	Global = &cfg
	return nil
}

// LoadDasipConfig carga la configuración de DASIP desde un archivo YAML
func LoadDasipConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo de configuración DASIP '%s': %w", configPath, err)
	}

	var cfg DasipConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parseando configuración DASIP: %w", err)
	}

	// Validar configuración DASIP
	if cfg.DefaultPath == "" {
		cfg.DefaultPath = "SCADA/RTU"
	}

	if len(cfg.DasipMapping) == 0 {
		return fmt.Errorf("el mapeo de DASIP está vacío")
	}

	Dasip = &cfg
	return nil
}

// applyDefaults aplica valores por defecto a la configuración
func applyDefaults(cfg *AppConfig) {
	// Configuración de procesamiento
	if cfg.Processing.MaxWorkers == 0 {
		cfg.Processing.MaxWorkers = runtime.NumCPU()
	}
	
	if cfg.Processing.BufferSize == 0 {
		cfg.Processing.BufferSize = 8192
	}

	// Directorio de salida
	if cfg.Files.OutputDir == "" {
		cfg.Files.OutputDir = "output"
	}

	// Formato de timestamp
	if cfg.Output.TimestampFormat == "" {
		cfg.Output.TimestampFormat = "20060102_150405"
	}

	// Nivel de logging
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
}

// validate valida la configuración cargada
func validate(cfg *AppConfig) error {
	// Validar que existan archivos críticos
	if cfg.Files.Templates == "" {
		return fmt.Errorf("ruta de templates no especificada")
	}

	if cfg.Files.DasipMapping == "" {
		return fmt.Errorf("ruta de configuración DASIP no especificada")
	}

	// Validar XML
	if cfg.XML.Lang == "" {
		return fmt.Errorf("idioma XML no especificado")
	}

	if cfg.XML.Version == "" {
		return fmt.Errorf("versión XML no especificada")
	}

	return nil
}

// GetIfsParentPath retorna el path IFS basado en el valor de DASIP
func GetIfsParentPath(dasIPVal string) string {
	if Dasip == nil {
		return "SCADA/RTU" // Fallback
	}

	if path, exists := Dasip.DasipMapping[dasIPVal]; exists {
		return path
	}

	return Dasip.DefaultPath
}

// EnsureOutputDir asegura que el directorio de salida exista
func EnsureOutputDir() error {
	if Global == nil {
		return fmt.Errorf("configuración no cargada")
	}

	if err := os.MkdirAll(Global.Files.OutputDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de salida: %w", err)
	}

	return nil
}

// GetOutputPath construye la ruta completa de un archivo de salida
func GetOutputPath(filename string) string {
	if Global == nil {
		return filename
	}
	return filepath.Join(Global.Files.OutputDir, filename)
}

// IsFormatSupported verifica si un formato de archivo está soportado
func IsFormatSupported(format string) bool {
	if Global == nil {
		return false
	}

	for _, supported := range Global.Files.SupportedInputFormats {
		if supported == format {
			return true
		}
	}
	return false
}

// GetTemplatesPath retorna la ruta al archivo de templates
func GetTemplatesPath() string {
	if Global == nil {
		return "templates.json"
	}
	return Global.Files.Templates
}

// GetDasipConfigPath retorna la ruta al archivo de configuración DASIP
func GetDasipConfigPath() string {
	if Global == nil {
		return "dasip_config.yaml"
	}
	return Global.Files.DasipMapping
}
