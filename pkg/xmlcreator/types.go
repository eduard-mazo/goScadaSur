// pkg/xmlcreator/types.go
package xmlcreator

import "encoding/xml"

// ===================================================================================
// ESTRUCTURAS PARA ELEMENTOS XML
// ===================================================================================

// ElementDef define un elemento que puede ser Analog, Discrete, Breaker o IfsPoint
type ElementDef struct {
	Analog   *Analog   `json:"Analog,omitempty" xml:"Analog,omitempty"`
	Discrete *Discrete `json:"Discrete,omitempty" xml:"Discrete,omitempty"`
	Breaker  *Breaker  `json:"Breaker,omitempty" xml:"Breaker,omitempty"`
	IfsPoint *IfsPoint `json:"IfsPoint,omitempty" xml:"IfsPoint,omitempty"`
}

// Analog representa un elemento analógico
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

// AnalogValue representa el valor de un analógico
type AnalogValue struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	Archive  string `json:"Archive,omitempty" xml:"Archive,attr"`
	InfoName string `json:"InfoName,omitempty" xml:"InfoName,attr"`
}

// AnalogInfo representa información adicional de un analógico
type AnalogInfo struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	Value    string `json:"Value,omitempty" xml:"Value,attr"`
	InfoName string `json:"InfoName,omitempty" xml:"InfoName,attr"`
}

// Discrete representa un elemento discreto
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

// DiscreteValue representa el valor de un discreto
type DiscreteValue struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	InfoName string `json:"InfoName,omitempty" xml:"InfoName,attr"`
}

// DiscreteInfo representa información adicional de un discreto
type DiscreteInfo struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	Value    string `json:"Value,omitempty" xml:"Value,attr"`
	InfoName string `json:"InfoName,omitempty" xml:"InfoName,attr"`
}

// Breaker representa un interruptor
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

// Terminal representa un terminal de un breaker
type Terminal struct {
	Name     string `json:"Name,omitempty" xml:"Name,attr"`
	EquipEnd string `json:"EquipEnd,omitempty" xml:"EquipEnd,attr"`
}

// IfsPoint representa un punto IFS
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

// Link_IfsPointLinksToInfo enlaza un IfsPoint con información
type Link_IfsPointLinksToInfo struct {
	PathB string `xml:"PathB,attr"`
}

// Link_TerminalMeasuredByMeasurement enlaza un terminal con mediciones
type Link_TerminalMeasuredByMeasurement struct {
	XMLName xml.Name `xml:"Link_TerminalMeasuredByMeasurement"`
	PathB   string   `xml:"PathB,attr"`
}

// LinkedTerminal representa un terminal con enlaces
type LinkedTerminal struct {
	XMLName xml.Name `xml:"Terminal"`
	Name    string   `xml:"Name,attr"`
	Links   []Link_TerminalMeasuredByMeasurement
}

// ===================================================================================
// ESTRUCTURAS PARA DOCUMENTO XML XDF
// ===================================================================================

// XDF es la estructura raíz del documento XML
type XDF struct {
	XMLName   xml.Name  `xml:"XDF"`
	Lang      string    `xml:"xml:lang,attr"`
	Version   string    `xml:"XdfTypeSyntaxVersion,attr"`
	Instances Instances `xml:"Instances"`
}

// Instances contiene los parents del XML
type Instances struct {
	Parents []Parent `xml:"Parent"`
}

// Parent representa un contenedor de elementos XML
type Parent struct {
	Path     string `xml:"Path,attr"`
	Elements []any  `xml:",any"`
}
