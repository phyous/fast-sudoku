.PHONY: all build test clean

# Build all applications
all: build test run

# Build API server
build:
	CGO_ENABLED=0 go build -o bin/simulator ./cmd/simulator

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

run: build
	./bin/simulator