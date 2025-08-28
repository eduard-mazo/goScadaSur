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
	"quindar/pkg/xmlcreator"
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
	user, password, host, path, aor string
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
			executeCommand("station_search", "", path, aor)
		},
	}

	stationSearchCmd.Flags().StringVar(&path, "path", "", "Path del sistema B1/B2/B3")
	stationSearchCmd.Flags().StringVar(&aor, "aor", "", "Area de responsabilidad")
	stationSearchCmd.MarkFlagRequired("path")
	stationSearchCmd.MarkFlagRequired("aor")

	var directQueryCmd = &cobra.Command{
		Use:   "direct-query [consulta SQL]",
		Short: "Ejecuta una consulta SQL directamente en la base de datos.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			executeCommand("direct_query", args[0], "", "")
		},
	}

	rootCmd.AddCommand(stationSearchCmd, directQueryCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("[ERROR] Ejecución del comando:", err)
		os.Exit(1)
	}
}

// CAMBIO: La firma de la función ahora acepta b3.
func executeCommand(mode, query, path, aor string) {
	empresa, region, b1, b2, b3, err := ParsearRuta(path)
	if err != nil {
		log.Printf("Error al parsear la ruta: %v", err)
	}
	// Use the helper function for each string input, making the code DRY (Don't Repeat Yourself)
	if host == "" {
		host = readStringInput("[INPUT]\thost: ")
	}
	if path == "" {
		path = readStringInput("[INPUT]\tPath: ")
	}
	if user == "" {
		user = readStringInput("[INPUT]\tUsuario: ")
	}
	// Password input still requires special handling due to `term.ReadPassword`
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
	fmt.Printf("[DEBUG] AOR:\t%s\n", aor)
	filename := time.Now().Format("20060102_150405") + "_output.csv"
	if err := saveToCSV(payloadJSON, filename, empresa, region, aor); err != nil {
		log.Fatalf("[FATAL]\tError al guardar en CSV: %v", err)
	}

	fmt.Printf("\n[OK]\tDatos guardados correctamente en '%s'.\n", filename)

	if err := xmlcreator.CreateXML(payloadJSON, empresa, region, aor); err != nil {
		log.Fatalf("[FATAL]\tError al decodificar JSON: %v", err)
	}

}

func saveToCSV(payloadJSON, filePath, empresa, region, aor string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("no se pudo crear el archivo: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Retrieve dynamic headers from the JSON
	columnsResult := gjson.Get(payloadJSON, "columns.#.name")
	var dynamicHeaders []string
	for _, name := range columnsResult.Array() {
		dynamicHeaders = append(dynamicHeaders, name.String())
	}

	// Combine constant headers with dynamic ones
	headers := append([]string{"EMPRESA", "REGION", "AOR"}, dynamicHeaders...)

	// Write the full header row
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("no se pudo escribir la cabecera: %w", err)
	}

	// Process each data row
	dataResult := gjson.Get(payloadJSON, "data")
	dataResult.ForEach(func(key, row gjson.Result) bool {
		// Start each record with the constant values
		record := []string{empresa, region, aor}

		// Append the dynamic values from the JSON row
		for _, header := range dynamicHeaders {
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

func ParsearRuta(path string) (empresa, region, b1, b2, b3 string, err error) {
	// Divide la cadena usando el separador "/"
	parts := strings.Split(path, "/")

	// Valida que el número de partes sea exactamente 5
	if len(parts) != 5 {
		return "", "", "", "", "", fmt.Errorf("la ruta no tiene el formato esperado, se esperaban 5 partes, pero se encontraron %d", len(parts))
	}

	empresa = parts[0]
	region = parts[1]
	b1 = parts[2]
	b2 = parts[3]
	b3 = parts[4]

	return empresa, region, b1, b2, b3, nil
}

// readStringInput is a helper function to read a trimmed string from the user.
func readStringInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
