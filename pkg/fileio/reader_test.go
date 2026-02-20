package fileio

import (
	"encoding/csv"
	"os"
	"reflect"
	"testing"
)

func TestReadData_CSV(t *testing.T) {
	// Crear CSV temporal
	content := [][]string{
		{"Header1", "Header2"},
		{"Value1", "Value2"},
	}

	tmpFile, err := os.CreateTemp("", "data_*.csv")
	if err != nil {
		t.Fatalf("Error creando archivo temporal: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	writer := csv.NewWriter(tmpFile)
	writer.WriteAll(content)
	writer.Flush()
	tmpFile.Close()

	// Probar ReadData
	headers, data, headerMap, err := ReadData(tmpFile.Name())
	if err != nil {
		t.Fatalf("ReadData falló: %v", err)
	}

	if !reflect.DeepEqual(headers, content[0]) {
		t.Errorf("Headers incorrectos")
	}

	if len(data) != 1 {
		t.Errorf("Esperado 1 fila de datos, obtenido %d", len(data))
	}

	if !reflect.DeepEqual(data[0], content[1]) {
		t.Errorf("Datos incorrectos")
	}

	if headerMap["Header1"] != 0 || headerMap["Header2"] != 1 {
		t.Errorf("Mapa de cabeceras incorrecto")
	}
}
