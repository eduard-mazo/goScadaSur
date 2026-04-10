// pkg/xmlcreator/creator.go
package xmlcreator

import (
	"fmt"
	"goScadaSur/pkg/config"
	"goScadaSur/pkg/fileio"
	"log"
	"strings"
)

// columnAliases mapea nombres alternativos de columnas al nombre canónico esperado.
var columnAliases = map[string]string{
	"M_LB": "MLB", "M_MB": "MMB", "M_HB": "MHB",
	"C_LB": "CLB", "C_MB": "CMB", "C_HB": "CHB",
	"SBE": "SBO",
}

// normalizeHeaderMap aplica los alias para que el resto del código encuentre las columnas.
func normalizeHeaderMap(h map[string]int) map[string]int {
	out := make(map[string]int, len(h))
	for k, v := range h {
		out[k] = v
	}
	for alias, canonical := range columnAliases {
		if idx, ok := h[alias]; ok {
			if _, exists := out[canonical]; !exists {
				out[canonical] = idx
			}
		}
	}
	return out
}

// CreateXMLFromFile lee un archivo (CSV/Excel), agrupa filas por estación
// (EMPRESA/REGION/B1/B2/B3) y genera un par de XML (IFS+IMM) por cada estación.
func CreateXMLFromFile(inputFilePath string, cfg *config.AppConfig, dasipCfg *config.DasipConfig, tm *TemplateManager) error {
	log.Printf("[INFO] Leyendo datos desde: %s", inputFilePath)
	_, dataRows, headerMap, err := fileio.ReadData(inputFilePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo: %w", err)
	}
	if len(dataRows) == 0 {
		log.Println("[WARN] El archivo no contiene datos para procesar")
		return nil
	}

	headerMap = normalizeHeaderMap(headerMap)

	if err := fileio.ValidateHeaders(headerMap, cfg.Validation.RequiredColumns); err != nil {
		return fmt.Errorf("validación de columnas fallida: %w", err)
	}

	// Agrupar filas por estación preservando el orden de aparición.
	groups := make(map[string][][]string)
	order := make([]string, 0)
	for _, row := range dataRows {
		if len(row) == 0 {
			continue
		}
		key := strings.Join([]string{
			fileio.GetCellValue(row, headerMap["EMPRESA"]),
			fileio.GetCellValue(row, headerMap["REGION"]),
			fileio.GetCellValue(row, headerMap["B1"]),
			fileio.GetCellValue(row, headerMap["B2"]),
			fileio.GetCellValue(row, headerMap["B3"]),
		}, "/")
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], row)
	}

	log.Printf("[OK] %d filas leídas, %d estaciones detectadas", len(dataRows), len(order))

	for _, key := range order {
		rows := groups[key]
		log.Printf("[INFO] Procesando estación %s (%d filas)", key, len(rows))
		result, err := processRows(rows, headerMap, tm)
		if err != nil {
			log.Printf("[WARN] estación %s: %v", key, err)
			continue
		}
		if err := generateXMLFiles(result, rows[0], headerMap, cfg, dasipCfg); err != nil {
			log.Printf("[WARN] estación %s: %v", key, err)
		}
	}
	return nil
}

// ====================================================================
// El resto del archivo permanece igual al original.
// ====================================================================

type ProcessingResult struct {
	ElementsIMM  []any
	ElementsIFS  []any
	BreakerName  string
	BreakerLinks []any
}

func processRows(dataRows [][]string, headerMap map[string]int, tm *TemplateManager) (*ProcessingResult, error) {
	result := &ProcessingResult{ElementsIMM: make([]any, 0), ElementsIFS: make([]any, 0), BreakerLinks: make([]any, 0)}
	var cbRowData []string
	for rowIdx, row := range dataRows {
		if len(row) == 0 {
			continue
		}
		elementKey := fileio.GetCellValue(row, headerMap["ELEMENT"])
		if elementKey == "" {
			log.Printf("[WARN] Fila %d: ELEMENT vacío", rowIdx+2)
			continue
		}
		template, isTemplateFound := tm.GetTemplate(elementKey)
		isBreakerType := (isTemplateFound && template.Breaker != nil) || elementKey == "CB"
		if elementKey == "CB" {
			cbRowData = row
		}
		displayName := generateDisplayName(elementKey, row, headerMap)
		result.ElementsIFS = append(result.ElementsIFS, createIfsPoint(row, headerMap, displayName, isBreakerType))
		if !isTemplateFound {
			log.Printf("[WARN] Plantilla '%s' no encontrada", elementKey)
			continue
		}
		element, err := createIMMElement(template, displayName, row, headerMap, tm)
		if err != nil {
			log.Printf("[WARN] elemento '%s': %v", elementKey, err)
			continue
		}
		if element != nil {
			result.ElementsIMM = append(result.ElementsIMM, element)
		}
	}
	if cbRowData != nil {
		result.BreakerLinks, result.BreakerName = createBreakerLinks(cbRowData, dataRows, headerMap)
	}
	return result, nil
}

func generateDisplayName(elementKey string, row []string, headerMap map[string]int) string {
	if fileio.GetCellValue(row, headerMap["INFO"]) == "MvMoment" {
		return strings.ReplaceAll(elementKey, "_", " ")
	}
	return elementKey
}

func createIfsPoint(row []string, headerMap map[string]int, displayName string, isBreakerType bool) *IfsPoint {
	var ifsNamePart, ifsPathPart string
	if isBreakerType {
		ifsNamePart = fmt.Sprintf("%s_%s", displayName, displayName)
		ifsPathPart = fmt.Sprintf("%s/%s", displayName, displayName)
	} else {
		ifsNamePart = displayName
		ifsPathPart = displayName
	}
	suffix := "M"
	if fileio.GetCellValue(row, headerMap["TYPE"]) == "SP_SC" {
		suffix = "MC"
	}
	sbo := fileio.GetCellValueOrDefault(row, headerMap, "SBO", "0")
	conType := "0"
	if fileio.GetCellValue(row, headerMap["TYPE"]) == "SP_SC" {
		conType = "45"
	}
	b1 := fileio.GetCellValue(row, headerMap["B1"])
	b2 := fileio.GetCellValue(row, headerMap["B2"])
	b3 := fileio.GetCellValue(row, headerMap["B3"])
	info := fileio.GetCellValue(row, headerMap["INFO"])
	ifsPointName := fmt.Sprintf("%s_%s_%s_%s_%s_%s", b1, b2, b3, ifsNamePart, info, suffix)
	empresa := fileio.GetCellValue(row, headerMap["EMPRESA"])
	region := fileio.GetCellValue(row, headerMap["REGION"])
	pathB := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s/%s/%s", empresa, region, b1, b2, b3, ifsPathPart, info)
	return &IfsPoint{
		Name:          ifsPointName,
		MonAddrHigh:   fileio.GetCellValueOrDefault(row, headerMap, "MHB", "0"),
		MonAddrMiddle: fileio.GetCellValueOrDefault(row, headerMap, "MMB", "0"),
		MonAddrLow:    fileio.GetCellValueOrDefault(row, headerMap, "MLB", "0"),
		MonType:       "0",
		ConAddrHigh:   fileio.GetCellValueOrDefault(row, headerMap, "CHB", "0"),
		ConAddrMiddle: fileio.GetCellValueOrDefault(row, headerMap, "CMB", "0"),
		ConAddrLow:    fileio.GetCellValueOrDefault(row, headerMap, "CLB", "0"),
		ConType:       conType, SelectBefore: sbo,
		Link_IfsPointLinksToInfo: &Link_IfsPointLinksToInfo{PathB: pathB},
	}
}

func createIMMElement(template ElementDef, displayName string, row []string, headerMap map[string]int, tm *TemplateManager) (any, error) {
	instance, err := tm.DeepCopyElement(template)
	if err != nil {
		return nil, err
	}
	aor := fileio.GetCellValue(row, headerMap["AOR"])
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

func createBreakerLinks(cbRow []string, allRows [][]string, headerMap map[string]int) ([]any, string) {
	targetMeasurements := map[string]bool{"P": true, "Q": true, "I_S": true, "U_RS": true}
	var links []Link_TerminalMeasuredByMeasurement
	empresa := fileio.GetCellValue(cbRow, headerMap["EMPRESA"])
	region := fileio.GetCellValue(cbRow, headerMap["REGION"])
	b1 := fileio.GetCellValue(cbRow, headerMap["B1"])
	b2 := fileio.GetCellValue(cbRow, headerMap["B2"])
	b3 := fileio.GetCellValue(cbRow, headerMap["B3"])
	basePathB := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s", empresa, region, b1, b2, b3)
	for _, row := range allRows {
		if len(row) <= headerMap["ELEMENT"] {
			continue
		}
		elementKey := fileio.GetCellValue(row, headerMap["ELEMENT"])
		if targetMeasurements[elementKey] {
			displayName := strings.ReplaceAll(elementKey, "_", " ")
			links = append(links, Link_TerminalMeasuredByMeasurement{PathB: fmt.Sprintf("%s/%s", basePathB, displayName)})
		}
	}
	if len(links) == 0 {
		return nil, ""
	}
	return []any{LinkedTerminal{Name: "T1", Links: links}}, "CB"
}

func generateXMLFiles(result *ProcessingResult, firstRow []string, headerMap map[string]int, cfg *config.AppConfig, dasipCfg *config.DasipConfig) error {
	b3 := fileio.GetCellValue(firstRow, headerMap["B3"])
	empresa := fileio.GetCellValue(firstRow, headerMap["EMPRESA"])
	region := fileio.GetCellValue(firstRow, headerMap["REGION"])
	b1 := fileio.GetCellValue(firstRow, headerMap["B1"])
	b2 := fileio.GetCellValue(firstRow, headerMap["B2"])
	dasIP := fileio.GetCellValueOrDefault(firstRow, headerMap, "DASIP", "")
	ifsParentPath := dasipCfg.GetIfsParentPath(dasIP)
	if err := generateIFSFile(b3, ifsParentPath, result.ElementsIFS, cfg); err != nil {
		return err
	}
	return generateIMMFile(b3, empresa, region, b1, b2, result, cfg)
}

func generateIFSFile(b3, parentPath string, elements []any, cfg *config.AppConfig) error {
	if len(elements) == 0 {
		return nil
	}
	parents := []Parent{{Path: parentPath, Elements: elements}}
	return createAndSaveXML(fmt.Sprintf("%s%s", b3, cfg.Output.Suffixes["ifs"]), parents, cfg)
}

func generateIMMFile(b3, empresa, region, b1, b2 string, result *ProcessingResult, cfg *config.AppConfig) error {
	if len(result.ElementsIMM) == 0 {
		return nil
	}
	immParentPath := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s", empresa, region, b1, b2, b3)
	parents := []Parent{{Path: immParentPath, Elements: result.ElementsIMM}}
	if len(result.BreakerLinks) > 0 && result.BreakerName != "" {
		parents = append(parents, Parent{Path: fmt.Sprintf("%s/%s", immParentPath, result.BreakerName), Elements: result.BreakerLinks})
	}
	return createAndSaveXML(fmt.Sprintf("%s%s", b3, cfg.Output.Suffixes["imm"]), parents, cfg)
}

func createAndSaveXML(fileName string, parents []Parent, cfg *config.AppConfig) error {
	xdf := XDF{Lang: cfg.XML.Lang, Version: cfg.XML.Version, Instances: Instances{Parents: parents}}
	fullPath := cfg.GetOutputPath(fileName)
	if err := fileio.NewXMLWriter(fullPath, cfg.XML.Indent).Write(xdf); err != nil {
		return fmt.Errorf("error escribiendo XML '%s': %w", fileName, err)
	}
	log.Printf("[OK] Archivo generado: %s", fileName)
	return nil
}
