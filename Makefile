APP_NAME := "BazaCars"
include .env

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ### this help information
	@awk 'BEGIN {FS = ":.*##"; printf "\nEngine Makefile help:\n  make \033[36m<target>\033[0m\n"} /^[.a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help

.DEFAULT_GOAL := help

build: ### build app
	@echo "Building ${APP_NAME}..."
	@GOOS=linux GOARCH=amd64 go build -o app ./cmd/main.go
	@echo "Stopping services (if running...)"
	docker compose down
	@echo "Starting ${APP_NAME} services..."
	@docker compose up --build -d
	@echo "Services built and started!"

test: ### run tests
	@echo "Running ${APP_NAME} tests..."
	go test --cover -timeout 30s ./...
	@echo "Done!"

up: ### up services
	@echo "Starting ${APP_NAME} services..."
	docker-compose up -d
	@echo "Services started!"
.PHONY:up

up_build: build ### build and up services
	@echo "Stopping services (if running...)"
	docker-compose down
	@echo "Starting ${APP_NAME} services..."
	@docker-compose up --build -d
	@echo "Services built and started!"
	@rm app
.PHONY:up_build

down: ### down services
	@echo "Stopping ${APP_NAME} services..."
	docker-compose down
	@echo "Done!"
.PHONY:down

generate: ### run go generate (for mocks)
	@echo ">  Building mocks for ${APP_NAME}...\n"
	go generate -v ./internal/...
.PHONY: generate
