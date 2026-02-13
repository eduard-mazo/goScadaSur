// pkg/xmlcreator/templates.go
package xmlcreator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// templateDB es la base de datos de plantillas de elementos
var templateDB map[string]ElementDef

// LoadTemplates carga las plantillas de elementos desde un archivo JSON
func LoadTemplates(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo de plantillas '%s': %w", filePath, err)
	}

	var db map[string]ElementDef
	if err := json.Unmarshal(data, &db); err != nil {
		return fmt.Errorf("error parseando plantillas JSON: %w", err)
	}

	if len(db) == 0 {
		return fmt.Errorf("el archivo de plantillas está vacío")
	}

	templateDB = db
	log.Printf("[OK] Plantillas cargadas: %d elementos definidos", len(templateDB))
	return nil
}

// GetTemplate obtiene una plantilla por su clave
func GetTemplate(key string) (ElementDef, bool) {
	if templateDB == nil {
		return ElementDef{}, false
	}
	template, exists := templateDB[key]
	return template, exists
}

// DeepCopyElement crea una copia profunda de un ElementDef
func DeepCopyElement(element ElementDef) (ElementDef, error) {
	// Usar JSON para hacer una copia profunda
	data, err := json.Marshal(element)
	if err != nil {
		return ElementDef{}, fmt.Errorf("error serializando elemento: %w", err)
	}

	var copy ElementDef
	if err := json.Unmarshal(data, &copy); err != nil {
		return ElementDef{}, fmt.Errorf("error deserializando elemento: %w", err)
	}

	return copy, nil
}

// GetTemplateStats retorna estadísticas sobre las plantillas cargadas
func GetTemplateStats() map[string]int {
	stats := map[string]int{
		"total":    len(templateDB),
		"analog":   0,
		"discrete": 0,
		"breaker":  0,
	}

	for _, template := range templateDB {
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
func ValidateTemplates() []string {
	var warnings []string

	for key, template := range templateDB {
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
