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

.PHONY: help build clean test run install deps lint format check

# Target por defecto
.DEFAULT_GOAL := help

## help: Muestra esta ayuda
help:
	@printf "\033[0;32mgoScadaSur Makefile\033[0m\n"
	@echo ""
	@echo "Targets disponibles:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## deps: Descarga dependencias
deps:
	@printf "\033[1;33mDescargando dependencias...\033[0m\n"
	@go mod download
	@go mod tidy
	@printf "\033[0;32m✓ Dependencias descargadas\033[0m\n"

## build: Compila la aplicación
build: deps
	@printf "\033[1;33mCompilando $(BINARY_NAME)...\033[0m\n"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@printf "\033[0;32m✓ Compilación exitosa: $(BUILD_DIR)/$(BINARY_NAME)\033[0m\n"

## build-all: Compila para todas las plataformas
build-all: deps
	@printf "\033[1;33mCompilando para todas las plataformas...\033[0m\n"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			output_name=$(BUILD_DIR)/$(BINARY_NAME)-$$platform-$$arch; \
			if [ $$platform = "windows" ]; then output_name=$${output_name}.exe; fi; \
			printf "Compilando para $$platform/$$arch...\n"; \
			GOOS=$$platform GOARCH=$$arch go build -o $$output_name ./$(CMD_DIR); \
		done; \
	done
	@printf "\033[0;32m✓ Compilación multi-plataforma completada\033[0m\n"

## run: Ejecuta la aplicación con argumentos (ej: make run ARGS="version")
run: build
	@printf "\033[1;33mEjecutando $(BINARY_NAME)...\033[0m\n"
	@./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## test: Ejecuta tests
test:
	@printf "\033[1;33mEjecutando tests...\033[0m\n"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@printf "\033[0;32m✓ Tests completados. Ver coverage.html\033[0m\n"

## test-short: Ejecuta tests rápidos (sin -race)
test-short:
	@printf "\033[1;33mEjecutando tests rápidos...\033[0m\n"
	@go test -short ./...
	@printf "\033[0;32m✓ Tests completados\033[0m\n"

## benchmark: Ejecuta benchmarks
benchmark:
	@printf "\033[1;33mEjecutando benchmarks...\033[0m\n"
	@go test -bench=. -benchmem ./...
	@printf "\033[0;32m✓ Benchmarks completados\033[0m\n"

## lint: Ejecuta linters
lint:
	@printf "\033[1;33mEjecutando linters...\033[0m\n"
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint no encontrado. Instalando..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run ./...
	@printf "\033[0;32m✓ Linting completado\033[0m\n"

## format: Formatea el código
format:
	@printf "\033[1;33mFormateando código...\033[0m\n"
	@go fmt ./...
	@goimports -w . 2>/dev/null || true
	@printf "\033[0;32m✓ Código formateado\033[0m\n"

## check: Ejecuta verificaciones (format, lint, test)
check: format lint test
	@printf "\033[0;32m✓ Todas las verificaciones pasaron\033[0m\n"

## install: Instala la aplicación globalmente
install: build
	@printf "\033[1;33mInstalando $(BINARY_NAME)...\033[0m\n"
	@go install ./$(CMD_DIR)
	@printf "\033[0;32m✓ $(BINARY_NAME) instalado en $$GOPATH/bin\033[0m\n"

## clean: Limpia archivos generados
clean:
	@printf "\033[1;33mLimpiando...\033[0m\n"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(OUTPUT_DIR)
	@rm -f coverage.out coverage.html
	@go clean
	@printf "\033[0;32m✓ Limpieza completada\033[0m\n"

## clean-output: Limpia solo archivos de salida
clean-output:
	@printf "\033[1;33mLimpiando directorio de salida...\033[0m\n"
	@rm -rf $(OUTPUT_DIR)/*
	@printf "\033[0;32m✓ Output limpiado\033[0m\n"

## setup: Configuración inicial del proyecto
setup:
	@printf "\033[1;33mConfigurando proyecto...\033[0m\n"
	@mkdir -p $(BUILD_DIR) $(OUTPUT_DIR) $(CONFIG_DIR)
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		echo "Creando config.yaml de ejemplo..."; \
		cp $(CONFIG_DIR)/config.yaml.example $(CONFIG_DIR)/config.yaml 2>/dev/null || true; \
	fi
	@if [ ! -f $(CONFIG_DIR)/dasip_config.yaml ]; then \
		echo "Creando dasip_config.yaml de ejemplo..."; \
		cp $(CONFIG_DIR)/dasip_config.yaml.example $(CONFIG_DIR)/dasip_config.yaml 2>/dev/null || true; \
	fi
	@printf "\033[0;32m✓ Proyecto configurado\033[0m\n"

## docker-build: Construye imagen Docker
docker-build:
	@printf "\033[1;33mConstruyendo imagen Docker...\033[0m\n"
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@printf "\033[0;32m✓ Imagen Docker creada: $(BINARY_NAME):$(VERSION)\033[0m\n"

## docker-run: Ejecuta en Docker
docker-run:
	@printf "\033[1;33mEjecutando en Docker...\033[0m\n"
	@docker run --rm -v $(PWD)/$(OUTPUT_DIR):/app/$(OUTPUT_DIR) $(BINARY_NAME):latest $(ARGS)

## release: Crea release (build-all + archivos comprimidos)
release: clean build-all
	@printf "\033[1;33mCreando release v$(VERSION)...\033[0m\n"
	@mkdir -p $(BUILD_DIR)/release
	@cd $(BUILD_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [[ $$binary == *.exe ]]; then \
			zip -q release/$${binary%.exe}.zip $$binary ../README.md ../LICENSE 2>/dev/null || true; \
		else \
			tar czf release/$$binary.tar.gz $$binary ../README.md ../LICENSE 2>/dev/null || true; \
		fi; \
	done
	@printf "\033[0;32m✓ Release v$(VERSION) creado en $(BUILD_DIR)/release/\033[0m\n"

## version: Muestra la versión
version:
	@echo "$(BINARY_NAME) v$(VERSION)"

## example-csv: Genera CSV de ejemplo
example-csv:
	@printf "\033[1;33mGenerando CSV de ejemplo...\033[0m\n"
	@echo "EMPRESA,REGION,AOR,B1,B2,B3,PKEY,MIEC104,CIEC104,TYPE,ELEMENT,INFO,SBO,MLB,MMB,MHB,CLB,CMB,CHB,DASIP" > example.csv
	@echo "EPM,RORIENTE,107,M20117,LACEJA,TEST001,10111027,1,,MV,I_R,MvMoment,,1,0,0,,,,1" >> example.csv
	@echo "EPM,RORIENTE,107,M20117,LACEJA,TEST001,10111028,2,,MV,I_S,MvMoment,,2,0,0,,,,1" >> example.csv
	@printf "\033[0;32m✓ Archivo de ejemplo creado: example.csv\033[0m\n"

## docs: Genera documentación
docs:
	@printf "\033[1;33mGenerando documentación...\033[0m\n"
	@mkdir -p docs
	@godoc -http=:6060 &
	@printf "\033[0;32m✓ Documentación disponible en http://localhost:6060/pkg/goScadaSur/\033[0m\n"
	@echo "  Presiona Ctrl+C para detener"
