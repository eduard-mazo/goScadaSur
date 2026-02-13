// main.go
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"goScadaSur/pkg/config"
	"goScadaSur/pkg/fileio"
	"goScadaSur/pkg/xmlcreator"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"golang.org/x/term"
)

const (
	defaultConfigPath = "configs/config.yaml"
)

// CSharpInput define la estructura para enviar datos a la aplicación C#
type CSharpInput struct {
	Mode     string `json:"mode"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Query    string `json:"query"`
	B1       string `json:"b1"`
	B2       string `json:"b2"`
	B3       string `json:"b3"`
}

var (
	// Flags globales
	configFile string
	user       string
	password   string
	host       string
	path       string
	aor        string
)

func main() {
	// Configurar logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Comando raíz
	rootCmd := &cobra.Command{
		Use:   "goScadaSur",
		Short: "Sistema de Gestión de Base de Datos SURVALENT",
		Long: `goScadaSur - Herramienta CLI para gestión de base de datos SURVALENT
		
Permite buscar estaciones, ejecutar queries directas, y generar archivos XML
desde archivos CSV o Excel.`,
		PersistentPreRun: initializeApp,
	}

	// Flags persistentes (disponibles en todos los comandos)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", defaultConfigPath, "Archivo de configuración")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "i", "", "Dirección IP del host")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "Usuario de la base de datos")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Contraseña de la base de datos")

	// Comando: station-search
	stationSearchCmd := &cobra.Command{
		Use:   "station-search",
		Short: "Busca una estación por nombre y retorna sus señales",
		Long: `Busca una estación en la base de datos SURVALENT y retorna
todas las señales asociadas. Requiere especificar el path (--path)
y el área de responsabilidad (--aor).`,
		Args: cobra.NoArgs,
		Run:  runStationSearch,
	}
	stationSearchCmd.Flags().StringVar(&path, "path", "", "Path del sistema (ej: B1/B2/B3)")
	stationSearchCmd.Flags().StringVar(&aor, "aor", "", "Área de responsabilidad")
	stationSearchCmd.MarkFlagRequired("path")
	stationSearchCmd.MarkFlagRequired("aor")

	// Comando: csv-xml (ahora soporta CSV y Excel)
	csvXmlCmd := &cobra.Command{
		Use:   "csv-xml",
		Short: "Genera archivos XML desde CSV o Excel",
		Long: `Genera archivos XML IFS e IMM a partir de archivos CSV o Excel.
		
Formatos soportados:
  - CSV (.csv)
  - Excel (.xlsx, .xls)
  
El archivo debe contener las columnas requeridas según la configuración.`,
		Args: cobra.NoArgs,
		Run:  runCSVToXML,
	}
	csvXmlCmd.Flags().StringVar(&path, "path", "", "Ruta del archivo CSV/Excel")
	csvXmlCmd.Flags().StringVar(&aor, "aor", "", "Área de responsabilidad")
	csvXmlCmd.MarkFlagRequired("path")
	csvXmlCmd.MarkFlagRequired("aor")

	// Comando: direct-query
	directQueryCmd := &cobra.Command{
		Use:   "direct-query [SQL query]",
		Short: "Ejecuta una query SQL directamente en la base de datos",
		Long: `Ejecuta una consulta SQL directa en la base de datos SURVALENT
y guarda los resultados en un archivo CSV.`,
		Args: cobra.ExactArgs(1),
		Run:  runDirectQuery,
	}

	// Comando: version
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Muestra la versión de la aplicación",
		Run: func(cmd *cobra.Command, args []string) {
			if config.Global != nil {
				fmt.Printf("%s v%s\n", config.Global.App.Name, config.Global.App.Version)
				fmt.Printf("%s\n", config.Global.App.Description)
			} else {
				fmt.Println("goScadaSur - versión no disponible (configuración no cargada)")
			}
		},
	}

	// Agregar comandos
	rootCmd.AddCommand(stationSearchCmd, directQueryCmd, csvXmlCmd, versionCmd)

	// Ejecutar
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
}

// initializeApp inicializa la configuración de la aplicación
func initializeApp(cmd *cobra.Command, args []string) {
	// Cargar configuración principal
	if err := config.Load(configFile); err != nil {
		log.Fatalf("❌ Error cargando configuración: %v", err)
	}

	log.Printf("✓ Configuración cargada desde: %s", configFile)

	// Cargar configuración DASIP
	dasipConfigPath := config.GetDasipConfigPath()
	if err := config.LoadDasipConfig(dasipConfigPath); err != nil {
		log.Printf("⚠️  Advertencia: Error cargando configuración DASIP: %v", err)
		log.Printf("   Usando configuración por defecto")
	} else {
		log.Printf("✓ Configuración DASIP cargada: %d mapeos", len(config.Dasip.DasipMapping))
	}

	// Cargar plantillas XML
	templatesPath := config.GetTemplatesPath()
	if err := xmlcreator.LoadTemplates(templatesPath); err != nil {
		log.Printf("⚠️  Advertencia: Error cargando plantillas: %v", err)
	}

	// Asegurar que el directorio de salida exista
	if err := config.EnsureOutputDir(); err != nil {
		log.Printf("⚠️  Advertencia: Error creando directorio de salida: %v", err)
	}
}

// runStationSearch ejecuta la búsqueda de estación
func runStationSearch(cmd *cobra.Command, args []string) {
	executeCommand("station_search", "", path, aor)
}

// runDirectQuery ejecuta una query directa
func runDirectQuery(cmd *cobra.Command, args []string) {
	query := args[0]
	executeCommand("direct_query", query, "", "")
}

// runCSVToXML ejecuta la conversión de CSV/Excel a XML
func runCSVToXML(cmd *cobra.Command, args []string) {
	// Verificar que el archivo exista
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("❌ Error: El archivo '%s' no existe", path)
	}

	// Verificar formato soportado
	ext := strings.ToLower(filepath.Ext(path))
	ext = strings.TrimPrefix(ext, ".")

	if !config.IsFormatSupported(ext) {
		log.Fatalf("❌ Error: Formato '%s' no soportado. Formatos válidos: %v",
			ext, config.Global.Files.SupportedInputFormats)
	}

	log.Printf("🚀 Procesando archivo: %s (formato: %s)", path, strings.ToUpper(ext))

	// Crear XMLs
	if err := xmlcreator.CreateXMLFromFile(path); err != nil {
		log.Fatalf("❌ Error generando XML: %v", err)
	}

	log.Println("✅ Proceso completado exitosamente")
}

// executeCommand ejecuta un comando hacia la base de datos C#
func executeCommand(mode, query, path, aor string) {
	empresa, region, b1, b2, b3, err := parsePath(path)
	if err != nil && mode != "direct_query" {
		log.Printf("⚠️  Error parseando path: %v", err)
	}

	// Solicitar credenciales si no están disponibles
	if host == "" {
		host = readInput("Host: ")
	}

	if user == "" {
		user = readInput("Usuario: ")
	}

	if password == "" {
		fmt.Print("Contraseña: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("❌ Error leyendo contraseña: %v", err)
		}
		fmt.Println()
		password = string(bytePassword)
	}

	// Preparar datos de entrada
	inputData := CSharpInput{
		Mode:     mode,
		User:     user,
		Password: password,
		Host:     host,
		Query:    query,
		B1:       b1,
		B2:       b2,
		B3:       b3,
	}

	inputBytes, _ := json.Marshal(inputData)
	inputJSON := string(inputBytes)

	// Ejecutar proceso C#
	csharpExe := config.Global.Database.CSharpExecutable
	cmd := exec.Command(csharpExe)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	stdinPipe, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		log.Fatalf("❌ Error iniciando proceso: %v", err)
	}

	// Enviar entrada
	go func() {
		defer stdinPipe.Close()
		fmt.Fprintln(stdinPipe, inputJSON)
	}()

	// Leer salida
	var wg sync.WaitGroup
	var outBuf, errBuf bytes.Buffer
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(&outBuf, stdoutPipe)
	}()

	go func() {
		defer wg.Done()
		io.Copy(&errBuf, stderrPipe)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		log.Printf("❌ Error en proceso: %v", err)
		if errBuf.Len() > 0 {
			log.Printf("STDERR: %s", errBuf.String())
		}
		return
	}

	if errBuf.Len() > 0 {
		fmt.Printf("⚠️  STDERR:\n%s\n", errBuf.String())
	}

	// Procesar salida
	output := outBuf.Bytes()
	if !gjson.ValidBytes(output) {
		log.Println("⚠️  La salida no es JSON válido")
		fmt.Printf("Salida:\n%s\n", outBuf.String())
		return
	}

	// Verificar integridad
	payloadJSON := gjson.GetBytes(output, "payload").Raw
	receivedChecksum := gjson.GetBytes(output, "checksum").String()

	if payloadJSON == "" || receivedChecksum == "" {
		log.Fatal("❌ Respuesta JSON inválida (falta payload o checksum)")
	}

	hasher := sha256.New()
	hasher.Write([]byte(payloadJSON))
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if calculatedChecksum != receivedChecksum {
		log.Fatal("❌ ERROR DE INTEGRIDAD! Los datos pueden estar corruptos")
	}

	log.Println("✓ Verificación de integridad exitosa")

	// Procesar según el modo
	switch mode {
	case "direct_query":
		timestamp := time.Now().Format(config.Global.Output.TimestampFormat)
		filename := fmt.Sprintf("%s_direct_query%s", timestamp, config.Global.Output.Suffixes["csv"])
		filename = config.GetOutputPath(filename)

		if err := saveDirectQueryToCSV(payloadJSON, filename); err != nil {
			log.Fatalf("❌ Error guardando query: %v", err)
		}
		log.Printf("✅ Resultados guardados en: %s", filename)

	case "station_search":
		timestamp := time.Now().Format(config.Global.Output.TimestampFormat)
		filename := fmt.Sprintf("%s_%s%s", timestamp, b3, config.Global.Output.Suffixes["csv"])
		filename = config.GetOutputPath(filename)

		if err := saveStationSearchToCSV(payloadJSON, filename, empresa, region, aor); err != nil {
			log.Fatalf("❌ Error guardando búsqueda: %v", err)
		}
		log.Printf("✅ Datos guardados en: %s", filename)

		// Generar XMLs automáticamente
		log.Println("🔄 Generando archivos XML...")
		if err := xmlcreator.CreateXMLFromFile(filename); err != nil {
			log.Fatalf("❌ Error generando XML: %v", err)
		}
	}
}

// saveDirectQueryToCSV guarda los resultados de una query directa
func saveDirectQueryToCSV(payloadJSON, filePath string) error {
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
		writer.WriteRow(record)
		return true
	})

	return nil
}

// saveStationSearchToCSV guarda los resultados de búsqueda de estación
func saveStationSearchToCSV(payloadJSON, filePath, empresa, region, aor string) error {
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
		writer.WriteRow(record)
		return true
	})

	return nil
}

// parsePath parsea un path en sus componentes
func parsePath(path string) (empresa, region, b1, b2, b3 string, err error) {
	if path == "" {
		return "", "", "", "", "", fmt.Errorf("path vacío")
	}

	parts := strings.Split(path, "/")
	if len(parts) != 5 {
		return "", "", "", "", "", fmt.Errorf("path inválido (esperados 5 partes, encontrados %d)", len(parts))
	}

	return parts[0], parts[1], parts[2], parts[3], parts[4], nil
}

// readInput lee input del usuario
func readInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
