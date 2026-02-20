// cmd/main.go
package main

import (
	"bufio"
	"fmt"
	"goScadaSur/pkg/api"
	"goScadaSur/pkg/config"
	"goScadaSur/pkg/database"
	"goScadaSur/pkg/utils"
	"goScadaSur/pkg/xmlcreator"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	defaultConfigPath = "configs/config.yaml"
)

var (
	// Flags globales
	configFile string
	user       string
	password   string
	host       string
	path       string
	aor        string
	port       int

	// Configuración y estado global de la ejecución
	appCfg   *config.AppConfig
	dasipCfg *config.DasipConfig
	tm       *xmlcreator.TemplateManager
	dbClient *database.DatabaseClient
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

	// Flags persistentes
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", defaultConfigPath, "Archivo de configuración")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "i", "", "Dirección IP del host")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "Usuario de la base de datos")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Contraseña de la base de datos")

	// Comando: station-search
	stationSearchCmd := &cobra.Command{
		Use:   "station-search",
		Short: "Busca una estación por nombre y retorna sus señales",
		Args:  cobra.NoArgs,
		Run:   runStationSearch,
	}
	stationSearchCmd.Flags().StringVar(&path, "path", "", "Path del sistema (ej: B1/B2/B3)")
	stationSearchCmd.Flags().StringVar(&aor, "aor", "", "Área de responsabilidad")

	// Comando: csv-xml
	csvXmlCmd := &cobra.Command{
		Use:   "csv-xml",
		Short: "Genera archivos XML desde CSV o Excel",
		Args:  cobra.NoArgs,
		Run:   runCSVToXML,
	}
	csvXmlCmd.Flags().StringVar(&path, "path", "", "Ruta del archivo CSV/Excel")
	csvXmlCmd.Flags().StringVar(&aor, "aor", "", "Área de responsabilidad")

	// Comando: direct-query
	directQueryCmd := &cobra.Command{
		Use:   "direct-query [SQL query]",
		Short: "Ejecuta una query SQL directamente",
		Args:  cobra.ExactArgs(1),
		Run:   runDirectQuery,
	}

	// Comando: serve (NUEVO)
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Inicia el servidor web para la interfaz de usuario",
		Args:  cobra.NoArgs,
		Run:   runServe,
	}
	serveCmd.Flags().IntVarP(&port, "port", "P", 8080, "Puerto para el servidor web")

	// Comando: version
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Muestra la versión de la aplicación",
		Run: func(cmd *cobra.Command, args []string) {
			if appCfg != nil {
				fmt.Printf("%s v%s\n", appCfg.App.Name, appCfg.App.Version)
			} else {
				fmt.Println("goScadaSur - versión no disponible")
			}
		},
	}

	rootCmd.AddCommand(stationSearchCmd, directQueryCmd, csvXmlCmd, serveCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
}

func initializeApp(cmd *cobra.Command, args []string) {
	var err error
	appCfg, err = config.Load(configFile)
	if err != nil {
		log.Fatalf("[ERROR] Error cargando configuración: %v", err)
	}

	dasipConfigPath := appCfg.GetDasipConfigPath()
	dasipCfg, err = config.LoadDasipConfig(dasipConfigPath)
	if err != nil {
		dasipCfg = &config.DasipConfig{DefaultPath: "SCADA/RTU", DasipMapping: make(map[string]string)}
	}

	templatesPath := appCfg.GetTemplatesPath()
	tm, err = xmlcreator.LoadTemplates(templatesPath)
	if err != nil {
		log.Printf("[WARN] Error cargando plantillas: %v", err)
	}

	dbClient = database.NewDatabaseClient(appCfg)

	if err := appCfg.EnsureOutputDir(); err != nil {
		log.Printf("[WARN] Error creando directorio de salida: %v", err)
	}
}

func runServe(cmd *cobra.Command, args []string) {
	server := api.NewServer(appCfg, dasipCfg, tm, configFile)

	// Abrir navegador automáticamente
	go func() {
		time.Sleep(500 * time.Millisecond) // Pequeña espera para asegurar que el server esté escuchando
		url := fmt.Sprintf("http://localhost:%d", port)
		log.Printf("[INFO] Abriendo navegador en %s...", url)
		if err := utils.OpenBrowser(url); err != nil {
			log.Printf("[WARN] No se pudo abrir el navegador: %v", err)
		}
	}()

	if err := server.Start(port); err != nil {
		log.Fatalf("[ERROR] Error iniciando servidor: %v", err)
	}
}

func runStationSearch(cmd *cobra.Command, args []string) {
	empresa, region, b1, b2, b3, err := database.ParsePath(path)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	if host == "" { host = readInput("Host: ") }
	if user == "" { user = readInput("Usuario: ") }
	if password == "" { password = readPassword() }

	input := database.CSharpInput{
		Mode: "station_search", User: user, Password: password, Host: host,
		B1: b1, B2: b2, B3: b3,
	}

	result, err := dbClient.ExecuteCommand(input)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	timestamp := time.Now().Format(appCfg.Output.TimestampFormat)
	filename := fmt.Sprintf("%s_%s%s", timestamp, b3, appCfg.Output.Suffixes["csv"])
	filename = appCfg.GetOutputPath(filename)

	if err := database.SaveStationSearchToCSV(result.PayloadJSON, filename, empresa, region, aor); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	log.Printf("[OK] Datos guardados en: %s", filename)

	if err := xmlcreator.CreateXMLFromFile(filename, appCfg, dasipCfg, tm); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
}

func runDirectQuery(cmd *cobra.Command, args []string) {
	query := args[0]
	if host == "" { host = readInput("Host: ") }
	if user == "" { user = readInput("Usuario: ") }
	if password == "" { password = readPassword() }

	input := database.CSharpInput{
		Mode: "direct_query", User: user, Password: password, Host: host,
		Query: query,
	}

	result, err := dbClient.ExecuteCommand(input)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	timestamp := time.Now().Format(appCfg.Output.TimestampFormat)
	filename := fmt.Sprintf("%s_direct_query%s", timestamp, appCfg.Output.Suffixes["csv"])
	filename = appCfg.GetOutputPath(filename)

	if err := database.SaveDirectQueryToCSV(result.PayloadJSON, filename); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	log.Printf("[OK] Resultados guardados en: %s", filename)
}

func runCSVToXML(cmd *cobra.Command, args []string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("[ERROR] El archivo '%s' no existe", path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	ext = strings.TrimPrefix(ext, ".")

	if !appCfg.IsFormatSupported(ext) {
		log.Fatalf("[ERROR] Formato '%s' no soportado.", ext)
	}

	if err := xmlcreator.CreateXMLFromFile(path, appCfg, dasipCfg, tm); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	log.Println("[OK] Proceso completado")
}

func readInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func readPassword() string {
	fmt.Print("Contraseña: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	fmt.Println()
	return string(bytePassword)
}
