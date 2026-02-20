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
	App        AppInfo           `yaml:"app" json:"app"`
	Files      FilesConfig       `yaml:"files" json:"files"`
	XML        XMLConfig         `yaml:"xml" json:"xml"`
	Logging    LoggingConfig     `yaml:"logging" json:"logging"`
	Database   DatabaseConfig    `yaml:"database" json:"database"`
	Output     OutputConfig      `yaml:"output" json:"output"`
	Validation ValidationConfig  `yaml:"validation" json:"validation"`
	Processing ProcessingConfig  `yaml:"processing" json:"processing"`
}

type AppInfo struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description" json:"description"`
}

type FilesConfig struct {
	Templates              string   `yaml:"templates" json:"templates"`
	DasipMapping           string   `yaml:"dasip_mapping" json:"dasip_mapping"`
	OutputDir              string   `yaml:"output_dir" json:"output_dir"`
	SupportedInputFormats  []string `yaml:"supported_input_formats" json:"supported_input_formats"`
}

type XMLConfig struct {
	Lang    string `yaml:"lang" json:"lang"`
	Version string `yaml:"version" json:"version"`
	Indent  string `yaml:"indent" json:"indent"`
}

type LoggingConfig struct {
	Level           string `yaml:"level" json:"level"`
	TimestampFormat string `yaml:"timestamp_format" json:"timestamp_format"`
}

type DatabaseConfig struct {
	ConnectionTimeout int    `yaml:"connection_timeout" json:"connection_timeout"`
	CSharpExecutable  string `yaml:"csharp_executable" json:"csharp_executable"`
}

type OutputConfig struct {
	TimestampFormat string            `yaml:"timestamp_format" json:"timestamp_format"`
	Suffixes        map[string]string `yaml:"suffixes" json:"suffixes"`
}

type ValidationConfig struct {
	RequiredColumns []string `yaml:"required_columns" json:"required_columns"`
	OptionalColumns []string `yaml:"optional_columns" json:"optional_columns"`
}

type ProcessingConfig struct {
	ParallelEnabled bool `yaml:"parallel_enabled" json:"parallel_enabled"`
	MaxWorkers      int  `yaml:"max_workers" json:"max_workers"`
	BufferSize      int  `yaml:"buffer_size" json:"buffer_size"`
}

// DasipConfig contiene la configuración del mapeo DASIP
type DasipConfig struct {
	DefaultPath  string            `yaml:"default_path" json:"default_path"`
	DasipMapping map[string]string `yaml:"dasip_mapping" json:"dasip_mapping"`
}

var (
	// No global variables to ensure thread safety and testability
)

// Load carga la configuración principal desde un archivo YAML
func Load(configPath string) (*AppConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo de configuración '%s': %w", configPath, err)
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parseando configuración YAML: %w", err)
	}

	// Aplicar valores por defecto si no están configurados
	cfg.applyDefaults()
	
	// Validar configuración
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	return &cfg, nil
}

// LoadDasipConfig carga la configuración de DASIP desde un archivo YAML
func LoadDasipConfig(configPath string) (*DasipConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo de configuración DASIP '%s': %w", configPath, err)
	}

	var cfg DasipConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parseando configuración DASIP: %w", err)
	}

	// Validar configuración DASIP
	if cfg.DefaultPath == "" {
		cfg.DefaultPath = "SCADA/RTU"
	}

	if len(cfg.DasipMapping) == 0 {
		return nil, fmt.Errorf("el mapeo de DASIP está vacío")
	}

	return &cfg, nil
}

// Save guarda la configuración principal en un archivo YAML
func (cfg *AppConfig) Save(configPath string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error serializando configuración principal: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo de configuración principal '%s': %w", configPath, err)
	}

	return nil
}

// applyDefaults aplica valores por defecto a la configuración
func (cfg *AppConfig) applyDefaults() {
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
func (cfg *AppConfig) validate() error {
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

// Save guarda la configuración de DASIP en un archivo YAML
func (d *DasipConfig) Save(configPath string) error {
	data, err := yaml.Marshal(d)
	if err != nil {
		return fmt.Errorf("error serializando configuración DASIP: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo de configuración DASIP '%s': %w", configPath, err)
	}

	return nil
}

// GetIfsParentPath retorna el path IFS basado en el valor de DASIP
func (d *DasipConfig) GetIfsParentPath(dasIPVal string) string {
	if d == nil {
		return "SCADA/RTU" // Fallback
	}

	if path, exists := d.DasipMapping[dasIPVal]; exists {
		return path
	}

	return d.DefaultPath
}

// EnsureOutputDir asegura que el directorio de salida exista
func (cfg *AppConfig) EnsureOutputDir() error {
	if err := os.MkdirAll(cfg.Files.OutputDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de salida: %w", err)
	}

	return nil
}

// GetOutputPath construye la ruta completa de un archivo de salida
func (cfg *AppConfig) GetOutputPath(filename string) string {
	return filepath.Join(cfg.Files.OutputDir, filename)
}

// IsFormatSupported verifica si un formato de archivo está soportado
func (cfg *AppConfig) IsFormatSupported(format string) bool {
	for _, supported := range cfg.Files.SupportedInputFormats {
		if supported == format {
			return true
		}
	}
	return false
}

// GetTemplatesPath retorna la ruta al archivo de templates
func (cfg *AppConfig) GetTemplatesPath() string {
	return cfg.Files.Templates
}

// GetDasipConfigPath retorna la ruta al archivo de configuración DASIP
func (cfg *AppConfig) GetDasipConfigPath() string {
	return cfg.Files.DasipMapping
}
