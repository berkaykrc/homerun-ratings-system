# Homerun Ratings System Makefile

.PHONY: help build test clean start stop logs restart docker-build
.PHONY: rating-build rating-test rating-run 
.PHONY: notification-build notification-test notification-run
.PHONY: db-start db-stop db-migrate testdata

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build both services
	@echo "Creating bin directory..."
	@mkdir -p bin
	@echo "Building Rating Service..."
	@cd rating-service && go build -o ../bin/rating-service ./cmd/server
	@echo "Building Notification Service..."
	@cd notification-service && go build -o ../bin/notification-service ./cmd/server
	@echo "Build complete! Binaries are in ./bin/"

test: ## Run tests for both services
	@echo "Testing Rating Service..."
	@cd rating-service && go test -v ./... 
	@echo "Testing Notification Service..."
	@cd notification-service && go test -v ./...

clean: ## Clean build artifacts
	@rm -rf bin/
	@cd rating-service && go clean
	@cd notification-service && go clean

start: ## Start services with Docker Compose
	@docker compose up -d

stop: ## Stop services
	@docker compose down

logs: ## Show logs from all services
	@docker compose logs -f

restart: stop start ## Restart all services

docker-build: ## Build Docker images for both services
	@echo "Building Docker images..."
	@docker compose build
	@echo "Docker images built successfully!"

# Rating Service specific targets
rating-build: ## Build rating service only
	@cd rating-service && make build

rating-test: ## Test rating service only
	@cd rating-service && make test

rating-run: ## Run rating service locally
	@cd rating-service && make run

# Notification Service specific targets  
notification-build: ## Build notification service only
	@cd notification-service && go build ./cmd/server

notification-test: ## Test notification service only
	@cd notification-service && go test ./...

notification-run: ## Run notification service locally
	@cd notification-service && go run ./cmd/server

# Database operations
db-start: ## Start PostgreSQL database
	@docker run --name homerun-postgres -d \
		-e POSTGRES_DB=homerun_ratings \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-p 5432:5432 \
		postgres:alpine

db-stop: ## Stop PostgreSQL database
	@docker stop homerun-postgres || true
	@docker rm homerun-postgres || true

db-migrate: ## Run database migrations
	@cd rating-service && make migrate

testdata: ## Load test data
	@cd rating-service && make testdata
