package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
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

// Estructura para enviar los datos a C# como un solo JSON
// Se añade el campo B3 para un contrato de datos explícito.
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
	user, password, host, b1, b2, b3 string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "launcher",
		Short: "Launcher para ejecutar consultas a la base de datos de Survalent.",
	}

	rootCmd.PersistentFlags().StringVarP(&host, "host", "i", "", "IP del host")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "Nombre de usuario para la base de datos")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Contraseña para la base de datos")

	var stationSearchCmd = &cobra.Command{
		Use:   "station-search",
		Short: "Busca una estación por su nombre (provisto con -b3) y recupera sus señales.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// CAMBIO: Se pasa b3 en su propio parámetro y la query va vacía.
			executeCommand("station_search", "", b1, b2, b3)
		},
	}

	stationSearchCmd.Flags().StringVar(&b1, "b1", "", "Parámetro B1")
	stationSearchCmd.Flags().StringVar(&b2, "b2", "", "Parámetro B2")
	stationSearchCmd.Flags().StringVar(&b3, "b3", "", "Parámetro B3, contiene el nombre de la estación a buscar")
	stationSearchCmd.MarkFlagRequired("b3")

	var directQueryCmd = &cobra.Command{
		Use:   "direct-query [consulta SQL]",
		Short: "Ejecuta una consulta SQL directamente en la base de datos.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			executeCommand("direct_query", args[0], "", "", "")
		},
	}

	rootCmd.AddCommand(stationSearchCmd, directQueryCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error ejecutando el comando:", err)
		os.Exit(1)
	}
}

// CAMBIO: La firma de la función ahora acepta b3.
func executeCommand(mode, query, b1, b2, b3 string) {
	if host == "" {
		fmt.Print("[INPUT]\thost: ")
		reader := bufio.NewReader(os.Stdin)
		hostInput, _ := reader.ReadString('\n')
		host = strings.TrimSpace(hostInput)
	}
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

	// Crear la estructura de entrada
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

	// Convertir la estructura a un string JSON
	inputBytes, err := json.Marshal(inputData)
	if err != nil {
		log.Fatalf("[FATAL]\tError al crear el JSON de entrada: %v", err)
	}
	inputJsonString := string(inputBytes)

	// --- Ejecución del Proceso Externo ---
	cmd := exec.Command("./survalentDB.exe")
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	stdinPipe, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		log.Fatalf("[FATAL]\tError al iniciar el proceso: %v", err)
	}

	// Enviar la única línea de JSON al stdin del proceso hijo
	go func() {
		defer stdinPipe.Close()
		fmt.Fprintln(stdinPipe, inputJsonString)
	}()

	var wg sync.WaitGroup
	var outBuf, errBuf bytes.Buffer
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(&outBuf, stdoutPipe) }()
	go func() { defer wg.Done(); io.Copy(&errBuf, stderrPipe) }()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		log.Printf("[ERROR]\tProceso terminó con error: %v", err)
		if errBuf.Len() > 0 {
			log.Printf("[STDERR]: %s", errBuf.String())
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
		return
	}
	fmt.Println("[OK]\tVerificación de integridad exitosa.")

	filename := time.Now().Format("20060102_150405") + "_output.csv"
	if err := saveToCSV(payloadJSON, filename); err != nil {
		log.Fatalf("[FATAL]\tError al guardar en CSV: %v", err)
	}
	fmt.Printf("\n[OK]\tDatos guardados correctamente en '%s'.\n", filename)
}

func saveToCSV(payloadJSON, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("no se pudo crear el archivo: %w", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	columnsResult := gjson.Get(payloadJSON, "columns.#.name")
	var headers []string
	for _, name := range columnsResult.Array() {
		headers = append(headers, name.String())
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("no se pudo escribir la cabecera: %w", err)
	}
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
		return true
	})
	return writer.Error()
}
