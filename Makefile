.PHONY: all build_shell test lint bench clean

export GOEXPERIMENT=arenas

PROJECT_NAME := "goannoy"
BUILD_DIR := "bin"
CMD_SHELL := "./cmd/shell"
TEST_DIR := "./tests"
TEST_TIMEOUT := 60s
COVERAGE_FILE := "coverage.out"

all: build_shell test lint

build: build_shell

build_shell:
	@echo "Building $(PROJECT_NAME) shell command..."
	@go build -o $(BUILD_DIR)/$(PROJECT_NAME) $(CMD_SHELL)

test:
	@echo "Running tests with timeout $(TEST_TIMEOUT) and generating coverage..."
	@go test -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_FILE) $(TEST_DIR)/... -v
	@go tool cover -func=$(COVERAGE_FILE)

lint:
	@echo "Running lint checks..."
	@golangci-lint run ./...

bench:
	@echo "Running benchmarks..."
	@go test -bench=. -run=none ./...

clean:
	@echo "Cleaning build artifacts and coverage information..."
	@rm -rf $(BUILD_DIR) $(COVERAGE_FILE)
	@rm -rf tests/*.ann
