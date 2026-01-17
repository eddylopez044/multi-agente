.PHONY: build test clean run lint fmt help

# Variables
BINARY_NAME=multi-agent
CLI_BINARY=multi-agent-cli
ORCHESTRATOR_BINARY=orchestrator

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Construye los binarios
	@echo "Building binaries..."
	@go build -o bin/$(ORCHESTRATOR_BINARY) ./cmd/orchestrator
	@go build -o bin/$(CLI_BINARY) ./cmd/cli
	@echo "Build complete!"

test: ## Ejecuta los tests
	@echo "Running tests..."
	@go test -v -cover ./...

lint: ## Ejecuta el linter
	@echo "Running linter..."
	@go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

fmt: ## Formatea el código
	@echo "Formatting code..."
	@go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi

run: build ## Ejecuta el orchestrator con un ejemplo
	@./bin/$(ORCHESTRATOR_BINARY) --task "optimize endpoint /api/search" --repo .

clean: ## Limpia los binarios
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf artifacts/
	@echo "Clean complete!"

deps: ## Instala las dependencias
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

example: ## Ejecuta el ejemplo simple
	@go run examples/simple/main.go

install-tools: ## Instala herramientas de desarrollo
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
