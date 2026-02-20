// pkg/xmlcreator/creator.go
package xmlcreator

import (
	"fmt"
	"goScadaSur/pkg/config"
	"goScadaSur/pkg/fileio"
	"log"
	"strings"
)

// CreateXMLFromFile procesa un archivo (CSV o Excel) y genera archivos XML
func CreateXMLFromFile(inputFilePath string, cfg *config.AppConfig, dasipCfg *config.DasipConfig, tm *TemplateManager) error {
	// Leer datos del archivo
	log.Printf("[INFO] Leyendo datos desde: %s", inputFilePath)
	_, dataRows, headerMap, err := fileio.ReadData(inputFilePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo: %w", err)
	}

	if len(dataRows) == 0 {
		log.Println("[WARN] El archivo no contiene datos para procesar")
		return nil
	}

	// Validar columnas requeridas
	if err := fileio.ValidateHeaders(headerMap, cfg.Validation.RequiredColumns); err != nil {
		return fmt.Errorf("validación de columnas fallida: %w", err)
	}

	log.Printf("[OK] Datos leídos correctamente: %d filas", len(dataRows))

	// Procesar las filas
	result, err := processRows(dataRows, headerMap, tm)
	if err != nil {
		return fmt.Errorf("error procesando filas: %w", err)
	}

	// Generar archivos XML
	if err := generateXMLFiles(result, dataRows[0], headerMap, cfg, dasipCfg); err != nil {
		return fmt.Errorf("error generando archivos XML: %w", err)
	}

	return nil
}

// ProcessingResult contiene los resultados del procesamiento de filas
type ProcessingResult struct {
	ElementsIMM  []any
	ElementsIFS  []any
	BreakerName  string
	BreakerLinks []any
}

// processRows procesa todas las filas del archivo de datos
func processRows(dataRows [][]string, headerMap map[string]int, tm *TemplateManager) (*ProcessingResult, error) {
	result := &ProcessingResult{
		ElementsIMM:  make([]any, 0),
		ElementsIFS:  make([]any, 0),
		BreakerLinks: make([]any, 0),
	}

	var cbRowData []string

	// Procesar cada fila
	for rowIdx, row := range dataRows {
		// Validación básica de la fila
		if len(row) == 0 {
			continue
		}

		elementKey := fileio.GetCellValue(row, headerMap["ELEMENT"])
		if elementKey == "" {
			log.Printf("[WARN] Fila %d: ELEMENT vacío, saltando...", rowIdx+2)
			continue
		}

		// Obtener plantilla
		template, isTemplateFound := tm.GetTemplate(elementKey)
		isBreakerType := (isTemplateFound && template.Breaker != nil) || elementKey == "CB"

		// Guardar datos del CB para enlaces posteriores
		if elementKey == "CB" {
			cbRowData = row
		}

		// Generar nombre de visualización
		displayName := generateDisplayName(elementKey, row, headerMap)

		// Procesar elemento IFS
		ifsPoint := createIfsPoint(row, headerMap, displayName, isBreakerType)
		result.ElementsIFS = append(result.ElementsIFS, ifsPoint)

		// Procesar elemento IMM
		if !isTemplateFound {
			log.Printf("[WARN] Plantilla '%s' no encontrada", elementKey)
			continue
		}

		element, err := createIMMElement(template, displayName, row, headerMap, tm)
		if err != nil {
			log.Printf("[WARN] Error procesando elemento '%s': %v", elementKey, err)
			continue
		}

		if element != nil {
			result.ElementsIMM = append(result.ElementsIMM, element)
		}
	}

	// Procesar enlaces de breaker
	if cbRowData != nil {
		result.BreakerLinks, result.BreakerName = createBreakerLinks(cbRowData, dataRows, headerMap)
	}

	return result, nil
}

// generateDisplayName genera el nombre de visualización para un elemento
func generateDisplayName(elementKey string, row []string, headerMap map[string]int) string {
	info := fileio.GetCellValue(row, headerMap["INFO"])
	if info == "MvMoment" {
		return strings.ReplaceAll(elementKey, "_", " ")
	}
	return elementKey
}

// createIfsPoint crea un punto IFS basado en los datos de la fila
func createIfsPoint(row []string, headerMap map[string]int, displayName string, isBreakerType bool) *IfsPoint {
	// Determinar partes del nombre IFS
	var ifsNamePart, ifsPathPart string
	if isBreakerType {
		ifsNamePart = fmt.Sprintf("%s_%s", displayName, displayName)
		ifsPathPart = fmt.Sprintf("%s/%s", displayName, displayName)
	} else {
		ifsNamePart = displayName
		ifsPathPart = displayName
	}

	// Determinar sufijo
	suffix := "M"
	if fileio.GetCellValue(row, headerMap["TYPE"]) == "SP_SC" {
		suffix = "MC"
	}

	// Obtener SBO
	sbo := fileio.GetCellValueOrDefault(row, headerMap, "SBO", "0")

	// Determinar ConType
	conType := "0"
	if fileio.GetCellValue(row, headerMap["TYPE"]) == "SP_SC" {
		conType = "45"
	}

	// Construir nombre del punto IFS
	b1 := fileio.GetCellValue(row, headerMap["B1"])
	b2 := fileio.GetCellValue(row, headerMap["B2"])
	b3 := fileio.GetCellValue(row, headerMap["B3"])
	info := fileio.GetCellValue(row, headerMap["INFO"])

	ifsPointName := fmt.Sprintf("%s_%s_%s_%s_%s_%s", b1, b2, b3, ifsNamePart, info, suffix)

	// Construir PathB
	empresa := fileio.GetCellValue(row, headerMap["EMPRESA"])
	region := fileio.GetCellValue(row, headerMap["REGION"])
	pathB := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s/%s/%s",
		empresa, region, b1, b2, b3, ifsPathPart, info)

	return &IfsPoint{
		Name:          ifsPointName,
		MonAddrHigh:   fileio.GetCellValueOrDefault(row, headerMap, "MHB", "0"),
		MonAddrMiddle: fileio.GetCellValueOrDefault(row, headerMap, "MMB", "0"),
		MonAddrLow:    fileio.GetCellValueOrDefault(row, headerMap, "MLB", "0"),
		MonType:       "0",
		ConAddrHigh:   fileio.GetCellValueOrDefault(row, headerMap, "CHB", "0"),
		ConAddrMiddle: fileio.GetCellValueOrDefault(row, headerMap, "CMB", "0"),
		ConAddrLow:    fileio.GetCellValueOrDefault(row, headerMap, "CLB", "0"),
		ConType:       conType,
		SelectBefore:  sbo,
		Link_IfsPointLinksToInfo: &Link_IfsPointLinksToInfo{
			PathB: pathB,
		},
	}
}

// createIMMElement crea un elemento IMM basado en una plantilla
func createIMMElement(template ElementDef, displayName string, row []string, headerMap map[string]int, tm *TemplateManager) (any, error) {
	// Hacer copia profunda de la plantilla
	instance, err := tm.DeepCopyElement(template)
	if err != nil {
		return nil, err
	}

	aor := fileio.GetCellValue(row, headerMap["AOR"])

	// Configurar según el tipo de elemento
	if instance.Analog != nil {
		instance.Analog.Name = displayName
		instance.Analog.AreaOfResponsibilityId = aor
		return instance.Analog, nil
	}

	if instance.Discrete != nil {
		instance.Discrete.Name = displayName
		instance.Discrete.AreaOfResponsibilityId = aor
		return instance.Discrete, nil
	}

	if instance.Breaker != nil {
		instance.Breaker.Name = displayName
		instance.Breaker.AreaOfResponsibilityId = aor
		if instance.Breaker.Discrete != nil {
			instance.Breaker.Discrete.Name = displayName
			instance.Breaker.Discrete.AreaOfResponsibilityId = aor
		}
		return instance.Breaker, nil
	}

	return nil, nil
}

// createBreakerLinks crea los enlaces del breaker
func createBreakerLinks(cbRow []string, allRows [][]string, headerMap map[string]int) ([]any, string) {
	targetMeasurements := map[string]bool{
		"P":    true,
		"Q":    true,
		"I_S":  true,
		"U_RS": true,
	}

	var links []Link_TerminalMeasuredByMeasurement

	empresa := fileio.GetCellValue(cbRow, headerMap["EMPRESA"])
	region := fileio.GetCellValue(cbRow, headerMap["REGION"])
	b1 := fileio.GetCellValue(cbRow, headerMap["B1"])
	b2 := fileio.GetCellValue(cbRow, headerMap["B2"])
	b3 := fileio.GetCellValue(cbRow, headerMap["B3"])

	basePathB := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s", empresa, region, b1, b2, b3)

	// Buscar mediciones objetivo
	for _, row := range allRows {
		if len(row) <= headerMap["ELEMENT"] {
			continue
		}

		elementKey := fileio.GetCellValue(row, headerMap["ELEMENT"])
		if targetMeasurements[elementKey] {
			displayName := strings.ReplaceAll(elementKey, "_", " ")
			link := Link_TerminalMeasuredByMeasurement{
				PathB: fmt.Sprintf("%s/%s", basePathB, displayName),
			}
			links = append(links, link)
		}
	}

	if len(links) == 0 {
		return nil, ""
	}

	linkedTerminal := LinkedTerminal{
		Name:  "T1",
		Links: links,
	}

	return []any{linkedTerminal}, "CB"
}

// generateXMLFiles genera los archivos XML IFS e IMM
func generateXMLFiles(result *ProcessingResult, firstRow []string, headerMap map[string]int, cfg *config.AppConfig, dasipCfg *config.DasipConfig) error {
	// Obtener valores básicos
	b3 := fileio.GetCellValue(firstRow, headerMap["B3"])
	empresa := fileio.GetCellValue(firstRow, headerMap["EMPRESA"])
	region := fileio.GetCellValue(firstRow, headerMap["REGION"])
	b1 := fileio.GetCellValue(firstRow, headerMap["B1"])
	b2 := fileio.GetCellValue(firstRow, headerMap["B2"])

	// Obtener DASIP y mapear a IFS path
	dasIP := fileio.GetCellValueOrDefault(firstRow, headerMap, "DASIP", "")
	ifsParentPath := dasipCfg.GetIfsParentPath(dasIP)
	log.Printf("[INFO] DASIP '%s' -> %s", dasIP, ifsParentPath)

	// Generar archivo IFS
	if err := generateIFSFile(b3, ifsParentPath, result.ElementsIFS, cfg); err != nil {
		return fmt.Errorf("error generando archivo IFS: %w", err)
	}

	// Generar archivo IMM
	if err := generateIMMFile(b3, empresa, region, b1, b2, result, cfg); err != nil {
		return fmt.Errorf("error generando archivo IMM: %w", err)
	}

	return nil
}

// generateIFSFile genera el archivo XML IFS
func generateIFSFile(b3, parentPath string, elements []any, cfg *config.AppConfig) error {
	if len(elements) == 0 {
		log.Printf("[INFO] No se generará archivo IFS (sin elementos)")
		return nil
	}

	parents := []Parent{{
		Path:     parentPath,
		Elements: elements,
	}}

	fileName := fmt.Sprintf("%s%s", b3, cfg.Output.Suffixes["ifs"])
	return createAndSaveXML(fileName, parents, cfg)
}

// generateIMMFile genera el archivo XML IMM
func generateIMMFile(b3, empresa, region, b1, b2 string, result *ProcessingResult, cfg *config.AppConfig) error {
	if len(result.ElementsIMM) == 0 {
		log.Printf("[INFO] No se generará archivo IMM (sin elementos)")
		return nil
	}

	immParentPath := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s", empresa, region, b1, b2, b3)

	parents := []Parent{{
		Path:     immParentPath,
		Elements: result.ElementsIMM,
	}}

	// Agregar enlaces de breaker si existen
	if len(result.BreakerLinks) > 0 && result.BreakerName != "" {
		parents = append(parents, Parent{
			Path:     fmt.Sprintf("%s/%s", immParentPath, result.BreakerName),
			Elements: result.BreakerLinks,
		})
	}

	fileName := fmt.Sprintf("%s%s", b3, cfg.Output.Suffixes["imm"])
	return createAndSaveXML(fileName, parents, cfg)
}

// createAndSaveXML crea y guarda un archivo XML
func createAndSaveXML(fileName string, parents []Parent, cfg *config.AppConfig) error {
	// Validar que haya contenido
	if len(parents) == 0 || (len(parents[0].Elements) == 0 && (len(parents) == 1 || len(parents[1].Elements) == 0)) {
		log.Printf("[INFO] No se generará '%s' (sin elementos)", fileName)
		return nil
	}

	// Construir estructura XDF
	xdf := XDF{
		Lang:    cfg.XML.Lang,
		Version: cfg.XML.Version,
		Instances: Instances{
			Parents: parents,
		},
	}

	// Escribir XML
	fullPath := cfg.GetOutputPath(fileName)
	writer := fileio.NewXMLWriter(fullPath, cfg.XML.Indent)

	if err := writer.Write(xdf); err != nil {
		return fmt.Errorf("error escribiendo XML '%s': %w", fileName, err)
	}

	log.Printf("[OK] Archivo generado: %s", fileName)
	return nil
}
