// pkg/database/interop.go
package database

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"goScadaSur/pkg/config"
	"io"
	"log"
	"os/exec"
	"sync"

	"github.com/tidwall/gjson"
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

// InteropResult representa el resultado de una operación en la base de datos
type InteropResult struct {
	PayloadJSON string
	Checksum    string
}

// DatabaseClient gestiona la comunicación con la aplicación C#
type DatabaseClient struct {
	AppCfg *config.AppConfig
}

// NewDatabaseClient crea un nuevo cliente de base de datos
func NewDatabaseClient(cfg *config.AppConfig) *DatabaseClient {
	return &DatabaseClient{
		AppCfg: cfg,
	}
}

// ExecuteCommand ejecuta un comando en la base de datos a través de la herramienta C#
func (c *DatabaseClient) ExecuteCommand(input CSharpInput) (*InteropResult, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error marshalling input: %w", err)
	}
	inputJSON := string(inputBytes)

	// Ejecutar proceso C#
	csharpExe := c.AppCfg.Database.CSharpExecutable
	cmd := exec.Command(csharpExe)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	stdinPipe, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error iniciando proceso C#: %w", err)
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
		if _, err := io.Copy(&outBuf, stdoutPipe); err != nil {
			log.Printf("[ERROR] Error leyendo stdout: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if _, err := io.Copy(&errBuf, stderrPipe); err != nil {
			log.Printf("[ERROR] Error leyendo stderr: %v", err)
		}
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if errBuf.Len() > 0 {
			return nil, fmt.Errorf("error en proceso C#: %v (stderr: %s)", err, errBuf.String())
		}
		return nil, fmt.Errorf("error en proceso C#: %w", err)
	}

	// Procesar salida
	output := outBuf.Bytes()
	if !gjson.ValidBytes(output) {
		return nil, fmt.Errorf("la salida de C# no es JSON válido: %s", outBuf.String())
	}

	// Verificar integridad
	payloadJSON := gjson.GetBytes(output, "payload").Raw
	receivedChecksum := gjson.GetBytes(output, "checksum").String()

	if payloadJSON == "" || receivedChecksum == "" {
		return nil, fmt.Errorf("respuesta JSON de C# inválida (falta payload o checksum)")
	}

	hasher := sha256.New()
	hasher.Write([]byte(payloadJSON))
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if calculatedChecksum != receivedChecksum {
		return nil, fmt.Errorf("error de integridad: los datos de C# pueden estar corruptos")
	}

	return &InteropResult{
		PayloadJSON: payloadJSON,
		Checksum:    receivedChecksum,
	}, nil
}
