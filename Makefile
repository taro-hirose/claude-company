.PHONY: build build-storm build-deploy clean install test

# Build all binaries
build: build-storm build-deploy

# Build storm binary (tmux session manager)
build-storm:
	@echo "⚡ Building storm binary..."
	@mkdir -p bin
	go build -o bin/storm ./cmd/ccs

# Build deploy binary (AI task manager)
build-deploy:
	@echo "🚀 Building deploy binary..."
	@mkdir -p bin
	go build -o bin/deploy .

# Build for multiple platforms
build-cross:
	@echo "🌍 Building for multiple platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/storm-linux-amd64 ./cmd/ccs
	GOOS=darwin GOARCH=amd64 go build -o bin/storm-darwin-amd64 ./cmd/ccs
	GOOS=darwin GOARCH=arm64 go build -o bin/storm-darwin-arm64 ./cmd/ccs
	GOOS=windows GOARCH=amd64 go build -o bin/storm-windows-amd64.exe ./cmd/ccs
	GOOS=linux GOARCH=amd64 go build -o bin/deploy-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o bin/deploy-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o bin/deploy-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o bin/deploy-windows-amd64.exe .

# Clean build artifacts
clean:
	rm -rf bin/

# Install to local bin (requires sudo)
install: build
	sudo cp bin/storm /usr/local/bin/
	sudo cp bin/deploy /usr/local/bin/

# Test
test:
	go test ./...

# Help
help:
	@echo "🌪️ Claude Company Tools"
	@echo "Available targets:"
	@echo "  build         - Build all binaries (storm + deploy)"
	@echo "  build-storm   - Build storm (tmux session manager)"
	@echo "  build-deploy  - Build deploy (AI task manager)"
	@echo "  build-cross   - Build for multiple platforms"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install to /usr/local/bin"
	@echo "  test          - Run tests"
	@echo "  help          - Show this help"