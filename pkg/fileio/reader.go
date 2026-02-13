// pkg/fileio/reader.go
package fileio

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// DataReader define la interfaz común para leer datos tabulares
type DataReader interface {
	ReadAll() ([][]string, error)
	Close() error
}

// CSVReader implementa DataReader para archivos CSV
type CSVReader struct {
	file   *os.File
	reader *csv.Reader
}

// ExcelReader implementa DataReader para archivos Excel
type ExcelReader struct {
	file      *excelize.File
	sheetName string
	rows      [][]string
}

// NewDataReader crea un reader apropiado basado en la extensión del archivo
func NewDataReader(filePath string) (DataReader, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".csv":
		return NewCSVReader(filePath)
	case ".xlsx", ".xls":
		return NewExcelReader(filePath)
	default:
		return nil, fmt.Errorf("formato de archivo no soportado: %s (use .csv, .xlsx o .xls)", ext)
	}
}

// NewCSVReader crea un nuevo lector de CSV
func NewCSVReader(filePath string) (*CSVReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo archivo CSV: %w", err)
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Permitir número variable de campos
	reader.TrimLeadingSpace = true

	return &CSVReader{
		file:   file,
		reader: reader,
	}, nil
}

// ReadAll lee todos los registros del CSV
func (r *CSVReader) ReadAll() ([][]string, error) {
	records, err := r.reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error leyendo CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("el archivo CSV debe tener al menos una cabecera y una fila de datos")
	}

	return records, nil
}

// Close cierra el archivo CSV
func (r *CSVReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// NewExcelReader crea un nuevo lector de Excel
func NewExcelReader(filePath string) (*ExcelReader, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo archivo Excel: %w", err)
	}

	// Obtener la primera hoja (o la hoja activa)
	sheetName := file.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("no se encontraron hojas en el archivo Excel")
	}

	return &ExcelReader{
		file:      file,
		sheetName: sheetName,
	}, nil
}

// ReadAll lee todos los registros del Excel
func (r *ExcelReader) ReadAll() ([][]string, error) {
	rows, err := r.file.GetRows(r.sheetName)
	if err != nil {
		return nil, fmt.Errorf("error leyendo hoja Excel '%s': %w", r.sheetName, err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("el archivo Excel debe tener al menos una cabecera y una fila de datos")
	}

	// Normalizar filas (asegurar que todas tengan la misma longitud)
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	normalizedRows := make([][]string, len(rows))
	for i, row := range rows {
		normalizedRows[i] = make([]string, maxCols)
		copy(normalizedRows[i], row)
	}

	r.rows = normalizedRows
	return normalizedRows, nil
}

// Close cierra el archivo Excel
func (r *ExcelReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// ReadData es una función de utilidad que lee datos de cualquier formato soportado
func ReadData(filePath string) (headers []string, data [][]string, headerMap map[string]int, err error) {
	reader, err := NewDataReader(filePath)
	if err != nil {
		return nil, nil, nil, err
	}
	defer reader.Close()

	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, nil, err
	}

	// Primera fila son las cabeceras
	headers = records[0]
	data = records[1:]

	// Crear mapa de índices de cabeceras
	headerMap = make(map[string]int)
	for i, h := range headers {
		headerMap[strings.TrimSpace(h)] = i
	}

	return headers, data, headerMap, nil
}

// ValidateHeaders verifica que las columnas requeridas estén presentes
func ValidateHeaders(headerMap map[string]int, requiredColumns []string) error {
	var missing []string

	for _, col := range requiredColumns {
		if _, exists := headerMap[col]; !exists {
			missing = append(missing, col)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("columnas requeridas faltantes: %v", missing)
	}

	return nil
}

// GetCellValue obtiene un valor de celda de manera segura
func GetCellValue(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

// GetCellValueOrDefault obtiene un valor de celda o retorna un valor por defecto
func GetCellValueOrDefault(row []string, headerMap map[string]int, columnName string, defaultValue string) string {
	if idx, exists := headerMap[columnName]; exists {
		value := GetCellValue(row, idx)
		if value != "" {
			return value
		}
	}
	return defaultValue
}
