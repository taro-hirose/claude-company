.PHONY: build build-ccs clean install test

# Build all binaries
build: build-ccs

# Build ccs binary
build-ccs:
	@echo "Building ccs binary..."
	go build -o bin/ccs ./cmd/ccs

# Build for multiple platforms
build-cross:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o bin/ccs-linux-amd64 ./cmd/ccs
	GOOS=darwin GOARCH=amd64 go build -o bin/ccs-darwin-amd64 ./cmd/ccs
	GOOS=darwin GOARCH=arm64 go build -o bin/ccs-darwin-arm64 ./cmd/ccs
	GOOS=windows GOARCH=amd64 go build -o bin/ccs-windows-amd64.exe ./cmd/ccs

# Clean build artifacts
clean:
	rm -rf bin/

# Install to local bin (requires sudo)
install: build-ccs
	sudo cp bin/ccs /usr/local/bin/

# Test
test:
	go test ./...

# Run
run: build-ccs
	./bin/ccs

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build ccs binary"
	@echo "  build-cross - Build for multiple platforms"
	@echo "  clean       - Clean build artifacts"
	@echo "  install     - Install to /usr/local/bin"
	@echo "  test        - Run tests"
	@echo "  run         - Build and run ccs"
	@echo "  help        - Show this help"