.PHONY: build run test clean proto install build-usermgr

# Build the server
build:
	go build -o bin/anyalert-server cmd/server/main.go

# Build the user manager CLI
build-usermgr:
	go build -o bin/usermgr cmd/usermgr/main.go

# Build all binaries
build-all-bins: build build-usermgr

# Run the server
run: build
	./bin/anyalert-server

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
	GOOS=linux GOARCH=amd64 go build -o bin/anyalert-server-linux-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/anyalert-server-darwin-amd64 cmd/server/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/anyalert-server-windows-amd64.exe cmd/server/main.go
