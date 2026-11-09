.PHONY: setup build run test test-race test-checkptr vet lint bench compat clean

RUNTIME_PACKAGES := .,./angular,./dot,./euclidean,./manhattan,./hamming,./internal/backend,./internal/codec,./internal/core,./internal/engine,./internal/metric,./cmd/precision

setup:
	go mod download

build:
	go build -o bin/precision.out ./cmd/precision
	go build -o bin/shell.out ./cmd/shell

run:
	go run ./cmd/precision

test:
	go test -timeout 120s -coverpkg=$(RUNTIME_PACKAGES) -coverprofile=coverage.out ./...
	go run ./internal/cmd/covercheck -profile coverage.out

test-race:
	go test -race ./...

test-checkptr:
	go test -gcflags=all=-d=checkptr=2 ./...

vet:
	go vet ./...

lint: vet
	@test -z "$$(gofmt -l $$(git ls-files '*.go') | head -1)"

bench:
	go test ./bench -run '^$$' -bench . -benchmem

compat:
	go test -tags cppcompat ./compat -v

clean:
	go clean ./...
	rm -f bin/precision.out bin/shell.out coverage.out
