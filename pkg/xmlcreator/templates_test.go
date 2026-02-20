package xmlcreator

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestLoadTemplates(t *testing.T) {
	// Crear archivo de templates temporal
	templatesContent := `{
		"TEST_ELEM": {
			"Analog": {
				"Name": "TestName",
				"UnitOfMeasure": "V"
			}
		}
	}`
	tmpFile, err := os.CreateTemp("", "templates_*.json")
	if err != nil {
		t.Fatalf("Error creando archivo temporal: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(templatesContent); err != nil {
		t.Fatalf("Error escribiendo archivo temporal: %v", err)
	}
	tmpFile.Close()

	// Probar carga
	tm, err := LoadTemplates(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadTemplates falló: %v", err)
	}

	elem, exists := tm.GetTemplate("TEST_ELEM")
	if !exists {
		t.Fatal("No se encontró el template TEST_ELEM")
	}

	if elem.Analog == nil || elem.Analog.Name != "TestName" {
		t.Errorf("Template cargado incorrectamente")
	}
}

func TestDeepCopyElement(t *testing.T) {
	original := ElementDef{
		Analog: &Analog{
			Name: "Original",
			AnalogValue: &AnalogValue{
				Name: "ValOriginal",
			},
		},
	}

	// Usar el método Clone directamente (o DeepCopyElement si tuviéramos un TM, pero Clone es lo que queremos probar)
	copy := original.Clone()

	// Verificar copia profunda
	if copy.Analog.Name != "Original" {
		t.Errorf("Nombre no copiado")
	}

	// Modificar copia
	copy.Analog.Name = "Modified"
	copy.Analog.AnalogValue.Name = "ValModified"

	// Verificar original intacto
	if original.Analog.Name != "Original" {
		t.Errorf("Original modificado al cambiar copia (Name)")
	}
	if original.Analog.AnalogValue.Name != "ValOriginal" {
		t.Errorf("Original modificado al cambiar copia (AnalogValue.Name)")
	}
}

func TestCloneComparison(t *testing.T) {
	original := ElementDef{
		Analog: &Analog{
			Name: "Original",
			AnalogValue: &AnalogValue{
				Name: "ValOriginal",
			},
		},
	}

	// 1. JSON Copy
	jsonCopy := func(e ElementDef) ElementDef {
		b, _ := json.Marshal(e)
		var c ElementDef
		json.Unmarshal(b, &c)
		return c
	}

	// 2. Manual Clone
	manualCopy := original.Clone()
	jCopy := jsonCopy(original)

	if !reflect.DeepEqual(manualCopy, jCopy) {
		t.Errorf("Clone manual difiere de JSON copy")
	}
}
