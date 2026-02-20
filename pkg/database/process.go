// pkg/database/process.go
package database

import (
	"fmt"
	"goScadaSur/pkg/fileio"
	"log"

	"github.com/tidwall/gjson"
)

// SaveDirectQueryToCSV guarda los resultados de una query directa
func SaveDirectQueryToCSV(payloadJSON, filePath string) error {
	writer, err := fileio.NewCSVWriter(filePath)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Obtener columnas
	columnsResult := gjson.Get(payloadJSON, "columns.#.name")
	var headers []string
	for _, name := range columnsResult.Array() {
		headers = append(headers, name.String())
	}

	if err := writer.WriteRow(headers); err != nil {
		return err
	}

	// Escribir datos
	dataResult := gjson.Get(payloadJSON, "data")
	dataResult.ForEach(func(key, row gjson.Result) bool {
		var record []string
		for _, header := range headers {
			value := row.Get(header)
			record = append(record, value.String())
		}
		if err := writer.WriteRow(record); err != nil {
			log.Printf("[ERROR] Error escribiendo fila: %v", err)
		}
		return true
	})

	return nil
}

// SaveStationSearchToCSV guarda los resultados de búsqueda de estación
func SaveStationSearchToCSV(payloadJSON, filePath, empresa, region, aor string) error {
	writer, err := fileio.NewCSVWriter(filePath)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Obtener columnas dinámicas
	columnsResult := gjson.Get(payloadJSON, "columns.#.name")
	var dynamicHeaders []string
	for _, name := range columnsResult.Array() {
		dynamicHeaders = append(dynamicHeaders, name.String())
	}

	// Cabeceras completas
	headers := append([]string{"EMPRESA", "REGION", "AOR"}, dynamicHeaders...)
	if err := writer.WriteRow(headers); err != nil {
		return err
	}

	// Escribir datos
	dataResult := gjson.Get(payloadJSON, "data")
	dataResult.ForEach(func(key, row gjson.Result) bool {
		record := []string{empresa, region, aor}
		for _, header := range dynamicHeaders {
			value := row.Get(header)
			record = append(record, value.String())
		}
		if err := writer.WriteRow(record); err != nil {
			log.Printf("[ERROR] Error escribiendo fila: %v", err)
		}
		return true
	})

	return nil
}

// ParsePath parsea un path en sus componentes
func ParsePath(path string) (empresa, region, b1, b2, b3 string, err error) {
	if path == "" {
		return "", "", "", "", "", fmt.Errorf("path vacío")
	}

	parts := []string{}
	// Usar strings.Split pero manejar diferentes separadores si es necesario
	// Por ahora mantenemos el comportamento original de main.go
	importStrings := "strings"
	_ = importStrings // placeholder for dynamic import analysis

	// We can manually split it
	lastIdx := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			parts = append(parts, path[lastIdx:i])
			lastIdx = i + 1
		}
	}
	parts = append(parts, path[lastIdx:])

	if len(parts) != 5 {
		return "", "", "", "", "", fmt.Errorf("path inválido (esperados 5 partes, encontrados %d)", len(parts))
	}

	return parts[0], parts[1], parts[2], parts[3], parts[4], nil
}
