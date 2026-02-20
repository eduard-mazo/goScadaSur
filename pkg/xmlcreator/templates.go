// pkg/xmlcreator/templates.go
package xmlcreator

import (
	"encoding/json"
	"fmt"
	"os"
)

// TemplateManager gestiona las plantillas de elementos
type TemplateManager struct {
	templates map[string]ElementDef
	filePath  string
}

// LoadTemplates carga las plantillas de elementos desde un archivo JSON
func LoadTemplates(filePath string) (*TemplateManager, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo de plantillas '%s': %w", filePath, err)
	}

	var db map[string]ElementDef
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("error parseando plantillas JSON: %w", err)
	}

	if len(db) == 0 {
		return nil, fmt.Errorf("el archivo de plantillas está vacío")
	}

	return &TemplateManager{
		templates: db,
		filePath:  filePath,
	}, nil
}

// GetRawJSON retorna el contenido JSON crudo de las plantillas
func (tm *TemplateManager) GetRawJSON() (string, error) {
	if tm == nil || tm.templates == nil {
		return "{}", nil
	}
	data, err := json.MarshalIndent(tm.templates, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveRawJSON actualiza las plantillas desde un string JSON y guarda el archivo
func (tm *TemplateManager) SaveRawJSON(rawJSON string) error {
	var db map[string]ElementDef
	if err := json.Unmarshal([]byte(rawJSON), &db); err != nil {
		return fmt.Errorf("error parseando JSON de plantillas: %w", err)
	}

	// Validar que no esté vacío
	if len(db) == 0 {
		return fmt.Errorf("las plantillas no pueden estar vacías")
	}

	// Guardar en disco
	if err := os.WriteFile(tm.filePath, []byte(rawJSON), 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo de plantillas: %w", err)
	}

	// Actualizar en memoria
	tm.templates = db
	return nil
}

// GetTemplate obtiene una plantilla por su clave
func (tm *TemplateManager) GetTemplate(key string) (ElementDef, bool) {
	if tm == nil || tm.templates == nil {
		return ElementDef{}, false
	}
	template, exists := tm.templates[key]
	return template, exists
}

// DeepCopyElement crea una copia profunda de un ElementDef usando el método Clone
func (tm *TemplateManager) DeepCopyElement(element ElementDef) (ElementDef, error) {
	return element.Clone(), nil
}

// GetTemplateStats retorna estadísticas sobre las plantillas cargadas
func (tm *TemplateManager) GetTemplateStats() map[string]int {
	stats := map[string]int{
		"total":    len(tm.templates),
		"analog":   0,
		"discrete": 0,
		"breaker":  0,
	}

	for _, template := range tm.templates {
		if template.Analog != nil {
			stats["analog"]++
		}
		if template.Discrete != nil {
			stats["discrete"]++
		}
		if template.Breaker != nil {
			stats["breaker"]++
		}
	}

	return stats
}

// ValidateTemplates valida que las plantillas tengan la estructura correcta
func (tm *TemplateManager) ValidateTemplates() []string {
	var warnings []string

	for key, template := range tm.templates {
		// Verificar que al menos un tipo esté definido
		hasType := template.Analog != nil || template.Discrete != nil || template.Breaker != nil

		if !hasType {
			warnings = append(warnings, fmt.Sprintf("plantilla '%s' no tiene ningún tipo definido", key))
			continue
		}

		// Validar Analog
		if template.Analog != nil {
			if template.Analog.Name == "" {
				warnings = append(warnings, fmt.Sprintf("plantilla '%s' (Analog) no tiene nombre", key))
			}
		}

		// Validar Discrete
		if template.Discrete != nil {
			if template.Discrete.Name == "" {
				warnings = append(warnings, fmt.Sprintf("plantilla '%s' (Discrete) no tiene nombre", key))
			}
		}

		// Validar Breaker
		if template.Breaker != nil {
			if template.Breaker.Name == "" {
				warnings = append(warnings, fmt.Sprintf("plantilla '%s' (Breaker) no tiene nombre", key))
			}
		}
	}

	return warnings
}
