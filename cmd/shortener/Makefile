# Variables for build information
BUILD_VERSION := 1.0.0
BUILD_DATE := $(shell date +%Y-%m-%d)
BUILD_COMMIT := $(shell git rev-parse HEAD)

# Go source files
SRC := main.go

# Output binary
OUTPUT := myapp

# LDFLAGS for passing build information
LDFLAGS := -X 'main.buildVersion=$(BUILD_VERSION)' \
           -X 'main.buildDate=$(BUILD_DATE)' \
           -X 'main.buildCommit=$(BUILD_COMMIT)'

# Default target
all: build

# Build target
build:
	@echo "Building the binary..."
	@go build -ldflags "$(LDFLAGS)" -o $(OUTPUT) $(SRC)

# Clean target
clean:
	@echo "Cleaning up..."
	@rm -f $(OUTPUT)

.PHONY: all build clean
