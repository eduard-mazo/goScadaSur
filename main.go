package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"goScadaSur/pkg/xmlcreator"
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

// CSharpInput defines the structure for sending data to the C# application.
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

var user, password, host, path, aor string

func main() {
	rootCmd := &cobra.Command{
		Use:   "launcher",
		Short: "Gestion base de datos SURVALENT",
	}

	rootCmd.PersistentFlags().StringVarP(&host, "host", "i", "", "Host IP address")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "Database username")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Database password")

	stationSearchCmd := &cobra.Command{
		Use:   "station-search",
		Short: "Busca una estación por el nombre (suministrado por el --path) y retorna sus señales.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			executeCommand("station_search", "", path, aor)
		},
	}

	stationSearchCmd.Flags().StringVar(&path, "path", "", "Path del sistema (e.j., B1/B2/B3)")
	stationSearchCmd.Flags().StringVar(&aor, "aor", "", "Area of Responsibility")
	if err := stationSearchCmd.MarkFlagRequired("path"); err != nil {
		log.Fatalf("[FATAL] Error marking 'path' flag as required: %v", err)
	}
	if err := stationSearchCmd.MarkFlagRequired("aor"); err != nil {
		log.Fatalf("[FATAL] Error marking 'aor' flag as required: %v", err)
	}

	fromCSVCmd := &cobra.Command{
		Use:   "csv-xml",
		Short: "Crea los XML a partir de un CSV",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			createFromCSV(path)
		},
	}

	fromCSVCmd.Flags().StringVar(&path, "path", "", "Ruta del csv")
	fromCSVCmd.Flags().StringVar(&aor, "aor", "", "Area de responsabilidad")
	if err := fromCSVCmd.MarkFlagRequired("path"); err != nil {
		log.Fatalf("[FATAL] Error marking 'path' flag as required: %v", err)
	}
	if err := fromCSVCmd.MarkFlagRequired("aor"); err != nil {
		log.Fatalf("[FATAL] Error marking 'aor' flag as required: %v", err)
	}

	directQueryCmd := &cobra.Command{
		Use:   "direct-query [SQL query]",
		Short: "Ejecuta una query directamente la en la base de datos de SURVALENT",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			executeCommand("direct_query", args[0], "", "")
		},
	}

	rootCmd.AddCommand(stationSearchCmd, directQueryCmd, fromCSVCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("[ERROR] Command execution failed:", err)
		os.Exit(1)
	}
}

func createFromCSV(path string) {
	if err := xmlcreator.CreateXML(path); err != nil {
		log.Fatalf("[FATAL]\tError creating XML: %v", err)
	}
}

func executeCommand(mode, query, path, aor string) {
	empresa, region, b1, b2, b3, err := ParsearRuta(path)
	if err != nil && mode != "direct_query" {
		log.Printf("Error parsing path: %v", err)
	}

	if host == "" {
		host = readStringInput("[INPUT]\tHost: ")
	}

	if user == "" {
		user = readStringInput("[INPUT]\tUser: ")
	}

	if password == "" {
		fmt.Print("[INPUT]\tPassword: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("[FATAL]\tError reading password: %v", err)
		}
		fmt.Println()
		password = string(bytePassword)
	}

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

	inputBytes, err := json.Marshal(inputData)
	if err != nil {
		log.Fatalf("[FATAL]\tError creating input JSON: %v", err)
	}
	inputJsonString := string(inputBytes)

	cmd := exec.Command("./survalentDB.exe")
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	stdinPipe, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		log.Fatalf("[FATAL]\tError starting process: %v", err)
	}

	go func() {
		defer stdinPipe.Close()
		fmt.Fprintln(stdinPipe, inputJsonString)
	}()

	var wg sync.WaitGroup
	var outBuf, errBuf bytes.Buffer
	wg.Add(2)

	go func() {
		defer wg.Done()
		if _, err := io.Copy(&outBuf, stdoutPipe); err != nil {
			log.Printf("[ERROR] Failed to read stdout: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if _, err := io.Copy(&errBuf, stderrPipe); err != nil {
			log.Printf("[ERROR] Failed to read stderr: %v", err)
		}
	}()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		log.Printf("[ERROR]\tProcess finished with error: %v", err)
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
		log.Println("[WARNING]\tOutput is not valid JSON.")
		fmt.Printf("[OUTPUT]\n%s\n", outBuf.String())
		return
	}

	payloadJSON := gjson.GetBytes(output, "payload").Raw
	receivedChecksum := gjson.GetBytes(output, "checksum").String()
	if payloadJSON == "" || receivedChecksum == "" {
		log.Fatal("[FATAL]\tJSON response is missing 'payload' or 'checksum'.")
	}

	hasher := sha256.New()
	hasher.Write([]byte(payloadJSON))
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if calculatedChecksum != receivedChecksum {
		log.Printf("[FATAL]\tINTEGRITY ERROR! Data may be corrupt.")
		return
	}
	fmt.Println("[OK]\tIntegrity check successful.")

	switch mode {
	case "direct_query":
		filename := time.Now().Format("20060102_150405") + "_direct_query.csv"
		if err := saveDirectQueryToCSV(payloadJSON, filename); err != nil {
			log.Fatalf("[FATAL]\tError saving direct query to CSV: %v", err)
		}
		fmt.Printf("\n[OK]\tDirect query data saved to '%s'.\n", filename)

	case "station_search":
		filename := time.Now().Format("20060102_150405") + "_" + b3 + ".csv"
		if err := saveToCSV(payloadJSON, filename, empresa, region, aor); err != nil {
			log.Fatalf("[FATAL]\tError saving to CSV: %v", err)
		}
		fmt.Printf("\n[OK]\tData saved successfully to '%s'.\n", filename)

		if err := xmlcreator.CreateXML(filename); err != nil {
			log.Fatalf("[FATAL]\tError creating XML: %v", err)
		}
	}
}

func saveDirectQueryToCSV(payloadJSON, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
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
		return fmt.Errorf("could not write header: %w", err)
	}

	dataResult := gjson.Get(payloadJSON, "data")
	dataResult.ForEach(func(key, row gjson.Result) bool {
		var record []string
		for _, header := range headers {
			value := row.Get(header)
			record = append(record, value.String())
		}
		if err := writer.Write(record); err != nil {
			log.Printf("Error writing row: %v", err)
		}
		return true
	})
	return writer.Error()
}

func saveToCSV(payloadJSON, filePath, empresa, region, aor string) error {
	fmt.Printf("AOR\t%s\tEMPRESA:\t%s\tREGION:\t%s", aor, empresa, region)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	columnsResult := gjson.Get(payloadJSON, "columns.#.name")
	var dynamicHeaders []string
	for _, name := range columnsResult.Array() {
		dynamicHeaders = append(dynamicHeaders, name.String())
	}

	headers := append([]string{"EMPRESA", "REGION", "AOR"}, dynamicHeaders...)
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("could not write header: %w", err)
	}

	dataResult := gjson.Get(payloadJSON, "data")
	dataResult.ForEach(func(key, row gjson.Result) bool {
		record := []string{empresa, region, aor}
		for _, header := range dynamicHeaders {
			value := row.Get(header)
			record = append(record, value.String())
		}
		if err := writer.Write(record); err != nil {
			log.Printf("Error writing row: %v", err)
		}
		return true
	})
	return writer.Error()
}

func ParsearRuta(path string) (empresa, region, b1, b2, b3 string, err error) {
	if path == "" {
		return "", "", "", "", "", fmt.Errorf("path is empty")
	}
	parts := strings.Split(path, "/")
	if len(parts) != 5 {
		return "", "", "", "", "", fmt.Errorf("path does not have the expected format; expected 5 parts, but found %d", len(parts))
	}

	empresa = parts[0]
	region = parts[1]
	b1 = parts[2]
	b2 = parts[3]
	b3 = parts[4]

	return empresa, region, b1, b2, b3, nil
}

func readStringInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
