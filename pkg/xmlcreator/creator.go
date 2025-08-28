// main.go
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
	Analog   *Analog   `xml:"Analog,omitempty"`
	Discrete *Discrete `xml:"Discrete,omitempty"`
	Breaker  *Breaker  `xml:"Breaker,omitempty"`
	IfsPoint *IfsPoint `xml:"IfsPoint,omitempty"`
}

type Analog struct {
	XMLName                xml.Name     `xml:"Analog"`
	Name                   string       `xml:"Name,attr"`
	UnitOfMeasure          string       `xml:"UnitOfMeasure,attr"`
	ElementType            string       `xml:"ElementType,attr"`
	ElementName            string       `xml:"ElementName,attr"`
	MeasurementType        string       `xml:"MeasurementType,attr"`
	AreaOfResponsibilityId string       `xml:"AreaOfResponsibilityId,attr"`
	AnalogValue            *AnalogValue `xml:"AnalogValue,omitempty"`
	AnalogInfo             *AnalogInfo  `xml:"AnalogInfo,omitempty"`
}

type AnalogValue struct {
	Name     string `xml:"Name,attr"`
	Archive  string `xml:"Archive,attr"`
	InfoName string `xml:"InfoName,attr"`
}
type AnalogInfo struct {
	Name     string `xml:"Name,attr"`
	Value    string `xml:"Value,attr"`
	InfoName string `xml:"InfoName,attr"`
}

type Discrete struct {
	XMLName                xml.Name       `xml:"Discrete"`
	Name                   string         `xml:"Name,attr"`
	ElementType            string         `xml:"ElementType,attr"`
	ElementName            string         `xml:"ElementName,attr"`
	MeasurementType        string         `xml:"MeasurementType,attr"`
	AreaOfResponsibilityId string         `xml:"AreaOfResponsibilityId,attr"`
	DiscreteValue          *DiscreteValue `xml:"DiscreteValue,omitempty"`
	DiscreteInfo           *DiscreteInfo  `xml:"DiscreteInfo,omitempty"`
}

type DiscreteValue struct {
	Name     string `xml:"Name,attr"`
	InfoName string `xml:"InfoName,attr"`
}
type DiscreteInfo struct {
	Name     string `xml:"Name,attr"`
	Value    string `xml:"Value,attr"`
	InfoName string `xml:"InfoName,attr"`
}

type Breaker struct {
	XMLName                xml.Name    `xml:"Breaker"`
	Name                   string      `xml:"Name,attr"`
	FlowBreakerFlag        string      `xml:"FlowBreakerFlag,attr"`
	VoltMagLimitCA         string      `xml:"VoltMagLimitCA,attr"`
	DMSFlag                string      `xml:"DMSFlag,attr"`
	AreaOfResponsibilityId string      `xml:"AreaOfResponsibilityId,attr"`
	Terminals              []*Terminal `xml:"Terminal,omitempty"`
	Discrete               *Discrete   `xml:"Discrete,omitempty"`
}

type Terminal struct {
	Name     string `xml:"Name,attr"`
	EquipEnd string `xml:"EquipEnd,attr"`
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
	SelectBefore             string                    `xml:"SelectBefore,attr"`
	Link_IfsPointLinksToInfo *Link_IfsPointLinksToInfo `xml:"Link_IfsPointLinksToInfo,omitempty"`
}

type Link_IfsPointLinksToInfo struct {
	PathB string `xml:"PathB,attr"`
}

// ===================================================================================
// 2. PLANTILLA BASE DE ELEMENTOS (Sin cambios)
// ===================================================================================
var elementDB = map[string]ElementDef{
	"AjProGr1": {
		Breaker: &Breaker{
			Name:                   "AjProGr1",
			FlowBreakerFlag:        "true",
			VoltMagLimitCA:         "0",
			DMSFlag:                "true",
			AreaOfResponsibilityId: "{AOR}",
			Terminals: []*Terminal{
				{Name: "T2", EquipEnd: "2"},
				{Name: "T1", EquipEnd: "1"},
			},
			Discrete: &Discrete{
				Name:                   "AjProGr1",
				ElementType:            "2208",
				ElementName:            "1755",
				MeasurementType:        "22",
				AreaOfResponsibilityId: "{AOR}",
				DiscreteValue:          &DiscreteValue{Name: "Status", InfoName: "60000003"},
				DiscreteInfo:           &DiscreteInfo{Name: "NormStat", Value: "1", InfoName: "60000023"},
			},
		},
	},
	"AjProGr2": {
		Breaker: &Breaker{
			Name:                   "AjProGr2",
			FlowBreakerFlag:        "true",
			VoltMagLimitCA:         "0",
			DMSFlag:                "true",
			AreaOfResponsibilityId: "{AOR}",
			Terminals: []*Terminal{
				{Name: "T2", EquipEnd: "2"},
				{Name: "T1", EquipEnd: "1"},
			},
			Discrete: &Discrete{
				Name:                   "AjProGr2",
				ElementType:            "2208",
				ElementName:            "1754",
				MeasurementType:        "22",
				AreaOfResponsibilityId: "{AOR}",
				DiscreteValue:          &DiscreteValue{Name: "Status", InfoName: "60000003"},
				DiscreteInfo:           &DiscreteInfo{Name: "NormStat", Value: "0", InfoName: "60000023"},
			},
		},
	},
	"AjProGr3": {
		Breaker: &Breaker{
			Name:                   "AjProGr3",
			FlowBreakerFlag:        "true",
			VoltMagLimitCA:         "0",
			DMSFlag:                "true",
			AreaOfResponsibilityId: "{AOR}",
			Terminals: []*Terminal{
				{Name: "T2", EquipEnd: "2"},
				{Name: "T1", EquipEnd: "1"},
			},
			Discrete: &Discrete{
				Name:                   "AjProGr3",
				ElementType:            "2208",
				ElementName:            "1756",
				MeasurementType:        "22",
				AreaOfResponsibilityId: "{AOR}",
				DiscreteValue:          &DiscreteValue{Name: "Status", InfoName: "60000003"},
				DiscreteInfo:           &DiscreteInfo{Name: "NormStat", Value: "0", InfoName: "60000023"},
			},
		},
	},
	"BAH_07": {
		Discrete: &Discrete{
			Name:                   "BAH_07",
			ElementType:            "2873",
			ElementName:            "1431",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"BAH_09": {
		Discrete: &Discrete{
			Name:                   "BAH_09",
			ElementType:            "2873",
			ElementName:            "1041",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"Bl_Spec": {
		Discrete: &Discrete{
			Name:                   "Bl Spec",
			ElementType:            "5",
			ElementName:            "1",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
		},
	},
	"CTR_01": {
		Discrete: &Discrete{
			Name:                   "CTR_01",
			ElementType:            "2873",
			ElementName:            "1174",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"CTR_06": {
		Discrete: &Discrete{
			Name:                   "CTR_06",
			ElementType:            "2873",
			ElementName:            "1179",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"CTR_12": {
		Discrete: &Discrete{
			Name:                   "CTR_12",
			ElementType:            "2873",
			ElementName:            "1185",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"CTR_15": {
		Discrete: &Discrete{
			Name:                   "CTR_15",
			ElementType:            "2863",
			ElementName:            "1186",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"CTR_16": {
		Discrete: &Discrete{
			Name:                   "CTR_16",
			ElementType:            "2863",
			ElementName:            "1187",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"CTR_17": {
		Discrete: &Discrete{
			Name:                   "CTR_17",
			ElementType:            "2863",
			ElementName:            "1188",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"CtrlBloq": {
		Breaker: &Breaker{
			Name:                   "CtrlBloq",
			FlowBreakerFlag:        "true",
			VoltMagLimitCA:         "0",
			DMSFlag:                "true",
			AreaOfResponsibilityId: "{AOR}",
			Terminals: []*Terminal{
				{Name: "T2", EquipEnd: "2"},
				{Name: "T1", EquipEnd: "1"},
			},
			Discrete: &Discrete{
				Name:                   "CtrlBloq",
				ElementType:            "2208",
				ElementName:            "1765",
				MeasurementType:        "22",
				AreaOfResponsibilityId: "{AOR}",
				DiscreteValue:          &DiscreteValue{Name: "Status", InfoName: "60000003"},
				DiscreteInfo:           &DiscreteInfo{Name: "NormStat", Value: "0", InfoName: "60000023"},
			},
		},
	},
	"INT_01": {
		Discrete: &Discrete{
			Name:                   "INT_01",
			ElementType:            "2873",
			ElementName:            "853",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"INT_30": {
		Discrete: &Discrete{
			Name:                   "INT_30",
			ElementType:            "2873",
			ElementName:            "1160",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"IRFalla": {
		Analog: &Analog{
			Name:                   "IRFalla",
			UnitOfMeasure:          "A",
			ElementType:            "10",
			ElementName:            "1809",
			MeasurementType:        "2",
			AreaOfResponsibilityId: "{AOR}",
			AnalogValue:            &AnalogValue{Name: "MvMoment", Archive: "true", InfoName: "20000001"},
			AnalogInfo:             &AnalogInfo{Name: "MvNomina", Value: "0", InfoName: "20000002"},
		},
	},
	"ISFalla": {
		Analog: &Analog{
			Name:                   "ISFalla",
			UnitOfMeasure:          "A",
			ElementType:            "10",
			ElementName:            "1810",
			MeasurementType:        "2",
			AreaOfResponsibilityId: "{AOR}",
			AnalogValue:            &AnalogValue{Name: "MvMoment", Archive: "true", InfoName: "20000001"},
			AnalogInfo:             &AnalogInfo{Name: "MvNomina", Value: "0", InfoName: "20000002"},
		},
	},
	"ITFalla": {
		Analog: &Analog{
			Name:                   "ITFalla",
			UnitOfMeasure:          "A",
			ElementType:            "10",
			ElementName:            "1811",
			MeasurementType:        "2",
			AreaOfResponsibilityId: "{AOR}",
			AnalogValue:            &AnalogValue{Name: "MvMoment", Archive: "true", InfoName: "20000001"},
			AnalogInfo:             &AnalogInfo{Name: "MvNomina", Value: "0", InfoName: "20000002"},
		},
	},
	"I_R": {
		Analog: &Analog{
			Name:                   "I R",
			UnitOfMeasure:          "A",
			ElementType:            "10",
			ElementName:            "1905",
			MeasurementType:        "2",
			AreaOfResponsibilityId: "{AOR}",
			AnalogValue:            &AnalogValue{Name: "MvMoment", Archive: "true", InfoName: "20000001"},
			AnalogInfo:             &AnalogInfo{Name: "MvNomina", Value: "0", InfoName: "20000002"},
		},
	},
	"I_S": {
		Analog: &Analog{
			Name:                   "I S",
			UnitOfMeasure:          "A",
			ElementType:            "10",
			ElementName:            "1906",
			MeasurementType:        "2",
			AreaOfResponsibilityId: "{AOR}",
			AnalogValue:            &AnalogValue{Name: "MvMoment", Archive: "true", InfoName: "20000001"},
			AnalogInfo:             &AnalogInfo{Name: "MvNomina", Value: "0", InfoName: "20000002"},
		},
	},
	"I_T": {
		Analog: &Analog{
			Name:                   "I T",
			UnitOfMeasure:          "A",
			ElementType:            "10",
			ElementName:            "1907",
			MeasurementType:        "2",
			AreaOfResponsibilityId: "{AOR}",
			AnalogValue:            &AnalogValue{Name: "MvMoment", Archive: "true", InfoName: "20000001"},
			AnalogInfo:             &AnalogInfo{Name: "MvNomina", Value: "0", InfoName: "20000002"},
		},
	},
	"P1_13": {
		Discrete: &Discrete{
			Name:                   "P1_13",
			ElementType:            "2863",
			ElementName:            "880",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"P1_14": {
		Discrete: &Discrete{
			Name:                   "P1_14",
			ElementType:            "2863",
			ElementName:            "881",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"P1_19": {
		Discrete: &Discrete{
			Name:                   "P1_19",
			ElementType:            "2863",
			ElementName:            "952",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"P1_20": {
		Discrete: &Discrete{
			Name:                   "P1_20",
			ElementType:            "2863",
			ElementName:            "954",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"P1_21": {
		Discrete: &Discrete{
			Name:                   "P1_21",
			ElementType:            "2863",
			ElementName:            "956",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"P1_22": {
		Discrete: &Discrete{
			Name:                   "P1_22",
			ElementType:            "2863",
			ElementName:            "953",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"P1_23": {
		Discrete: &Discrete{
			Name:                   "P1_23",
			ElementType:            "2863",
			ElementName:            "955",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"P1_24": {
		Discrete: &Discrete{
			Name:                   "P1_24",
			ElementType:            "2863",
			ElementName:            "957",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"P1_27": {
		Discrete: &Discrete{
			Name:                   "P1_27",
			ElementType:            "2863",
			ElementName:            "972",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"P1_48": {
		Discrete: &Discrete{
			Name:                   "P1_48",
			ElementType:            "2873",
			ElementName:            "1048",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"PrLinViv": {
		Breaker: &Breaker{
			Name:                   "PrLinViv",
			FlowBreakerFlag:        "true",
			VoltMagLimitCA:         "0",
			DMSFlag:                "true",
			AreaOfResponsibilityId: "{AOR}",
			Terminals: []*Terminal{
				{Name: "T2", EquipEnd: "2"},
				{Name: "T1", EquipEnd: "1"},
			},
			Discrete: &Discrete{
				Name:                   "PrLinViv",
				ElementType:            "2208",
				ElementName:            "1763",
				MeasurementType:        "22",
				AreaOfResponsibilityId: "{AOR}",
				DiscreteValue:          &DiscreteValue{Name: "Status", InfoName: "60000003"},
				DiscreteInfo:           &DiscreteInfo{Name: "NormStat", Value: "0", InfoName: "60000023"},
			},
		},
	},
	"PrTierra": {
		Breaker: &Breaker{
			Name:                   "PrTierra",
			FlowBreakerFlag:        "true",
			VoltMagLimitCA:         "0",
			DMSFlag:                "true",
			AreaOfResponsibilityId: "{AOR}",
			Terminals: []*Terminal{
				{Name: "T2", EquipEnd: "2"},
				{Name: "T1", EquipEnd: "1"},
			},
			Discrete: &Discrete{
				Name:                   "PrTierra",
				ElementType:            "2208",
				ElementName:            "1760",
				MeasurementType:        "22",
				AreaOfResponsibilityId: "{AOR}",
				DiscreteValue:          &DiscreteValue{Name: "Status", InfoName: "60000003"},
				DiscreteInfo:           &DiscreteInfo{Name: "NormStat", Value: "1", InfoName: "60000023"},
			},
		},
	},
	"Protcion": {
		Breaker: &Breaker{
			Name:                   "Protcion",
			FlowBreakerFlag:        "true",
			VoltMagLimitCA:         "0",
			DMSFlag:                "true",
			AreaOfResponsibilityId: "{AOR}",
			Terminals: []*Terminal{
				{Name: "T2", EquipEnd: "2"},
				{Name: "T1", EquipEnd: "1"},
			},
			Discrete: &Discrete{
				Name:                   "Protcion",
				ElementType:            "2208",
				ElementName:            "1759",
				MeasurementType:        "22",
				AreaOfResponsibilityId: "{AOR}",
				DiscreteValue:          &DiscreteValue{Name: "Status", InfoName: "60000003"},
			},
		},
	},
	"Reclos": {
		Breaker: &Breaker{
			Name:                   "Reclos",
			FlowBreakerFlag:        "true",
			VoltMagLimitCA:         "0",
			DMSFlag:                "true",
			AreaOfResponsibilityId: "{AOR}",
			Terminals: []*Terminal{
				{Name: "T2", EquipEnd: "2"},
				{Name: "T1", EquipEnd: "1"},
			},
			Discrete: &Discrete{
				Name:                   "Reclos",
				ElementType:            "2208",
				ElementName:            "731",
				MeasurementType:        "22",
				AreaOfResponsibilityId: "{AOR}",
				DiscreteValue:          &DiscreteValue{Name: "Status", InfoName: "60000003"},
				DiscreteInfo:           &DiscreteInfo{Name: "NormStat", Value: "1", InfoName: "60000023"},
			},
		},
	},
	"SA_05": {
		Discrete: &Discrete{
			Name:                   "SA_05",
			ElementType:            "2873",
			ElementName:            "1882",
			MeasurementType:        "0",
			AreaOfResponsibilityId: "{AOR}",
			DiscreteInfo:           &DiscreteInfo{Name: "AlStat", Value: "", InfoName: "180000007"},
		},
	},
	"U_RS": {
		Analog: &Analog{
			Name:                   "U RS",
			UnitOfMeasure:          "kV",
			ElementType:            "17",
			ElementName:            "1953",
			MeasurementType:        "2",
			AreaOfResponsibilityId: "{AOR}",
			AnalogValue:            &AnalogValue{Name: "MvMoment", Archive: "true", InfoName: "20000001"},
			AnalogInfo:             &AnalogInfo{Name: "MvNomina", Value: "0", InfoName: "20000002"},
		},
	},
	"U_ST": {
		Analog: &Analog{
			Name:                   "U ST",
			UnitOfMeasure:          "kV",
			ElementType:            "17",
			ElementName:            "1954",
			MeasurementType:        "2",
			AreaOfResponsibilityId: "{AOR}",
			AnalogValue:            &AnalogValue{Name: "MvMoment", Archive: "true", InfoName: "20000001"},
			AnalogInfo:             &AnalogInfo{Name: "MvNomina", Value: "0", InfoName: "20000002"},
		},
	},
	"U_TR": {
		Analog: &Analog{
			Name:                   "U TR",
			UnitOfMeasure:          "kV",
			ElementType:            "17",
			ElementName:            "1955",
			MeasurementType:        "2",
			AreaOfResponsibilityId: "{AOR}",
			AnalogValue:            &AnalogValue{Name: "MvMoment", Archive: "true", InfoName: "20000001"},
			AnalogInfo:             &AnalogInfo{Name: "MvNomina", Value: "0", InfoName: "20000002"},
		},
	},
}

// deepCopy crea una copia independiente de un elemento para evitar modificar la plantilla original.
func deepCopy(elementDef ElementDef) (ElementDef, error) {
	var copiedDef ElementDef
	bytes, err := json.Marshal(elementDef)
	if err != nil {
		return copiedDef, err
	}
	err = json.Unmarshal(bytes, &copiedDef)
	return copiedDef, err
}

// ===================================================================================
// 3. LÓGICA PRINCIPAL DEL PROGRAMA (Refactorizada)
// ===================================================================================

const (
	xmlLang       = "EN"
	xmlVersion    = "2.0.00"
	ifsParentPath = "SCADA/RTU" // Path base para el archivo IFS
)

func CreateXML(payloadJSON, empresa, region, aor string) error {
	// --- 1. Decodificar JSON para capturar Header y Rows ---
	dataRows, headerMap, err := readCSVFromJSON(payloadJSON, empresa, region, aor)
	if err != nil {
		return err
	}

	// --- 2. Procesar filas para generar los elementos de ambos XML en un solo paso ---
	elementsForIMM, elementsForIFS := processRows(dataRows, headerMap)

	// --- 3. Ensamblar y guardar los archivos XML ---
	firstRow := dataRows[0]
	b3Value := firstRow[headerMap["B3"]]

	// Generar Path para el archivo IMM usando datos de la primera fila
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

	// Crear y guardar el archivo IFS
	fileNameIFS := b3Value + "_IFS.xml"
	if err := createAndSaveXML(fileNameIFS, ifsParentPath, elementsForIFS); err != nil {
		return err
	}

	return nil
}

// readCSVFromJSON processes a JSON payload to extract CSV-like data and headers.
// It expects the JSON to have "columns" (an array of objects with "name" fields)
// and "data" (an array of objects representing rows).
func readCSVFromJSON(payloadJSON, empresa, region, aor string) ([][]string, map[string]int, error) {
	if !gjson.Valid(payloadJSON) {
		return nil, nil, errors.New("payloadJSON no es un JSON válido")
	}

	// 1. Combine headers: constants + dynamic
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

	// 2. Build Header Map
	headerMap := make(map[string]int)
	for i, name := range headers {
		headerMap[name] = i
	}

	// 3. Extract Data Rows
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
		// Start each record with the constant values
		record := []string{empresa, region, aor}

		// Append the dynamic values from the JSON row
		for _, header := range dynamicHeaders {
			value := row.Get(header)
			record = append(record, value.String())
		}
		records = append(records, record)
		return true
	})

	if len(records) == 0 {
		return records, headerMap, nil
	}

	return records, headerMap, nil
}

// processRows itera una sola vez sobre los registros del CSV para generar las listas de elementos
// para los archivos IMM e IFS, mejorando la eficiencia.
func processRows(dataRows [][]string, headerMap map[string]int) (elementsForIMM []any, elementsForIFS []any) {
	for _, row := range dataRows {
		elementKey := row[headerMap["ELEMENT"]]

		// --- Generación de elementos IFS para cada fila ---
		var suffix, sbo string
		if row[headerMap["TYPE"]] == "SP_SC" {
			suffix = "MC"
		} else {
			suffix = "M"
		}

		if row[headerMap["SBO"]] == "" {
			sbo = "0"
		} else {
			sbo = row[headerMap["SBO"]]
		}

		ifsPoint := &IfsPoint{
			Name: fmt.Sprintf("%s_%s_%s_%s_%s_%s",
				row[headerMap["B1"]], row[headerMap["B2"]], row[headerMap["B3"]],
				elementKey, row[headerMap["INFO"]], suffix),
			MonAddrHigh:   row[headerMap["HB"]],
			MonAddrMiddle: row[headerMap["MB"]],
			MonAddrLow:    row[headerMap["LB"]],
			MonType:       "0", // Valor constante
			SelectBefore:  sbo,
			Link_IfsPointLinksToInfo: &Link_IfsPointLinksToInfo{
				PathB: fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s/%s/%s",
					row[headerMap["EMPRESA"]], row[headerMap["REGION"]], row[headerMap["B1"]],
					row[headerMap["B2"]], row[headerMap["B3"]], elementKey, row[headerMap["INFO"]]),
			},
		}
		elementsForIFS = append(elementsForIFS, ifsPoint)

		// --- Generación de elementos IMM para cada fila ---
		template, ok := elementDB[elementKey]
		if !ok {
			log.Printf("Advertencia: La llave '%s' del CSV no fue encontrada en la plantilla. Se omite para la generación IMM.", elementKey)
			continue
		}

		instance, err := deepCopy(template)
		if err != nil {
			log.Printf("Error al copiar la plantilla para la llave '%s': %v. Se omite para la generación IMM.", elementKey, err)
			continue
		}

		aor := row[headerMap["AOR"]]
		var elementToAppend any

		if instance.Analog != nil {
			instance.Analog.Name = strings.ReplaceAll(elementKey, "_", " ")
			instance.Analog.AreaOfResponsibilityId = aor
			elementToAppend = instance.Analog
		} else if instance.Discrete != nil {
			instance.Discrete.Name = elementKey
			instance.Discrete.AreaOfResponsibilityId = aor
			elementToAppend = instance.Discrete
		} else if instance.Breaker != nil {
			instance.Breaker.Name = elementKey
			instance.Breaker.AreaOfResponsibilityId = aor
			if instance.Breaker.Discrete != nil {
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
		return fmt.Errorf("error al codificar el XML: %w", err)
	}
	if err := os.WriteFile(fileName, out.Bytes(), 0644); err != nil {
		return fmt.Errorf("error al escribir el archivo: %w", err)
	}

	fmt.Printf("✅ Archivo generado exitosamente: %s\n", fileName)
	return nil
}
