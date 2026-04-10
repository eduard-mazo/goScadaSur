# Makefile para goScadaSur

# Variables
BINARY_NAME=goScadaSur
BUILD_DIR=dist
CMD_DIR=cmd
CONFIG_DIR=configs
WEB_DIR=web

# Versión (leer desde config.yaml)
VERSION=$(shell grep 'version:' $(CONFIG_DIR)/config.yaml | awk '{print $$2}' | tr -d '"' 2>/dev/null || echo "1.0.0")

.PHONY: help build-web build-go build clean test test-cover lint run

# Target por defecto
.DEFAULT_GOAL := help

## help: Muestra esta ayuda
help:
	@printf "\033[0;32mgoScadaSur Makefile\033[0m\n"
	@echo ""
	@echo "Targets disponibles:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## deps: Descarga dependencias de Go y Node
deps:
	@printf "\033[1;33mDescargando dependencias...\033[0m\n"
	@go mod tidy
	@cd $(WEB_DIR) && pnpm install
	@printf "\033[0;32m✓ Dependencias descargadas\033[0m\n"

## lint: Ejecuta verificadores de calidad de código (Go y Frontend)
lint:
	@printf "\033[1;33mEjecutando Lint...\033[0m\n"
	@go vet ./...
	@cd $(WEB_DIR) && pnpm run lint
	@printf "\033[0;32m✓ Verificación completada\033[0m\n"

## test: Ejecuta la suite de pruebas de integridad
test:
	@mkdir -p $(WEB_DIR)/dist
	@touch $(WEB_DIR)/dist/placeholder.txt
	@printf "\033[1;33mEjecutando suite de pruebas...\033[0m\n"
	@go test -v ./pkg/...
	@printf "\033[0;32m✓ Pruebas finalizadas con éxito\033[0m\n"

## test-cover: Ejecuta pruebas con reporte de cobertura
test-cover:
	@mkdir -p $(WEB_DIR)/dist
	@touch $(WEB_DIR)/dist/placeholder.txt
	@printf "\033[1;33mCalculando cobertura de código...\033[0m\n"
	@go test -coverprofile=coverage.out ./pkg/...
	@go tool cover -func=coverage.out
	@printf "\033[0;32m✓ Reporte generado en coverage.out\033[0m\n"

## build-web: Compila el frontend React
build-web:
	@printf "\033[1;33mCompilando Frontend...\033[0m\n"
	@cd $(WEB_DIR) && pnpm run build
	@printf "\033[0;32m✓ Frontend compilado\033[0m\n"

## build-go: Compila el binario de Go (incluye el frontend embebido)
build-go: test
	@printf "\033[1;33mCompilando Backend...\033[0m\n"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@printf "\033[0;32m✓ Backend compilado: $(BUILD_DIR)/$(BINARY_NAME)\033[0m\n"

## build: Compila toda la aplicación garantizando la integridad
build: lint test build-web build-go
	@printf "\033[1;33mEmpaquetando aplicación...\033[0m\n"
	@cp -r $(CONFIG_DIR) $(BUILD_DIR)/
	@printf "\033[0;32m✓ Aplicación robusta lista en el directorio $(BUILD_DIR)/\033[0m\n"

## run: Inicia la aplicación en modo servidor
run: build-go
	@printf "\033[1;33mIniciando $(BINARY_NAME)...\033[0m\n"
	@./$(BUILD_DIR)/$(BINARY_NAME) serve

## csv-xml: Procesa un xlsx/csv masivo y genera XMLs por estación. Uso: make csv-xml FILE=masivo_Recos.xlsx
csv-xml: build-go
	@if [ -z "$(FILE)" ]; then \
		printf "\033[0;31m✗ Debes indicar FILE=ruta/al/archivo.xlsx\033[0m\n"; exit 1; \
	fi
	@printf "\033[1;33mProcesando $(FILE)...\033[0m\n"
	@./$(BUILD_DIR)/$(BINARY_NAME) csv-xml --path $(FILE)
	@printf "\033[0;32m✓ XMLs generados en output/\033[0m\n"

## clean: Limpia archivos generados
clean:
	@printf "\033[1;33mLimpiando...\033[0m\n"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(WEB_DIR)/dist
	@rm -f coverage.out
	@printf "\033[0;32m✓ Limpieza completada\033[0m\n"