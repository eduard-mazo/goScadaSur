// goScadaSur/pkg/xmlcreator/creator.go
package xmlcreator

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

// ===================================================================================
// 1. DEFINICIÓN DE ESTRUCTURAS PARA EL XML (Sin cambios)
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
	ElementType            string       `json:"ElementType,omitempty" xml:"ElementType,attr"`
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

type XDF struct {
	XMLName   xml.Name  `xml:"XDF"`
	Lang      string    `xml:"xml:lang,attr"`
	Version   string    `xml:"XdfTypeSyntaxVersion,attr"`
	Instances Instances `xml:"Instances"`
}

type Instances struct {
	Parent Parent `xml:"Parent"`
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
// 2. LÓGICA PRINCIPAL DEL PROGRAMA (Refactorizada)
// ===================================================================================

const (
	xmlLang          = "EN"
	xmlVersion       = "2.0.00"
	templateFilePath = "templates.json" // Nombre del archivo de plantillas
)

// Variable global para almacenar las plantillas cargadas
var elementDB map[string]ElementDef

// deepCopy crea una copia profunda de un ElementDef para evitar modificar la plantilla original.
func deepCopy(elementDef ElementDef) (ElementDef, error) {
	var copiedDef ElementDef
	bytes, err := json.Marshal(elementDef)
	if err != nil {
		return copiedDef, fmt.Errorf("error al serializar plantilla para copia: %w", err)
	}
	err = json.Unmarshal(bytes, &copiedDef)
	return copiedDef, err
}

// loadElementDB carga las plantillas de elementos desde un archivo JSON.
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

func CreateXML(payloadJSON, empresa, region, aor string) error {
	// --- 1. Cargar plantillas de elementos desde archivo externo ---
	var err error
	elementDB, err = loadElementDB(templateFilePath)
	if err != nil {
		return fmt.Errorf("no se pudieron inicializar las plantillas de elementos: %w", err)
	}

	// --- 2. Decodificar JSON para capturar Header y Rows ---
	dataRows, headerMap, err := readCSVFromJSON(payloadJSON, empresa, region, aor)
	if err != nil {
		return err
	}

	if len(dataRows) == 0 {
		log.Println("Advertencia: No se encontraron datos en el payload JSON para procesar.")
		return nil
	}

	// --- 3. Procesar filas para generar los elementos de ambos XML ---
	elementsForIMM, elementsForIFS := processRows(dataRows, headerMap)

	// --- 4. Ensamblar y guardar los archivos XML ---
	firstRow := dataRows[0]
	b3Value := firstRow[headerMap["B3"]]

	// --- 4a. Determinar el Path para el archivo IFS dinámicamente ---
	var ifsParentPath string
	dasIPValue := firstRow[headerMap["DASIP"]]
	switch dasIPValue {
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
	default:
		// Fallback por si el valor no es el esperado
		log.Printf("Advertencia: Valor de DASIP no reconocido ('%s'). Usando 'SCADA/RTU' como path por defecto.", dasIPValue)
		ifsParentPath = "SCADA/RTU"
	}

	parts := strings.Split(ifsParentPath, "/")
	log.Printf("Señales pertenecientes a ->\t%s", parts[4])

	// --- 4b. Generar Path para el archivo IMM ---
	parentPathIMM := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s",
		firstRow[headerMap["EMPRESA"]],
		firstRow[headerMap["REGION"]],
		firstRow[headerMap["B1"]],
		firstRow[headerMap["B2"]],
		b3Value,
	)

	// Crear y guardar el archivo IMM
	fileNameIMM := b3Value + "_IMM.xml"
	if err := createAndSaveXML(fileNameIMM, parentPathIMM, elementsForIMM); err != nil {
		return err
	}

	// Crear y guardar el archivo IFS usando el path dinámico
	fileNameIFS := b3Value + "_IFS.xml"
	if err := createAndSaveXML(fileNameIFS, ifsParentPath, elementsForIFS); err != nil {
		return err
	}

	return nil
}

func readCSVFromJSON(payloadJSON, empresa, region, aor string) ([][]string, map[string]int, error) {
	// (Esta función no requiere cambios)
	if !gjson.Valid(payloadJSON) {
		return nil, nil, errors.New("payloadJSON no es un JSON válido")
	}

	constantHeaders := []string{"EMPRESA", "REGION", "AOR"}
	dynamicHeadersResult := gjson.Get(payloadJSON, "columns.#.name")

	if !dynamicHeadersResult.Exists() || !dynamicHeadersResult.IsArray() {
		return nil, nil, errors.New("el campo 'columns' no se encontró o no es un array válido en el JSON")
	}

	var dynamicHeaders []string
	for _, nameResult := range dynamicHeadersResult.Array() {
		dynamicHeaders = append(dynamicHeaders, nameResult.String())
	}

	headers := append(constantHeaders, dynamicHeaders...)

	headerMap := make(map[string]int)
	for i, name := range headers {
		headerMap[name] = i
	}

	dataResult := gjson.Get(payloadJSON, "data")
	if !dataResult.Exists() || !dataResult.IsArray() {
		return nil, nil, errors.New("el campo 'data' no se encontró o no es un array válido en el JSON")
	}

	var records [][]string
	dataResult.ForEach(func(key, row gjson.Result) bool {
		if !row.IsObject() {
			log.Printf("Advertencia: Se encontró una fila no-objeto. Saltando.")
			return true
		}
		record := []string{empresa, region, aor}

		for _, header := range dynamicHeaders {
			value := row.Get(header)
			record = append(record, value.String())
		}
		records = append(records, record)
		return true
	})

	return records, headerMap, nil
}

// processRows itera sobre los registros para generar las listas de elementos para los archivos IMM e IFS.
func processRows(dataRows [][]string, headerMap map[string]int) (elementsForIMM []any, elementsForIFS []any) {
	for _, row := range dataRows {
		// Usar la clave original para búsquedas en el mapa de plantillas
		elementKey := row[headerMap["ELEMENT"]]
		template, isTemplateFound := elementDB[elementKey]
		isBreaker := isTemplateFound && template.Breaker != nil

		// --- Paso 1: Generación del elemento IFS ---
		var ifsNameElementPart string
		var ifsPathElementPart string

		// El nombre para el IMM y el nombre para el IFS pueden ser diferentes
		var displayName string

		if row[headerMap["INFO"]] == "MvMoment" {
			displayName = strings.ReplaceAll(elementKey, "_", " ")
		} else {
			displayName = elementKey
		}

		if isBreaker || row[headerMap["ELEMENT"]] == "CB" {
			// Nomenclatura para Breaker: repetir el nombre del elemento
			ifsNameElementPart = fmt.Sprintf("%s_%s", displayName, displayName)
			ifsPathElementPart = fmt.Sprintf("%s/%s", displayName, displayName) // El path usa la llave original
		} else {
			ifsNameElementPart = displayName
			ifsPathElementPart = displayName // El path usa la llave original
		}

		// Determinar sufijo y valor SBO
		suffix := "M"
		if row[headerMap["TYPE"]] == "SP_SC" {
			suffix = "MC"
		}
		sbo := "0"
		if sboValue := row[headerMap["SBO"]]; sboValue != "" {
			sbo = sboValue
		}

		ifsPoint := &IfsPoint{
			Name: fmt.Sprintf("%s_%s_%s_%s_%s_%s",
				row[headerMap["B1"]], row[headerMap["B2"]], row[headerMap["B3"]],
				ifsNameElementPart, row[headerMap["INFO"]], suffix),
			MonAddrHigh:   row[headerMap["MHB"]],
			MonAddrMiddle: row[headerMap["MMB"]],
			MonAddrLow:    row[headerMap["MLB"]],
			MonType:       "0",
			ConAddrHigh:   row[headerMap["CHB"]],
			ConAddrMiddle: row[headerMap["CMB"]],
			ConAddrLow:    row[headerMap["CLB"]],
			ConType:       map[bool]string{true: "45", false: "0"}[row[headerMap["TYPE"]] == "SP_SC"],
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
			log.Printf("Advertencia: La llave '%s' no fue encontrada en la plantilla. Se omite para la generación IMM.", elementKey)
			continue
		}

		instance, err := deepCopy(template)
		if err != nil {
			log.Printf("Error al copiar la plantilla para la llave '%s': %v. Se omite.", elementKey, err)
			continue
		}

		aor := row[headerMap["AOR"]]
		var elementToAppend any

		if instance.Analog != nil {
			instance.Analog.Name = displayName // Usar el nombre con espacios si aplica
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
				instance.Breaker.Discrete.Name = displayName // También actualizar el discreto anidado
				instance.Breaker.Discrete.AreaOfResponsibilityId = aor
			}
			elementToAppend = instance.Breaker
		}

		if elementToAppend != nil {
			elementsForIMM = append(elementsForIMM, elementToAppend)
		}
	}
	return
}

func createAndSaveXML(fileName, parentPath string, elements []any) error {
	// (Esta función no requiere cambios)
	if len(elements) == 0 {
		log.Printf("Información: No se generará el archivo '%s' porque no hay elementos para incluir.", fileName)
		return nil
	}

	outputStruct := XDF{
		Lang:    xmlLang,
		Version: xmlVersion,
		Instances: Instances{
			Parent: Parent{
				Path:     parentPath,
				Elements: elements,
			},
		},
	}

	var out bytes.Buffer
	out.WriteString(xml.Header)
	encoder := xml.NewEncoder(&out)
	encoder.Indent("", "    ")

	if err := encoder.Encode(outputStruct); err != nil {
		return fmt.Errorf("error al codificar el XML para '%s': %w", fileName, err)
	}
	if err := os.WriteFile(fileName, out.Bytes(), 0644); err != nil {
		return fmt.Errorf("error al escribir el archivo '%s': %w", fileName, err)
	}

	fmt.Printf("✅ Archivo generado exitosamente: %s\n", fileName)
	return nil
}
