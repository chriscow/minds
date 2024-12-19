# marked .PHONY to avoid conflicts with files or directories of the same name.
.PHONY: up test test_v test_short test_race test_stress test_codecov test_reconnect build fmt wait run-examples

# Variables
EXAMPLES_DIR := ./_examples
EXAMPLES := $(wildcard $(EXAMPLES_DIR)/*)

# Default target: run all examples
.PHONY: run-examples
run-examples:
	@echo "Running all examples..."
	@for example in $(EXAMPLES); do \
		if [ -f $$example/main.go ]; then \
			echo "Running $$example"; \
			(cd $$example && go run .); \
		fi; \
	done
	
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  test             - Run all tests"
	@echo "  test_v           - Run all tests with verbose output"
	@echo "  test_short       - Run short tests"
	@echo "  test_race        - Run tests with race detection"
	@echo "  test_codecov     - Run tests with coverage profiling"
	@echo "  build            - Build the project"
	@echo "  fmt              - Format the code with gofmt and goimports"
	@echo "  run-examples     - Run all examples in $(EXAMPLES_DIR)"

# add -p to parallelize tests
test:
	go test -p=1 ./...

test_v:
	go test -v -p=1 ./...

test_short:
	go test ./... -short

test_race:
	go test ./... -short -race

# test_stress:
# 	go test -tags=stress -timeout=30m ./...

test_codecov:
	go test -coverprofile=coverage.out -covermode=atomic ./...

# test_reconnect:
# 	go test -tags=reconnect ./...

build:
	go build ./...

fmt:
	go fmt ./...
	goimports -l -w .

