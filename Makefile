# Makefile para goScadaSur

# Variables
BINARY_NAME=goScadaSur
BUILD_DIR=dist
CMD_DIR=cmd
CONFIG_DIR=configs
OUTPUT_DIR=output

# Plataformas de compilación
PLATFORMS=windows linux darwin
ARCHITECTURES=amd64 arm64

# Versión (leer desde config.yaml)
VERSION=$(shell grep 'version:' $(CONFIG_DIR)/config.yaml | awk '{print $$2}' | tr -d '"')

# Colores para output
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

.PHONY: help build clean test run install deps lint format check

# Target por defecto
.DEFAULT_GOAL := help

## help: Muestra esta ayuda
help:
	@echo "$(GREEN)goScadaSur Makefile$(NC)"
	@echo ""
	@echo "Targets disponibles:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## deps: Descarga dependencias
deps:
	@echo "$(YELLOW)Descargando dependencias...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Dependencias descargadas$(NC)"

## build: Compila la aplicación
build: deps
	@echo "$(YELLOW)Compilando $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "$(GREEN)✓ Compilación exitosa: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-all: Compila para todas las plataformas
build-all: deps
	@echo "$(YELLOW)Compilando para todas las plataformas...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			output_name=$(BUILD_DIR)/$(BINARY_NAME)-$$platform-$$arch; \
			if [ $$platform = "windows" ]; then output_name=$${output_name}.exe; fi; \
			echo "Compilando para $$platform/$$arch..."; \
			GOOS=$$platform GOARCH=$$arch go build -o $$output_name ./$(CMD_DIR); \
		done; \
	done
	@echo "$(GREEN)✓ Compilación multi-plataforma completada$(NC)"

## run: Ejecuta la aplicación con argumentos (ej: make run ARGS="version")
run: build
	@echo "$(YELLOW)Ejecutando $(BINARY_NAME)...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## test: Ejecuta tests
test:
	@echo "$(YELLOW)Ejecutando tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Tests completados. Ver coverage.html$(NC)"

## test-short: Ejecuta tests rápidos (sin -race)
test-short:
	@echo "$(YELLOW)Ejecutando tests rápidos...$(NC)"
	go test -short ./...
	@echo "$(GREEN)✓ Tests completados$(NC)"

## benchmark: Ejecuta benchmarks
benchmark:
	@echo "$(YELLOW)Ejecutando benchmarks...$(NC)"
	go test -bench=. -benchmem ./...
	@echo "$(GREEN)✓ Benchmarks completados$(NC)"

## lint: Ejecuta linters
lint:
	@echo "$(YELLOW)Ejecutando linters...$(NC)"
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint no encontrado. Instalando..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...
	@echo "$(GREEN)✓ Linting completado$(NC)"

## format: Formatea el código
format:
	@echo "$(YELLOW)Formateando código...$(NC)"
	go fmt ./...
	goimports -w .
	@echo "$(GREEN)✓ Código formateado$(NC)"

## check: Ejecuta verificaciones (format, lint, test)
check: format lint test
	@echo "$(GREEN)✓ Todas las verificaciones pasaron$(NC)"

## install: Instala la aplicación globalmente
install: build
	@echo "$(YELLOW)Instalando $(BINARY_NAME)...$(NC)"
	go install ./$(CMD_DIR)
	@echo "$(GREEN)✓ $(BINARY_NAME) instalado en $$GOPATH/bin$(NC)"

## clean: Limpia archivos generados
clean:
	@echo "$(YELLOW)Limpiando...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -rf $(OUTPUT_DIR)
	rm -f coverage.out coverage.html
	go clean
	@echo "$(GREEN)✓ Limpieza completada$(NC)"

## clean-output: Limpia solo archivos de salida
clean-output:
	@echo "$(YELLOW)Limpiando directorio de salida...$(NC)"
	rm -rf $(OUTPUT_DIR)/*
	@echo "$(GREEN)✓ Output limpiado$(NC)"

## setup: Configuración inicial del proyecto
setup:
	@echo "$(YELLOW)Configurando proyecto...$(NC)"
	@mkdir -p $(BUILD_DIR) $(OUTPUT_DIR) $(CONFIG_DIR)
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		echo "Creando config.yaml de ejemplo..."; \
		cp $(CONFIG_DIR)/config.yaml.example $(CONFIG_DIR)/config.yaml 2>/dev/null || true; \
	fi
	@if [ ! -f $(CONFIG_DIR)/dasip_config.yaml ]; then \
		echo "Creando dasip_config.yaml de ejemplo..."; \
		cp $(CONFIG_DIR)/dasip_config.yaml.example $(CONFIG_DIR)/dasip_config.yaml 2>/dev/null || true; \
	fi
	@echo "$(GREEN)✓ Proyecto configurado$(NC)"

## docker-build: Construye imagen Docker
docker-build:
	@echo "$(YELLOW)Construyendo imagen Docker...$(NC)"
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "$(GREEN)✓ Imagen Docker creada: $(BINARY_NAME):$(VERSION)$(NC)"

## docker-run: Ejecuta en Docker
docker-run:
	@echo "$(YELLOW)Ejecutando en Docker...$(NC)"
	docker run --rm -v $(PWD)/$(OUTPUT_DIR):/app/$(OUTPUT_DIR) $(BINARY_NAME):latest $(ARGS)

## release: Crea release (build-all + archivos comprimidos)
release: clean build-all
	@echo "$(YELLOW)Creando release v$(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@cd $(BUILD_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [[ $$binary == *.exe ]]; then \
			zip -q release/$${binary%.exe}.zip $$binary ../README.md ../LICENSE 2>/dev/null || true; \
		else \
			tar czf release/$$binary.tar.gz $$binary ../README.md ../LICENSE 2>/dev/null || true; \
		fi; \
	done
	@echo "$(GREEN)✓ Release v$(VERSION) creado en $(BUILD_DIR)/release/$(NC)"

## version: Muestra la versión
version:
	@echo "$(BINARY_NAME) v$(VERSION)"

## example-csv: Genera CSV de ejemplo
example-csv:
	@echo "$(YELLOW)Generando CSV de ejemplo...$(NC)"
	@echo "EMPRESA,REGION,AOR,B1,B2,B3,PKEY,MIEC104,CIEC104,TYPE,ELEMENT,INFO,SBO,MLB,MMB,MHB,CLB,CMB,CHB,DASIP" > example.csv
	@echo "EPM,RORIENTE,107,M20117,LACEJA,TEST001,10111027,1,,MV,I_R,MvMoment,,1,0,0,,,,1" >> example.csv
	@echo "EPM,RORIENTE,107,M20117,LACEJA,TEST001,10111028,2,,MV,I_S,MvMoment,,2,0,0,,,,1" >> example.csv
	@echo "$(GREEN)✓ Archivo de ejemplo creado: example.csv$(NC)"

## docs: Genera documentación
docs:
	@echo "$(YELLOW)Generando documentación...$(NC)"
	@mkdir -p docs
	godoc -http=:6060 &
	@echo "$(GREEN)✓ Documentación disponible en http://localhost:6060/pkg/goScadaSur/$(NC)"
	@echo "  Presiona Ctrl+C para detener"
