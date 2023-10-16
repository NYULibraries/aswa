# Makefile

# Variables
SLACK_CHANNEL_ID ?=
SKIP_BUILD ?= 0

# Targets
.PHONY: all build up down clean test run run-app check_env init conditional-build

# Default target: build the Docker images
all: build

# Build Docker images for all services
build:
	docker-compose build

# Start containers in detached mode
up:
	docker-compose up -d

# Stop and remove containers, networks, and volumes
down:
	docker-compose down

# Remove all unused containers, networks, images, and volumes
clean:
	docker system prune -a --volumes

# Run tests using the aswa_test container and ensure images are built before running
test: conditional-build
	docker-compose run --rm test

# Run the aswa container, remove it after completion, and ensure images are built before running
run: conditional-build
	docker-compose run --rm aswa

# Run the synthetic test on the app with the given name
run-app: conditional-build
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then echo "Error: Please provide an app name. Usage: make run-app your_app_name_here"; exit 1; fi
	docker-compose run --rm aswa /aswa $(filter-out $@,$(MAKECMDGOALS))

# Ignore any targets that are not files
%:
	@:

# Check if the required environment variables are set
check_env:
	@if [ -z "$(SLACK_CHANNEL_ID)" ]; then echo "SLACK_CHANNEL_ID is not set. Please set it and try again."; exit 1; fi

# Initialize the project by checking if the environment variables are set
init: check_env

# Conditionally run the build target based on the value of SKIP_BUILD
conditional-build:
	@if [ "$(SKIP_BUILD)" = "0" ]; then $(MAKE) build; fi
