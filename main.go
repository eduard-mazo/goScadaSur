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
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"golang.org/x/term"
)

var (
	devMode bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "launcher",
		Short: "Ejecuta ConsoleApp1.exe, verifica integridad y guarda en CSV",
		Run: func(cmd *cobra.Command, args []string) {
			runApp(devMode)
		},
	}

	rootCmd.Flags().BoolVar(&devMode, "devMode", false, "Ejecutar en modo desarrollo con credenciales por defecto")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error ejecutando comando:", err)
		os.Exit(1)
	}
}

func runApp(devMode bool) {
	var user, pass, query string

	if devMode {
		user = "admin"
		pass = ".Qwe123456789"
		query = "SELECT pkey FROM AnalogPoints WHERE pkey >= 300000 AND pkey <= 300000" // Query más pequeña para desarrollo
		fmt.Println("[DEV]\tModo desarrollo activado.")
	} else {
		// Lógica para pedir datos al usuario (sin cambios)
		fmt.Print("[INPUT]\tUsuario: ")
		reader := bufio.NewReader(os.Stdin)
		userInput, _ := reader.ReadString('\n')
		user = userInput[:len(userInput)-1]

		fmt.Print("[INPUT]\tContraseña: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatal("Error al leer la contraseña")
		}
		fmt.Println()
		pass = string(bytePassword)

		fmt.Print("[INPUT]\tQuery: ")
		reader = bufio.NewReader(os.Stdin)
		userInput, _ = reader.ReadString('\n')
		query = userInput[:len(userInput)-1]
	}

	cmd := exec.Command("./survalentDB.exe")

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	stdinPipe, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		log.Fatalf("[FATAL]\tError al iniciar el proceso: %v", err)
	}

	go func() {
		defer stdinPipe.Close()
		fmt.Fprintln(stdinPipe, user)
		fmt.Fprintln(stdinPipe, pass)
		fmt.Fprintln(stdinPipe, query)
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
			log.Printf("stderr: %s", errBuf.String())
		}
		return
	}

	if errBuf.Len() > 0 {
		fmt.Printf("[STDERR]\n%s\n", errBuf.String())
	}

	output := outBuf.Bytes()
	if !gjson.ValidBytes(output) {
		log.Printf("[WARNING]\tLa salida no es un JSON válido.\n")
		return
	}

	// --- VERIFICACIÓN DE INTEGRIDAD ---
	payloadJSON := gjson.GetBytes(output, "payload").Raw
	receivedChecksum := gjson.GetBytes(output, "checksum").String()

	if payloadJSON == "" || receivedChecksum == "" {
		log.Fatal("[FATAL]\tLa respuesta no contiene 'payload' o 'checksum'.")
	}

	hasher := sha256.New()
	hasher.Write([]byte(payloadJSON))
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if calculatedChecksum != receivedChecksum {
		log.Printf("\t[FATAL]\t¡ERROR DE INTEGRIDAD! Los datos pueden estar corruptos.")
		log.Printf("\tChecksum Recibido:\t%s", receivedChecksum)
		log.Printf("\tChecksum Calculado:\t%s", calculatedChecksum)
		fmt.Println(string(output))
		return
	}
	fmt.Println("[OK]\tVerificación de integridad exitosa.")

	// --- IMPRIMIR DATOS SOLICITADOS ---
	firstResult := gjson.Get(payloadJSON, "data.0")
	firstColumn := gjson.Get(payloadJSON, "columns.0")

	fmt.Println("\n--- PRIMER RESULTADO (data[0]) ---")
	fmt.Println(firstResult.String())
	fmt.Println("\n--- PRIMERA COLUMNA (columns[0]) ---")
	fmt.Println(firstColumn.String())

	// --- GUARDAR EN CSV ---
	t := time.Now().Format("20060102_150405") // AAAAMMDD_HHMMSS
	filename := t + "_output.csv"

	if err := saveToCSV(payloadJSON, filename); err != nil {
		log.Fatalf("\t[FATAL]\tError al guardar en CSV: %v", err)
	}
	fmt.Printf("\n[OK]\tDatos guardados correctamente en 'output.csv'.\n")
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

	// Escribir filas
	dataResult := gjson.Get(payloadJSON, "data")
	for _, row := range dataResult.Array() {
		var record []string
		// Iterar en el orden de las cabeceras para mantener la consistencia
		for _, header := range headers {
			// gjson permite escapar caracteres especiales en los nombres de campo
			value := gjson.Get(row.Raw, gjson.Escape(header))
			record = append(record, value.String())
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("no se pudo escribir la fila: %w", err)
		}
	}

	return nil
}
