.PHONY: all build_shell test lint bench clean

export GOEXPERIMENT=arenas

PROJECT_NAME := "goannoy"
BUILD_DIR := "bin"
CMD_SHELL := "./cmd/shell"
TEST_DIR := "./tests"
TEST_TIMEOUT := 720s
COVERAGE_FILE := "coverage.out"

all: build_shell test lint

build: build_precision

build_precision:
	@echo "Building $(PROJECT_NAME) precision command..."
ifeq ($(HW_SUPPORT), avx256)
	go build -tags avx256 -o $(BUILD_DIR)/$(PROJECT_NAME)-precision ./cmd/precision
else ifeq ($(HW_SUPPORT), avx512)
	go build -tags avx512 -o $(BUILD_DIR)/$(PROJECT_NAME)-precision ./cmd/precision
else ifeq ($(HW_SUPPORT), neon)
	go build -tags neon -o $(BUILD_DIR)/$(PROJECT_NAME)-precision ./cmd/precision
else
	go build -o $(BUILD_DIR)/$(PROJECT_NAME)-precision ./cmd/precision
endif	

test:
	@echo "Running tests with timeout $(TEST_TIMEOUT) and generating coverage..."
	@go test -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_FILE) $(TEST_DIR)/... -v -coverpkg ./...
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
	@rm -rf tests/*.ann *.ann results.txt tests/results.txt
