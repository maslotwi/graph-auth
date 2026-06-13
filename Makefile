BINARY_NAME=graph-auth-server
MAIN_PATH=.

.PHONY: all fmt doc build run clean help

all: fmt doc build

## fmt:         Format standard Go code across the project
fmt:
	@echo "Formatting Go code..."
	@go fmt

## doc:         Format/generate Swagger docs and strictly align comments
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

## build:       Format/generate docs and compile the Go binary
build: doc
	@echo "Building binary..."
	@go build -o $(BINARY_NAME) $(MAIN_PATH)

## run:         Format/generate docs and run the application instantly
run: doc
	@echo "Starting application..."
	@go run $(MAIN_PATH)

## clean:       Remove generated binaries and documentation files
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -rf ./api/docs

## help:        Show this help message with available targets
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v fgrep | sed -e 's/\\$$//' | sed -e 's/## //'