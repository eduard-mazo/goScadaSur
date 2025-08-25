package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"golang.org/x/term"
)

// Variables para almacenar las credenciales y los nuevos argumentos
var (
	user       string
	password   string
	b1, b2, b3 string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "launcher",
		Short: "Launcher para ejecutar consultas a la base de datos de Survalent.",
		Long: `Este programa actúa como una interfaz para survalentDB.exe, 
facilitando la ejecución de consultas de dos maneras:
1. station-search: Busca una estación y recupera todas sus señales analógicas y digitales.
2. direct-query: Ejecuta una consulta SQL directa en la base de datos.

Ambos modos verifican la integridad de los datos recibidos y guardan el resultado en un archivo CSV.`,
	}

	// Flags persistentes para todos los subcomandos
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "Nombre de usuario para la base de datos")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Contraseña para la base de datos")

	// --- Subcomando para buscar por estación (Modificado) ---
	var stationSearchCmd = &cobra.Command{
		Use:   "station-search",
		Short: "Busca una estación por su nombre (provisto con -b3) y recupera sus señales.",
		Long:  "Usa los flags -b1, -b2, y -b3 para especificar los parámetros de búsqueda. -b3 es obligatorio.",
		Args:  cobra.NoArgs, // Ya no acepta argumentos posicionales
		Run: func(cmd *cobra.Command, args []string) {
			// El query para este modo es el valor del flag -b3
			stationName := b3
			executeCommand("station_search", stationName)
		},
	}

	// Añadir los flags específicos para este subcomando
	stationSearchCmd.Flags().StringVar(&b1, "b1", "", "Parámetro B1 (actualmente sin uso directo)")
	stationSearchCmd.Flags().StringVar(&b2, "b2", "", "Parámetro B2 (actualmente sin uso directo)")
	stationSearchCmd.Flags().StringVar(&b3, "b3", "", "Parámetro B3, contiene el nombre de la estación a buscar")
	stationSearchCmd.MarkFlagRequired("b3") // -b3 es ahora obligatorio

	// Subcomando para consulta directa (sin cambios)
	var directQueryCmd = &cobra.Command{
		Use:   "direct-query [consulta SQL]",
		Short: "Ejecuta una consulta SQL directamente en la base de datos.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sqlQuery := args[0]
			executeCommand("direct_query", sqlQuery)
		},
	}

	rootCmd.AddCommand(stationSearchCmd, directQueryCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error ejecutando el comando:", err)
		os.Exit(1)
	}
}

// executeCommand es la función principal que maneja la ejecución del proceso C#
func executeCommand(mode, query string) {
	// Obtener credenciales si no fueron provistas por los flags
	if user == "" {
		fmt.Print("[INPUT]\tUsuario: ")
		reader := bufio.NewReader(os.Stdin)
		userInput, _ := reader.ReadString('\n')
		user = strings.TrimSpace(userInput)
	}

	if password == "" {
		fmt.Print("[INPUT]\tContraseña: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("[FATAL]\tError al leer la contraseña: %v", err)
		}
		fmt.Println()
		password = string(bytePassword)
	}

	// --- Ejecución del Proceso Externo ---
	cmd := exec.Command("./survalentDB.exe")
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	stdinPipe, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		log.Fatalf("[FATAL]\tError al iniciar el proceso: %v", err)
	}

	// Enviar datos al stdin del proceso hijo en una goroutine
	go func() {
		defer stdinPipe.Close()
		fmt.Fprintln(stdinPipe, mode)     // 1. Modo de operación
		fmt.Fprintln(stdinPipe, user)     // 2. Usuario
		fmt.Fprintln(stdinPipe, password) // 3. Contraseña
		fmt.Fprintln(stdinPipe, query)    // 4. Query o término de búsqueda
	}()

	// Capturar stdout y stderr de forma concurrente
	var wg sync.WaitGroup
	var outBuf, errBuf bytes.Buffer
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(&outBuf, stdoutPipe) }()
	go func() { defer wg.Done(); io.Copy(&errBuf, stderrPipe) }()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		log.Printf("[ERROR]\tProceso terminó con error: %v", err)
		if errBuf.Len() > 0 {
			log.Printf("stderr: %s", errBuf.String())
		}
		return
	}

	if errBuf.Len() > 0 {
		fmt.Printf("[STDERR]\n%s\n", errBuf.String())
	}

	output := outBuf.Bytes()
	if !gjson.ValidBytes(output) {
		log.Println("[WARNING]\tLa salida no es un JSON válido.")
		fmt.Printf("[OUTPUT]\n%s\n", outBuf.String())
		return
	}

	// --- Verificación de Integridad ---
	payloadJSON := gjson.GetBytes(output, "payload").Raw
	receivedChecksum := gjson.GetBytes(output, "checksum").String()

	if payloadJSON == "" || receivedChecksum == "" {
		log.Fatal("[FATAL]\tLa respuesta JSON no contiene 'payload' o 'checksum'.")
	}

	hasher := sha256.New()
	hasher.Write([]byte(payloadJSON))
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if calculatedChecksum != receivedChecksum {
		log.Printf("[FATAL]\t¡ERROR DE INTEGRIDAD! Los datos pueden estar corruptos.")
		log.Printf("\tChecksum Recibido:\t%s", receivedChecksum)
		log.Printf("\tChecksum Calculado:\t%s", calculatedChecksum)
		return
	}
	fmt.Println("[OK]\tVerificación de integridad exitosa.")

	// --- Guardado en CSV ---
	filename := time.Now().Format("20060102_150405") + "_output.csv"
	if err := saveToCSV(payloadJSON, filename); err != nil {
		log.Fatalf("[FATAL]\tError al guardar en CSV: %v", err)
	}
	fmt.Printf("\n[OK]\tDatos guardados correctamente en '%s'.\n", filename)
}

// saveToCSV convierte el JSON del payload a un archivo CSV.
func saveToCSV(payloadJSON, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("no se pudo crear el archivo: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Escribir cabeceras
	columnsResult := gjson.Get(payloadJSON, "columns.#.name")
	var headers []string
	for _, name := range columnsResult.Array() {
		headers = append(headers, name.String())
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("no se pudo escribir la cabecera: %w", err)
	}

	// Escribir filas de datos
	dataResult := gjson.Get(payloadJSON, "data")
	dataResult.ForEach(func(key, row gjson.Result) bool {
		var record []string
		for _, header := range headers {
			value := row.Get(header)
			record = append(record, value.String())
		}
		if err := writer.Write(record); err != nil {
			log.Printf("Error al escribir la fila: %v", err)
		}
		return true // continuar iterando
	})

	return writer.Error()
}
