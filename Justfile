# Justfile for managing a Go codebase

# Variables
BINARY_NAME := "bookmagym"

# Directives

# Lint the code using golangci-lint
lint:
    @echo "Linting code..."
    golangci-lint run ./...

# Format the code using gofmt
fmt:
    @echo "Formatting code..."
    gofmt -w .

# Run tests
test:
    @echo "Running tests..."
    go test ./...

# Build the application
build:
    @echo "Building the application..."
    mkdir -p bin
    go build -o bin/{{BINARY_NAME}} .

# Run the application (you can modify this as needed)
run: build
    @echo "Running the application..."
    ./bin/{{BINARY_NAME}}

# Clean up the bin/ directory
clean:
    @echo "Cleaning up..."
    rm -rf bin/