// pkg/fileio/writer.go
package fileio

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
)

// CSVWriter proporciona funcionalidad para escribir archivos CSV
type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
}

// NewCSVWriter crea un nuevo escritor de CSV
func NewCSVWriter(filePath string) (*CSVWriter, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("error creando archivo CSV: %w", err)
	}

	return &CSVWriter{
		file:   file,
		writer: csv.NewWriter(file),
	}, nil
}

// WriteRow escribe una fila en el CSV
func (w *CSVWriter) WriteRow(row []string) error {
	return w.writer.Write(row)
}

// WriteRows escribe múltiples filas en el CSV
func (w *CSVWriter) WriteRows(rows [][]string) error {
	for _, row := range rows {
		if err := w.WriteRow(row); err != nil {
			return err
		}
	}
	return nil
}

// Flush asegura que todos los datos se escriban al archivo
func (w *CSVWriter) Flush() error {
	w.writer.Flush()
	return w.writer.Error()
}

// Close cierra el archivo CSV
func (w *CSVWriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// XMLWriter proporciona funcionalidad para escribir archivos XML
type XMLWriter struct {
	filePath string
	indent   string
}

// NewXMLWriter crea un nuevo escritor de XML
func NewXMLWriter(filePath string, indent string) *XMLWriter {
	return &XMLWriter{
		filePath: filePath,
		indent:   indent,
	}
}

// Write escribe una estructura al archivo XML
func (w *XMLWriter) Write(v interface{}) error {
	var buf bytes.Buffer

	// Escribir header XML
	buf.WriteString(xml.Header)

	// Crear encoder con indentación
	encoder := xml.NewEncoder(&buf)
	if w.indent != "" {
		encoder.Indent("", w.indent)
	}

	// Codificar la estructura
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("error codificando XML: %w", err)
	}

	// Escribir al archivo
	if err := os.WriteFile(w.filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo XML: %w", err)
	}

	return nil
}

// WriteCSVWithHeaders escribe un archivo CSV con cabeceras y datos
func WriteCSVWithHeaders(filePath string, headers []string, data [][]string) error {
	writer, err := NewCSVWriter(filePath)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Escribir cabeceras
	if err := writer.WriteRow(headers); err != nil {
		return fmt.Errorf("error escribiendo cabeceras: %w", err)
	}

	// Escribir datos
	if err := writer.WriteRows(data); err != nil {
		return fmt.Errorf("error escribiendo datos: %w", err)
	}

	return nil
}
