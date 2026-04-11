// pkg/xmlcreator/creator.go
package xmlcreator

import (
	"fmt"
	"goScadaSur/pkg/config"
	"goScadaSur/pkg/fileio"
	"log"
	"strings"
)

type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type defaultLogger struct{}

func (defaultLogger) Infof(f string, a ...any)  { log.Printf("[INFO] "+f, a...) }
func (defaultLogger) Warnf(f string, a ...any)  { log.Printf("[WARN] "+f, a...) }
func (defaultLogger) Errorf(f string, a ...any) { log.Printf("[ERROR] "+f, a...) }

var columnAliases = map[string]string{
	"M_LB": "MLB", "M_MB": "MMB", "M_HB": "MHB",
	"C_LB": "CLB", "C_MB": "CMB", "C_HB": "CHB",
	"SBE": "SBO",
}

func normalizeHeaderMap(h map[string]int) map[string]int {
	out := make(map[string]int, len(h))
	for k, v := range h {
		out[k] = v
	}
	for a, c := range columnAliases {
		if idx, ok := h[a]; ok {
			if _, ex := out[c]; !ex {
				out[c] = idx
			}
		}
	}
	return out
}

// CreateXMLFromFile mantiene la firma original (CLI).
func CreateXMLFromFile(p string, cfg *config.AppConfig, dCfg *config.DasipConfig, tm *TemplateManager) error {
	return CreateXMLFromFileWithLogger(p, cfg, dCfg, tm, defaultLogger{})
}

// CreateXMLFromFileWithLogger agrupa filas por estación y genera un par IFS+IMM por cada una.
func CreateXMLFromFileWithLogger(inputFilePath string, cfg *config.AppConfig, dasipCfg *config.DasipConfig, tm *TemplateManager, lg Logger) error {
	lg.Infof("Leyendo datos desde: %s", inputFilePath)
	_, dataRows, headerMap, err := fileio.ReadData(inputFilePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo: %w", err)
	}
	if len(dataRows) == 0 {
		lg.Warnf("El archivo no contiene datos para procesar")
		return nil
	}
	headerMap = normalizeHeaderMap(headerMap)
	if err := fileio.ValidateHeaders(headerMap, cfg.Validation.RequiredColumns); err != nil {
		return fmt.Errorf("validación de columnas fallida: %w", err)
	}

	groups := make(map[string][][]string)
	order := []string{}
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
	lg.Infof("%d filas leídas, %d estaciones detectadas", len(dataRows), len(order))

	for _, key := range order {
		rows := groups[key]
		lg.Infof("Procesando estación %s (%d filas)", key, len(rows))
		result, err := processRows(rows, headerMap, tm, lg)
		if err != nil {
			lg.Warnf("estación %s: %v", key, err)
			continue
		}
		if err := generateXMLFiles(result, rows[0], headerMap, cfg, dasipCfg, lg); err != nil {
			lg.Warnf("estación %s: %v", key, err)
		}
	}
	return nil
}

type ProcessingResult struct {
	ElementsIMM  []any
	ElementsIFS  []any
	BreakerName  string
	BreakerLinks []any
}

func processRows(dataRows [][]string, h map[string]int, tm *TemplateManager, lg Logger) (*ProcessingResult, error) {
	r := &ProcessingResult{ElementsIMM: []any{}, ElementsIFS: []any{}, BreakerLinks: []any{}}
	var cb []string
	for i, row := range dataRows {
		if len(row) == 0 {
			continue
		}
		k := fileio.GetCellValue(row, h["ELEMENT"])
		if k == "" {
			lg.Warnf("Fila %d: ELEMENT vacío", i+2)
			continue
		}
		tpl, found := tm.GetTemplate(k)
		isBrk := (found && tpl.Breaker != nil) || k == "CB"
		if k == "CB" {
			cb = row
		}
		dn := generateDisplayName(k, row, h)
		r.ElementsIFS = append(r.ElementsIFS, createIfsPoint(row, h, dn, isBrk))
		if !found {
			lg.Warnf("Plantilla '%s' no encontrada", k)
			continue
		}
		el, err := createIMMElement(tpl, dn, row, h, tm)
		if err != nil {
			lg.Warnf("elemento '%s': %v", k, err)
			continue
		}
		if el != nil {
			r.ElementsIMM = append(r.ElementsIMM, el)
		}
	}
	if cb != nil {
		r.BreakerLinks, r.BreakerName = createBreakerLinks(cb, dataRows, h)
	}
	return r, nil
}

func generateDisplayName(k string, row []string, h map[string]int) string {
	if fileio.GetCellValue(row, h["INFO"]) == "MvMoment" {
		return strings.ReplaceAll(k, "_", " ")
	}
	return k
}

func createIfsPoint(row []string, h map[string]int, dn string, brk bool) *IfsPoint {
	var nm, pt string
	if brk {
		nm = fmt.Sprintf("%s_%s", dn, dn)
		pt = fmt.Sprintf("%s/%s", dn, dn)
	} else {
		nm, pt = dn, dn
	}
	suf, ct := "M", "0"
	if fileio.GetCellValue(row, h["TYPE"]) == "SP_SC" {
		suf, ct = "MC", "45"
	}
	sbo := fileio.GetCellValueOrDefault(row, h, "SBO", "0")
	b1 := fileio.GetCellValue(row, h["B1"])
	b2 := fileio.GetCellValue(row, h["B2"])
	b3 := fileio.GetCellValue(row, h["B3"])
	info := fileio.GetCellValue(row, h["INFO"])
	emp := fileio.GetCellValue(row, h["EMPRESA"])
	reg := fileio.GetCellValue(row, h["REGION"])
	return &IfsPoint{
		Name:                     fmt.Sprintf("%s_%s_%s_%s_%s_%s", b1, b2, b3, nm, info, suf),
		MonAddrHigh:              fileio.GetCellValueOrDefault(row, h, "MHB", "0"),
		MonAddrMiddle:            fileio.GetCellValueOrDefault(row, h, "MMB", "0"),
		MonAddrLow:               fileio.GetCellValueOrDefault(row, h, "MLB", "0"),
		MonType:                  "0",
		ConAddrHigh:              fileio.GetCellValueOrDefault(row, h, "CHB", "0"),
		ConAddrMiddle:            fileio.GetCellValueOrDefault(row, h, "CMB", "0"),
		ConAddrLow:               fileio.GetCellValueOrDefault(row, h, "CLB", "0"),
		ConType:                  ct,
		SelectBefore:             sbo,
		Link_IfsPointLinksToInfo: &Link_IfsPointLinksToInfo{PathB: fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s/%s/%s", emp, reg, b1, b2, b3, pt, info)},
	}
}

func createIMMElement(tpl ElementDef, dn string, row []string, h map[string]int, tm *TemplateManager) (any, error) {
	inst, err := tm.DeepCopyElement(tpl)
	if err != nil {
		return nil, err
	}
	aor := fileio.GetCellValue(row, h["AOR"])
	if inst.Analog != nil {
		inst.Analog.Name = dn
		inst.Analog.AreaOfResponsibilityId = aor
		return inst.Analog, nil
	}
	if inst.Discrete != nil {
		inst.Discrete.Name = dn
		inst.Discrete.AreaOfResponsibilityId = aor
		return inst.Discrete, nil
	}
	if inst.Breaker != nil {
		inst.Breaker.Name = dn
		inst.Breaker.AreaOfResponsibilityId = aor
		if inst.Breaker.Discrete != nil {
			inst.Breaker.Discrete.Name = dn
			inst.Breaker.Discrete.AreaOfResponsibilityId = aor
		}
		return inst.Breaker, nil
	}
	return nil, nil
}

func createBreakerLinks(cb []string, all [][]string, h map[string]int) ([]any, string) {
	tgt := map[string]bool{"P": true, "Q": true, "I_S": true, "U_RS": true}
	var links []Link_TerminalMeasuredByMeasurement
	base := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s",
		fileio.GetCellValue(cb, h["EMPRESA"]), fileio.GetCellValue(cb, h["REGION"]),
		fileio.GetCellValue(cb, h["B1"]), fileio.GetCellValue(cb, h["B2"]), fileio.GetCellValue(cb, h["B3"]))
	for _, row := range all {
		if len(row) <= h["ELEMENT"] {
			continue
		}
		k := fileio.GetCellValue(row, h["ELEMENT"])
		if tgt[k] {
			links = append(links, Link_TerminalMeasuredByMeasurement{PathB: fmt.Sprintf("%s/%s", base, strings.ReplaceAll(k, "_", " "))})
		}
	}
	if len(links) == 0 {
		return nil, ""
	}
	return []any{LinkedTerminal{Name: "T1", Links: links}}, "CB"
}

func generateXMLFiles(r *ProcessingResult, first []string, h map[string]int, cfg *config.AppConfig, dCfg *config.DasipConfig, lg Logger) error {
	b3 := fileio.GetCellValue(first, h["B3"])
	emp := fileio.GetCellValue(first, h["EMPRESA"])
	reg := fileio.GetCellValue(first, h["REGION"])
	b1 := fileio.GetCellValue(first, h["B1"])
	b2 := fileio.GetCellValue(first, h["B2"])
	dasIP := fileio.GetCellValueOrDefault(first, h, "DASIP", "")
	if err := generateIFSFile(b3, dCfg.GetIfsParentPath(dasIP), r.ElementsIFS, cfg, lg); err != nil {
		return err
	}
	return generateIMMFile(b3, emp, reg, b1, b2, r, cfg, lg)
}

func generateIFSFile(b3, pp string, els []any, cfg *config.AppConfig, lg Logger) error {
	if len(els) == 0 {
		return nil
	}
	return createAndSaveXML(fmt.Sprintf("%s%s", b3, cfg.Output.Suffixes["ifs"]), []Parent{{Path: pp, Elements: els}}, cfg, lg)
}

func generateIMMFile(b3, emp, reg, b1, b2 string, r *ProcessingResult, cfg *config.AppConfig, lg Logger) error {
	if len(r.ElementsIMM) == 0 {
		return nil
	}
	pp := fmt.Sprintf("ELECTRICITY/NETWORK/%s/%s/%s/%s/%s", emp, reg, b1, b2, b3)
	parents := []Parent{{Path: pp, Elements: r.ElementsIMM}}
	if len(r.BreakerLinks) > 0 && r.BreakerName != "" {
		parents = append(parents, Parent{Path: fmt.Sprintf("%s/%s", pp, r.BreakerName), Elements: r.BreakerLinks})
	}
	return createAndSaveXML(fmt.Sprintf("%s%s", b3, cfg.Output.Suffixes["imm"]), parents, cfg, lg)
}

func createAndSaveXML(fn string, parents []Parent, cfg *config.AppConfig, lg Logger) error {
	xdf := XDF{Lang: cfg.XML.Lang, Version: cfg.XML.Version, Instances: Instances{Parents: parents}}
	if err := fileio.NewXMLWriter(cfg.GetOutputPath(fn), cfg.XML.Indent).Write(xdf); err != nil {
		return fmt.Errorf("error escribiendo XML '%s': %w", fn, err)
	}
	lg.Infof("Archivo generado: %s", fn)
	return nil
}
