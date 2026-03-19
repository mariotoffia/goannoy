.PHONY: all build_shell test lint bench bench-scalar bench-accel-arm64 bench-accel-amd64 bench-compare clean

export GOEXPERIMENT=arenas

PROJECT_NAME := "goannoy"
BUILD_DIR := "bin"
CMD_SHELL := "./cmd/shell"
TEST_DIR := "./tests"
TEST_TIMEOUT := 720s
COVERAGE_FILE := "coverage.out"
BENCH_DIR := "bench-results"
BENCH_PACKAGES := ./vector ./tests
BENCH_PATTERN := Benchmark(DotFloat32Backends|AccelerationBuild|AccelerationSearch)$$

all: build_shell test lint

build: build_shell

build_shell:
	@echo "Building $(PROJECT_NAME) shell command..."
	@go build -o $(BUILD_DIR)/$(PROJECT_NAME) $(CMD_SHELL)

test:
	@echo "Running tests with timeout $(TEST_TIMEOUT) and generating coverage..."
	@go test -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_FILE) $(TEST_DIR)/... -v -coverpkg ./...
	@go tool cover -func=$(COVERAGE_FILE)

lint:
	@echo "Running lint checks..."
	@golangci-lint run ./...

bench:
	@echo "Running benchmarks..."
	@go test -bench='$(BENCH_PATTERN)' -run=none $(BENCH_PACKAGES)

bench-scalar:
	@mkdir -p $(BENCH_DIR)
	@out="$(BENCH_DIR)/scalar-$$(date +%Y%m%d-%H%M%S).txt"; \
	echo "Running scalar benchmarks..."; \
	go test -bench='$(BENCH_PATTERN)' -run=none $(BENCH_PACKAGES) | tee "$$out"; \
	echo "Scalar benchmark output: $$out"

bench-accel-arm64:
	@arch=$$(go env GOARCH); \
	if [ "$$arch" != "arm64" ]; then \
		echo "bench-accel-arm64 must run on an arm64 host"; \
		exit 1; \
	fi; \
	mkdir -p $(BENCH_DIR); \
	out="$(BENCH_DIR)/accel-arm64-$$(date +%Y%m%d-%H%M%S).txt"; \
	echo "Running arm64 accelerated benchmarks..."; \
	go test -tags accelerate -bench='$(BENCH_PATTERN)' -run=none $(BENCH_PACKAGES) | tee "$$out"; \
	echo "arm64 accelerated benchmark output: $$out"

bench-accel-amd64:
	@arch=$$(go env GOARCH); \
	if [ "$$arch" != "amd64" ]; then \
		echo "bench-accel-amd64 must run on an amd64 host"; \
		exit 1; \
	fi; \
	mkdir -p $(BENCH_DIR); \
	out="$(BENCH_DIR)/accel-amd64-$$(date +%Y%m%d-%H%M%S).txt"; \
	echo "Running amd64 accelerated benchmarks with GOEXPERIMENT=arenas,simd..."; \
	GOEXPERIMENT=arenas,simd go test -tags accelerate -bench='$(BENCH_PATTERN)' -run=none $(BENCH_PACKAGES) | tee "$$out"; \
	echo "amd64 accelerated benchmark output: $$out"

bench-compare:
	@mkdir -p $(BENCH_DIR); \
	ts=$$(date +%Y%m%d-%H%M%S); \
	scalar="$(BENCH_DIR)/scalar-$$ts.txt"; \
	accel="$(BENCH_DIR)/accelerated-$$ts.txt"; \
	echo "Running scalar benchmarks..."; \
	go test -bench='$(BENCH_PATTERN)' -run=none $(BENCH_PACKAGES) | tee "$$scalar"; \
	arch=$$(go env GOARCH); \
	case "$$arch" in \
	arm64) \
		echo "Running arm64 accelerated benchmarks..."; \
		go test -tags accelerate -bench='$(BENCH_PATTERN)' -run=none $(BENCH_PACKAGES) | tee "$$accel"; \
		;; \
	amd64) \
		echo "Running amd64 accelerated benchmarks with GOEXPERIMENT=arenas,simd..."; \
		GOEXPERIMENT=arenas,simd go test -tags accelerate -bench='$(BENCH_PATTERN)' -run=none $(BENCH_PACKAGES) | tee "$$accel"; \
		;; \
	*) \
		echo "bench-compare only supports arm64 and amd64 hosts"; \
		exit 1; \
		;; \
	esac; \
	echo "Scalar output: $$scalar"; \
	echo "Accelerated output: $$accel"; \
	if command -v benchstat >/dev/null 2>&1; then \
		echo "Running benchstat..."; \
		benchstat "$$scalar" "$$accel"; \
	else \
		echo "benchstat not installed; compare the output files manually."; \
	fi

clean:
	@echo "Cleaning build artifacts and coverage information..."
	@rm -rf $(BUILD_DIR) $(COVERAGE_FILE)
	@rm -rf tests/*.ann *.ann results.txt tests/results.txt $(BENCH_DIR)
