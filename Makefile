BINARY_NAME=graph-auth-server
MAIN_PATH=.
FRONTEND_DIR=frontend

.PHONY: all fmt doc build run clean help \
        frontend-install frontend-build frontend-dev frontend-lint frontend-format frontend-typecheck \
        dev build-all vet lint services-up services-down mailpit

all: fmt doc build

## fmt:                 Format standard Go code across the project
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...

## vet:         Run go vet static analysis on all packages
vet:
	@echo "Running Go vet..."
	@go vet ./...

## doc:                 Format/generate Swagger docs and strictly align comments
doc: fmt
	@echo "Generating Swagger docs..."
	@go generate
	@echo "Enforcing strict 20-column alignment on Swagger comments..."
	@find ./api -type f -name "*.go" -exec sh -c ' \
		for f do \
			awk '\''/^[ \t]*\/\/[ \t]*@/ { \
				sub(/^[ \t]*\/\/[ \t]*/, ""); \
				tag = $$1; \
				$$1 = ""; \
				sub(/^[ \t]+/, ""); \
				gsub(/[ \t]+/, " "); \
				if (length($$0) > 0) { \
					printf "// %-20s %s\n", tag, $$0; \
				} else { \
					printf "// %s\n", tag; \
				} \
				next; \
			} \
			{ print }'\'' "$$f" > "$$f.tmp" && mv "$$f.tmp" "$$f"; \
		done' _ {} +

## build:               Format/generate docs and compile the Go binary
build: doc
	@echo "Building binary..."
	@go build -o $(BINARY_NAME) $(MAIN_PATH)

## services-up:         Start Neo4j and Redis via Docker Compose
services-up:
	@docker compose up -d --wait
	@echo "Neo4j and Redis containers started"
	@echo "Neo4j is available at http://localhost:7474"

## services-down:       Stop Neo4j and Redis containers
services-down:
	@docker compose down

## mailpit:             Start Mailpit email testing UI (SMTP :1025, UI http://localhost:8025)
mailpit:
	@docker run -d --name mailpit -p 1025:1025 -p 8025:8025 axllent/mailpit 2>/dev/null || docker start mailpit
	@echo "Mailpit UI: http://localhost:8025"

## run:                 Start databases, format/generate docs, and run the application
run: doc services-up
	@echo "Starting application..."
	@go run $(MAIN_PATH)

## frontend-install:    Install frontend dependencies with bun
frontend-install:
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && bun install

## frontend-build:      Build the frontend for production
frontend-build:
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && bun run build

## frontend-dev:        Start the frontend Vite dev server
frontend-dev:
	@cd $(FRONTEND_DIR) && bun run dev

## dev:                 Start backend and frontend dev servers concurrently (Ctrl+C stops both)
## frontend-lint:     Lint frontend TypeScript with ESLint
frontend-lint:
	@echo "Linting frontend..."
	@cd $(FRONTEND_DIR) && bun run lint

## frontend-format:   Format frontend TypeScript with Prettier
frontend-format:
	@echo "Formatting frontend..."
	@cd $(FRONTEND_DIR) && bun run format

## frontend-typecheck: Type-check frontend TypeScript
frontend-typecheck:
	@echo "Type-checking frontend..."
	@cd $(FRONTEND_DIR) && bun run typecheck

## lint:              Run all linters (go vet + ESLint)
lint: vet frontend-lint

## dev:          Start services, backend, and frontend dev servers concurrently (Ctrl+C stops both)
dev: doc services-up
	@echo "Starting dev servers..."
	@trap 'kill 0' INT; \
		go run $(MAIN_PATH) & \
		(cd $(FRONTEND_DIR) && bun run dev) & \
		wait

## build-all:           Build frontend then compile Go binary (production)
build-all: frontend-build build

## clean:               Remove generated binaries, docs, and frontend dist
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -rf ./api/docs
	@rm -rf $(FRONTEND_DIR)/dist

## help:                Show this help message with available targets
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v fgrep | sed -e 's/\\$$//' | sed -e 's/## //'