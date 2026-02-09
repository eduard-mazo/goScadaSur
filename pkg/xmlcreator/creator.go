// goScadaSur/pkg/xmlcreator/creator.go
package xmlcreator

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
)

// ===================================================================================
// 1. DEFINICIÓN DE ESTRUCTURAS PARA EL XML
// ===================================================================================

type ElementDef struct {
	Analog   *Analog   `json:"Analog,omitempty" xml:"Analog,omitempty"`
	Discrete *Discrete `json:"Discrete,omitempty" xml:"Discrete,omitempty"`
	Breaker  *Breaker  `json:"Breaker,omitempty" xml:"Breaker,omitempty"`
	IfsPoint *IfsPoint `json:"IfsPoint,omitempty" xml:"IfsPoint,omitempty"`
}

type Analog struct {
	XMLName                xml.Name     `json:"-" xml:"Analog"`
	Name                   string       `json:"Name,omitempty" xml:"Name,attr"`
	UnitOfMeasure          string       `json:"UnitOfMeasure,omitempty" xml:"UnitOfMeasure,attr"`
	WeightingSE            string       `json:"WeightingSE,omitempty" xml:"WeightingSE,attr"`
	Multiplier             string       `json:"Multiplier,omitempty" xml:"Multiplier,attr"`
	ElementType            string       `json:"ElementType,omitempty" xml:"ElementType,attr"`
	Phases                 string       `json:"Phases,omitempty" xml:"Phases,attr"`
	ElementName            string       `json:"ElementName,omitempty" xml:"ElementName,attr"`
	MeasurementType        string       `json:"MeasurementType,omitempty" xml:"MeasurementType,attr"`
	AreaOfResponsibilityId string       `json:"AreaOfResponsibilityId,omitempty" xml:"AreaOfResponsibilityId,attr"`
	AnalogValue            *AnalogValue `json:"AnalogValue,omitempty" xml:"AnalogValue,omitempty"`
	AnalogInfo             *AnalogInfo  `json:"AnalogInfo,omitempty" xml:"AnalogInfo,omitempty"`
}

type AnalogValue struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	Archive  string `json:"Archive,omitempty" xml:"Archive,attr"`
	InfoName string `json:"InfoName,omitempty" xml:"InfoName,attr"`
}

type AnalogInfo struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	Value    string `json:"Value,omitempty" xml:"Value,attr"`
	InfoName string `json:"InfoName,omitempty" xml:"InfoName,attr"`
}

type Discrete struct {
	XMLName                xml.Name       `json:"-" xml:"Discrete"`
	Name                   string         `json:"Name,omitempty" xml:"Name,attr"`
	ElementType            string         `json:"ElementType,omitempty" xml:"ElementType,attr"`
	ElementName            string         `json:"ElementName,omitempty" xml:"ElementName,attr"`
	MeasurementType        string         `json:"MeasurementType,omitempty" xml:"MeasurementType,attr"`
	AreaOfResponsibilityId string         `json:"AreaOfResponsibilityId,omitempty" xml:"AreaOfResponsibilityId,attr"`
	DiscreteValue          *DiscreteValue `json:"DiscreteValue,omitempty" xml:"DiscreteValue,omitempty"`
	DiscreteInfo           *DiscreteInfo  `json:"DiscreteInfo,omitempty" xml:"DiscreteInfo,omitempty"`
}

type DiscreteValue struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	InfoName string `json:"InfoName,omitempty" xml:"InfoName,attr"`
}
type DiscreteInfo struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	Value    string `json:"Value,omitempty" xml:"Value,attr"`
	InfoName string `json:"InfoName,omitempty" xml:"InfoName,attr"`
}

type Breaker struct {
	XMLName                xml.Name    `json:"-" xml:"Breaker"`
	Name                   string      `json:"Name,omitempty" xml:"Name,attr"`
	FlowBreakerFlag        string      `json:"FlowBreakerFlag,omitempty" xml:"FlowBreakerFlag,attr"`
	VoltMagLimitCA         string      `json:"VoltMagLimitCA,omitempty" xml:"VoltMagLimitCA,attr"`
	DMSFlag                string      `json:"DMSFlag,omitempty" xml:"DMSFlag,attr"`
	AreaOfResponsibilityId string      `json:"AreaOfResponsibilityId,omitempty" xml:"AreaOfResponsibilityId,attr"`
	Terminals              []*Terminal `json:"Terminals,omitempty" xml:"Terminal,omitempty"`
	Discrete               *Discrete   `json:"Discrete,omitempty" xml:"Discrete,omitempty"`
}

type Terminal struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	EquipEnd string `json:"EquipEnd,omitempty" xml:"EquipEnd,attr"`
}

type Link_TerminalMeasuredByMeasurement struct {
	XMLName xml.Name `xml:"Link_TerminalMeasuredByMeasurement"`
	PathB   string   `xml:"PathB,attr"`
}

type LinkedTerminal struct {
	XMLName xml.Name `xml:"Terminal"`
	Name    string   `xml:"Name,attr"`
	Links   []Link_TerminalMeasuredByMeasurement
}

type XDF struct {
	XMLName   xml.Name  `xml:"XDF"`
	Lang      string    `xml:"xml:lang,attr"`
	Version   string    `xml:"XdfTypeSyntaxVersion,attr"`
	Instances Instances `xml:"Instances"`
}

type Instances struct {
	Parents []Parent `xml:"Parent"`
}

type Parent struct {
	Path     string `xml:"Path,attr"`
	Elements []any  `xml:",any"`
}

type IfsPoint struct {
	XMLName                  xml.Name                  `xml:"IfsPoint"`
	Name                     string                    `xml:"Name,attr"`
	MonAddrHigh              string                    `xml:"MonAddrHigh,attr"`
	MonAddrLow               string                    `xml:"MonAddrLow,attr"`
	MonAddrMiddle            string                    `xml:"MonAddrMiddle,attr"`
	MonType                  string                    `xml:"MonType,attr"`
	ConAddrHigh              string                    `xml:"ConAddrHigh,attr"`
	ConAddrLow               string                    `xml:"ConAddrLow,attr"`
	ConAddrMiddle            string                    `xml:"ConAddrMiddle,attr"`
	ConType                  string                    `xml:"ConType,attr"`
	SelectBefore             string                    `xml:"SelectBefore,attr"`
	Link_IfsPointLinksToInfo *Link_IfsPointLinksToInfo `xml:"Link_IfsPointLinksToInfo,omitempty"`
}

type Link_IfsPointLinksToInfo struct {
	PathB string `xml:"PathB,attr"`
}

// ===================================================================================
// 2. LÓGICA PRINCIPAL DEL PROGRAMA
// ===================================================================================

const (
	xmlLang          = "EN"
	xmlVersion       = "2.0.00"
	templateFilePath = "templates.json"
)

var elementDB map[string]ElementDef

// deepCopy y loadElementDB se mantienen igual
func deepCopy(elementDef ElementDef) (ElementDef, error) {
	var copiedDef ElementDef
	bytes, err := json.Marshal(elementDef)
	if err != nil {
		return copiedDef, fmt.Errorf("error al serializar plantilla para copia: %w", err)
	}
	err = json.Unmarshal(bytes, &copiedDef)
	return copiedDef, err
}

func loadElementDB(filePath string) (map[string]ElementDef, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo de plantillas '%s': %w", filePath, err)
	}

	var db map[string]ElementDef
	if err := json.Unmarshal(fileBytes, &db); err != nil {
		return nil, fmt.Errorf("no se pudo decodificar el JSON de plantillas desde '%s': %w", filePath, err)
	}
	return db, nil
}

func CreateXML(csvFilePath string) error {
	var err error
	elementDB, err = loadElementDB(templateFilePath)
	if err != nil {
		return fmt.Errorf("no se pudieron inicializar las plantillas de elementos: %w", err)
	}

	// 2. Leer el CSV desde el disco
	log.Printf("Leyendo datos desde CSV: %s", csvFilePath)
	dataRows, headerMap, err := readCSV(csvFilePath)
	if err != nil {
		return err
	}

	if len(dataRows) == 0 {
		log.Println("Advertencia: El CSV no contiene datos para procesar.")
		return nil
	}

	// 3. Procesar filas
	elementsForIMM, breakerName, breakerLinks, elementsForIFS := processRows(dataRows, headerMap)

	firstRow := dataRows[0]

	// Aseguramos que existan las columnas clave creadas en main.go
	if _, ok := headerMap["B3"]; !ok {
		return fmt.Errorf("columna B3 no encontrada en CSV")
	}
	if _, ok := headerMap["EMPRESA"]; !ok {
		return fmt.Errorf("columna EMPRESA no encontrada en CSV")
	}
	if _, ok := headerMap["REGION"]; !ok {
		return fmt.Errorf("columna REGION no encontrada en CSV")
	}
	if _, ok := headerMap["B1"]; !ok {
		return fmt.Errorf("columna B1 no encontrada en CSV")
	}
	if _, ok := headerMap["B2"]; !ok {
		return fmt.Errorf("columna B2 no encontrada en CSV")
	}

	b3Value := firstRow[headerMap["B3"]]
	empresa := firstRow[headerMap["EMPRESA"]]
	region := firstRow[headerMap["REGION"]]
	b1 := firstRow[headerMap["B1"]]
	b2 := firstRow[headerMap["B2"]]

	// --- Lógica para el archivo IFS ---
	var ifsParentPath string

	// Validar si existe DASIP, si no, usar default
	dasIPVal := ""
	if idx, ok := headerMap["DASIP"]; ok {
		dasIPVal = firstRow[idx]
	}

	switch dasIPVal {
	case "1":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0133/DASip1"
	case "6":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0135/DASip2"
	case "7":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0137/DASip3"
	case "11":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0139/DASip4"
	case "12":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0141/DASip5"
	case "8":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0143/DASip6"
	case "15":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0145/DASip7"
	case "9":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0147/DASip8"
	case "16":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0173/DASip9"
	case "14":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0152/DASip10"
	case "18":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0155/DASip11"
	case "21":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0157/DASip12"
	case "22":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0189/DASip13"
	case "23":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0191/DASip14"
	case "24":
		ifsParentPath = "PI/IFS/EPM_P1_1/Chan0193/DASip15"
	default:
		log.Printf("Advertencia: Valor de DASIP no reconocido ('%s'). Usando 'SCADA/RTU' como path por defecto.", dasIPVal)
		ifsParentPath = "SCADA/RTU"
	}

	parts := strings.Split(ifsParentPath, "/")
	if len(parts) > 4 {
		log.Printf("Señales pertenecientes a ->\t%s", parts[4])
	}

	fileNameIFS := b3Value + "_IFS.xml"
	parentsForIFS := []Parent{{Path: ifsParentPath, Elements: elementsForIFS}}
	if err := createAndSaveXML(fileNameIFS, parentsForIFS); err != nil {
		return err
	}

	// --- Lógica para el archivo IMM ---
	parentPathIMM := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s",
		empresa,
		region,
		b1,
		b2,
		b3Value,
	)

	mainParentIMM := Parent{Path: parentPathIMM, Elements: elementsForIMM}
	parentsForIMM := []Parent{mainParentIMM}

	if len(breakerLinks) > 0 {
		breakerParentPath := fmt.Sprintf("%s/%s", parentPathIMM, breakerName)
		breakerParent := Parent{
			Path:     breakerParentPath,
			Elements: breakerLinks,
		}
		parentsForIMM = append(parentsForIMM, breakerParent)
	}

	fileNameIMM := b3Value + "_IMM.xml"
	if err := createAndSaveXML(fileNameIMM, parentsForIMM); err != nil {
		return err
	}

	return nil
}

// Nueva función para leer CSV estándar
func readCSV(filePath string) ([][]string, map[string]int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo abrir el archivo CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Permitir número variable de campos si es necesario, aunque idealmente debe ser consistente
	reader.FieldsPerRecord = -1

	rawRows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("error leyendo el contenido del CSV: %w", err)
	}

	if len(rawRows) < 2 {
		return nil, nil, fmt.Errorf("el archivo CSV debe tener al menos cabecera y una fila de datos")
	}

	headers := rawRows[0]
	dataRows := rawRows[1:]

	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	return dataRows, headerMap, nil
}

func processRows(dataRows [][]string, headerMap map[string]int) (elementsForIMM []any, breakerName string, breakerLinks []any, elementsForIFS []any) {
	var cbRowData []string

	// Validaciones de columnas mínimas requeridas para evitar panic
	requiredCols := []string{"ELEMENT", "INFO", "TYPE", "B1", "B2", "B3", "AOR", "EMPRESA", "REGION"}
	for _, col := range requiredCols {
		if _, ok := headerMap[col]; !ok {
			log.Printf("[ERROR CRITICO] Columna requerida '%s' no encontrada en el CSV", col)
			return
		}
	}

	for _, row := range dataRows {
		// Seguridad por si alguna fila está incompleta
		if len(row) <= headerMap["ELEMENT"] {
			continue
		}

		elementKey := row[headerMap["ELEMENT"]]
		template, isTemplateFound := elementDB[elementKey]

		isBreakerType := (isTemplateFound && template.Breaker != nil) || elementKey == "CB"

		if elementKey == "CB" {
			cbRowData = row
		}

		displayName := elementKey
		if row[headerMap["INFO"]] == "MvMoment" {
			displayName = strings.ReplaceAll(elementKey, "_", " ")
		}

		// --- Paso 1: Generación del elemento IFS ---
		var ifsNameElementPart, ifsPathElementPart string
		if isBreakerType {
			ifsNameElementPart = fmt.Sprintf("%s_%s", displayName, displayName)
			ifsPathElementPart = fmt.Sprintf("%s/%s", displayName, displayName)
		} else {
			ifsNameElementPart = displayName
			ifsPathElementPart = displayName
		}

		suffix := "M"
		if row[headerMap["TYPE"]] == "SP_SC" {
			suffix = "MC"
		}

		sbo := "0"
		if idx, ok := headerMap["SBO"]; ok && len(row) > idx {
			if sboValue := row[idx]; sboValue != "" {
				sbo = sboValue
			}
		}

		// Helpers seguros para columnas que podrían no existir o estar vacías
		getVal := func(col string) string {
			if idx, ok := headerMap[col]; ok && len(row) > idx {
				return row[idx]
			}
			return "0" // Default safe value
		}

		conTypeVal := "0"
		if row[headerMap["TYPE"]] == "SP_SC" {
			conTypeVal = "45"
		}

		ifsPoint := &IfsPoint{
			Name: fmt.Sprintf("%s_%s_%s_%s_%s_%s",
				row[headerMap["B1"]], row[headerMap["B2"]], row[headerMap["B3"]],
				ifsNameElementPart, row[headerMap["INFO"]], suffix),
			MonAddrHigh:   getVal("MHB"),
			MonAddrMiddle: getVal("MMB"),
			MonAddrLow:    getVal("MLB"),
			MonType:       "0",
			ConAddrHigh:   getVal("CHB"),
			ConAddrMiddle: getVal("CMB"),
			ConAddrLow:    getVal("CLB"),
			ConType:       conTypeVal,
			SelectBefore:  sbo,
			Link_IfsPointLinksToInfo: &Link_IfsPointLinksToInfo{
				PathB: fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s/%s/%s",
					row[headerMap["EMPRESA"]], row[headerMap["REGION"]], row[headerMap["B1"]],
					row[headerMap["B2"]], row[headerMap["B3"]], ifsPathElementPart, row[headerMap["INFO"]]),
			},
		}
		elementsForIFS = append(elementsForIFS, ifsPoint)

		// --- Paso 2: Generación del elemento IMM ---
		if !isTemplateFound {
			log.Printf("Advertencia: La llave '%s' no fue encontrada en la plantilla...", elementKey)
			continue
		}
		instance, err := deepCopy(template)
		if err != nil {
			log.Printf("Error al copiar plantilla: %v", err)
			continue
		}

		aor := row[headerMap["AOR"]]
		var elementToAppend any
		if instance.Analog != nil {
			instance.Analog.Name = displayName
			instance.Analog.AreaOfResponsibilityId = aor
			elementToAppend = instance.Analog
		} else if instance.Discrete != nil {
			instance.Discrete.Name = displayName
			instance.Discrete.AreaOfResponsibilityId = aor
			elementToAppend = instance.Discrete
		} else if instance.Breaker != nil {
			instance.Breaker.Name = displayName
			instance.Breaker.AreaOfResponsibilityId = aor
			if instance.Breaker.Discrete != nil {
				instance.Breaker.Discrete.Name = displayName
				instance.Breaker.Discrete.AreaOfResponsibilityId = aor
			}
			elementToAppend = instance.Breaker
		}
		if elementToAppend != nil {
			elementsForIMM = append(elementsForIMM, elementToAppend)
		}
	}

	// --- LÓGICA DE ENLACES ---
	if cbRowData != nil {
		targetMeasurements := map[string]bool{"P": true, "Q": true, "I_S": true, "U_RS": true}
		var links []Link_TerminalMeasuredByMeasurement

		basePathB := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s",
			cbRowData[headerMap["EMPRESA"]], cbRowData[headerMap["REGION"]],
			cbRowData[headerMap["B1"]], cbRowData[headerMap["B2"]], cbRowData[headerMap["B3"]])

		for _, row := range dataRows {
			// Check bounds
			if len(row) <= headerMap["ELEMENT"] {
				continue
			}

			elementKey := row[headerMap["ELEMENT"]]
			if targetMeasurements[elementKey] {
				displayName := strings.ReplaceAll(elementKey, "_", " ")
				link := Link_TerminalMeasuredByMeasurement{
					PathB: fmt.Sprintf("%s/%s", basePathB, displayName),
				}
				links = append(links, link)
			}
		}

		if len(links) > 0 {
			linkedTerminal := LinkedTerminal{
				Name:  "T1",
				Links: links,
			}
			breakerLinks = append(breakerLinks, linkedTerminal)
			breakerName = "CB"
		}
	}

	return
}

func createAndSaveXML(fileName string, parents []Parent) error {
	if len(parents) == 0 || (len(parents[0].Elements) == 0 && (len(parents) == 1 || len(parents[1].Elements) == 0)) {
		log.Printf("Información: No se generará el archivo '%s' porque no hay elementos.", fileName)
		return nil
	}

	outputStruct := XDF{
		Lang:    xmlLang,
		Version: xmlVersion,
		Instances: Instances{
			Parents: parents,
		},
	}

	var out bytes.Buffer
	out.WriteString(xml.Header)
	encoder := xml.NewEncoder(&out)
	encoder.Indent("", "    ")

	if err := encoder.Encode(outputStruct); err != nil {
		return fmt.Errorf("error codificando XML '%s': %w", fileName, err)
	}
	if err := os.WriteFile(fileName, out.Bytes(), 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo '%s': %w", fileName, err)
	}

	fmt.Printf("✅ Archivo generado exitosamente: %s\n", fileName)
	return nil
}
