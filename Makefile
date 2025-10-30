# Makefile for ASWA project

# Variables
CLUSTER_INFO ?=
DOCKER_COMPOSE_CMD ?= docker compose run --rm
PROM_AGGREGATION_GATEWAY_URL ?=
SLACK_CHANNEL_ID ?=
SKIP_BUILD ?= 0
# Set to 1 to skip building the images

# Targets
.PHONY: all build check-env clean conditional-build container container-run go-run init lint lint-install run run-app test

# Default target: build the images, run tests and run the app in a container
all: container test container-run

# Build binary for aswa
build:
	@echo "Building aswa binary..."
	go build

# Check if the required environment variables are set
check-env:
	@MISSING_VARS=""; \
	UNSET_COUNT=0; \
	if [ -z "$(SLACK_CHANNEL_ID)" ]; then MISSING_VARS="SLACK_CHANNEL_ID"; UNSET_COUNT=$$((UNSET_COUNT + 1)); fi; \
	if [ -z "$(CLUSTER_INFO)" ]; then MISSING_VARS="$$MISSING_VARS$${MISSING_VARS:+, }CLUSTER_INFO"; UNSET_COUNT=$$((UNSET_COUNT + 1)); fi; \
	if [ -z "$(PROM_AGGREGATION_GATEWAY_URL)" ]; then MISSING_VARS="$$MISSING_VARS$${MISSING_VARS:+, }PROM_AGGREGATION_GATEWAY_URL"; UNSET_COUNT=$$((UNSET_COUNT + 1)); fi; \
	if [ $$UNSET_COUNT -gt 0 ]; then \
	    echo "$$MISSING_VARS $$([ $$UNSET_COUNT -gt 1 ] && echo 'are' || echo 'is') not set. Please set $$([ $$UNSET_COUNT -gt 1 ] && echo 'them' || echo 'it') and try again."; \
	    exit 1; \
	fi

# Clean up the project
clean:
	@echo "Cleaning up..."
	rm -f aswa
	docker compose down --rmi all --volumes --remove-orphans
	@echo "Remember to manually unset YAML_PATH if it was exported in your shell."

# Conditionally run the build target based on the value of SKIP_BUILD
conditional-build:
	@if [ "$(SKIP_BUILD)" = "0" ]; then $(MAKE) container; fi

# Build images for the aswa app and the test container
container:
	@echo "Building images..."
	docker compose build

# Run the aswa app in a Docker container
container-run: conditional-build
	@echo "Running aswa in a Docker container..."
	$(DOCKER_COMPOSE_CMD) aswa

# Run the aswa app with 'go run'
go-run:
	@echo "Running aswa with 'go run'..."
	go run .

# Initialize the project by checking if the environment variables are set
init: check-env

lint: ## golangci-lint (must be preinstalled)
	@command -v golangci-lint >/dev/null || { \
	echo "golangci-lint not found. Install with:" ; \
	echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" ; \
	exit 1 ; \
	}
	golangci-lint run ./...

lint-install: ## Install golangci-lint (ensure GOBIN is in PATH)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run aswa using the built binary
run: build
	@echo "Running aswa..."
	./aswa

# Run the synthetic test on the app with the given name in a container
run-app: conditional-build
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then echo "Error: Please provide an app name. Usage: make run-app your_app_name_here"; exit 1; fi
	$(DOCKER_COMPOSE_CMD) aswa /aswa $(filter-out $@,$(MAKECMDGOALS))

# Ignore any targets that are not files
%:
	@:

# Run tests using the aswa_test container and ensure images are built before running
test: conditional-build
	@echo "Running tests..."
	$(DOCKER_COMPOSE_CMD) test



