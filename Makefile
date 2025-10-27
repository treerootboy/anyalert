.PHONY: build run test clean proto install

# Build the unified anyalert binary
build:
	go build -o bin/anyalert cmd/anyalert/*.go

# Build legacy binaries (for backward compatibility)
build-legacy:
	go build -o bin/anyalert-server cmd/server/main.go
	go build -o bin/usermgr cmd/usermgr/main.go
	go build -o bin/tokenmgr cmd/tokenmgr/main.go

# Run the server
run: build
	./bin/anyalert server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Generate proto files
proto:
	bash scripts/generate-proto.sh

# Install dependencies
install:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/anyalert-linux-amd64 cmd/anyalert/*.go
	GOOS=darwin GOARCH=amd64 go build -o bin/anyalert-darwin-amd64 cmd/anyalert/*.go
	GOOS=windows GOARCH=amd64 go build -o bin/anyalert-windows-amd64.exe cmd/anyalert/*.go
