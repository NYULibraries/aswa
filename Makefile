# Variables
SLACK_CHANNEL_ID ?=
SKIP_BUILD ?= 0
# Set to 1 to skip building the images
# Targets
.PHONY: all build container run run-app check_env init test conditional-build

# Default target: build the binary, run the app
all: build run

# Build binary for aswa
build:
	go build -o aswa ./cmd

# Build images for the aswa app and the test container
container:
	docker compose build

# Run the aswa app
run:
	./aswa

# Run the synthetic test on the app with the given name in a container
run-app: conditional-build
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then echo "Error: Please provide an app name. Usage: make run-app your_app_name_here"; exit 1; fi
	docker compose run --rm aswa /aswa $(filter-out $@,$(MAKECMDGOALS))

# Ignore any targets that are not files
%:
	@:

# Check if the required environment variables are set
check_env:
	@MISSING_VARS=""; \
	UNSET_COUNT=0; \
	if [ -z "$(SLACK_CHANNEL_ID)" ]; then MISSING_VARS="SLACK_CHANNEL_ID"; UNSET_COUNT=$$((UNSET_COUNT + 1)); fi; \
	if [ -z "$(CLUSTER_INFO)" ]; then MISSING_VARS="$$MISSING_VARS$${MISSING_VARS:+, }CLUSTER_INFO"; UNSET_COUNT=$$((UNSET_COUNT + 1)); fi; \
	if [ $$UNSET_COUNT -gt 0 ]; then \
	    echo "$$MISSING_VARS $$([ $$UNSET_COUNT -gt 1 ] && echo 'are' || echo 'is') not set. Please set $$([ $$UNSET_COUNT -gt 1 ] && echo 'them' || echo 'it') and try again."; \
	    exit 1; \
	fi

# Initialize the project by checking if the environment variables are set
init: check_env

# Run tests using the aswa_test container and ensure images are built before running
test: conditional-build
	docker compose run --rm test

# Conditionally run the build target based on the value of SKIP_BUILD
conditional-build:
	@if [ "$(SKIP_BUILD)" = "0" ]; then $(MAKE) container; fi
