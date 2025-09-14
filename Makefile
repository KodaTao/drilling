# Makefile for Drilling Platform
.PHONY: help build build-backend build-frontend clean dev test lint docker-build docker-run deploy

# Variables
APP_NAME = drilling
BACKEND_DIR = .
FRONTEND_DIR = web
DOCKER_IMAGE = drilling-platform
VERSION ?= latest

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
build: build-frontend build-backend ## Build both frontend and backend

build-all: build-frontend
	@echo "Building backend..."
	@echo "Building arm64 darwin"
	GOOS=darwin GOARCH=arm64 go build -o bin/$(APP_NAME)-darwin_arm64 ./cmd
	@echo "Building amd64 darwin"
	GOOS=darwin GOARCH=amd64 go build -o bin/$(APP_NAME)-darwin_amd64 ./cmd
	@echo "Building amd64 windows"
	GOOS=windows GOARCH=amd64 go build -o bin/$(APP_NAME)-windows_amd64.exe ./cmd

build-backend: build-frontend ## Build Go backend (depends on frontend)
	@echo "Building backend..."
	cd $(BACKEND_DIR) && go mod tidy
	cd $(BACKEND_DIR) && go build -o bin/$(APP_NAME) ./cmd

build-frontend: ## Build React frontend
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && npm install
	cd $(FRONTEND_DIR) && npm run build

# Development targets
dev: ## Start development environment
	@echo "Starting development environment..."
	@$(MAKE) -j2 dev-backend dev-frontend

dev-backend: ## Start backend in development mode
	@echo "Starting backend development server..."
	cd $(BACKEND_DIR) && go run ./cmd

dev-frontend: ## Start frontend in development mode
	@echo "Starting frontend development server..."
	cd $(FRONTEND_DIR) && npm run dev

# Test targets
test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	@echo "Running backend tests..."
	cd $(BACKEND_DIR) && go test ./...

test-frontend: ## Run frontend tests
	@echo "Running frontend tests..."
	cd $(FRONTEND_DIR) && npm test

# Lint targets
lint: lint-backend lint-frontend ## Run all linters

lint-backend: ## Lint Go code
	@echo "Linting backend..."
	cd $(BACKEND_DIR) && go vet ./...
	cd $(BACKEND_DIR) && go fmt ./...

lint-frontend: ## Lint frontend code
	@echo "Linting frontend..."
	cd $(FRONTEND_DIR) && npm run build

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

docker-run: ## Run application in Docker
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 3000:3000 $(DOCKER_IMAGE):$(VERSION)

# Clean targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(FRONTEND_DIR)/node_modules

# Install dependencies
install: ## Install all dependencies
	@echo "Installing dependencies..."
	cd $(BACKEND_DIR) && go mod download
	cd $(FRONTEND_DIR) && npm install

# Production targets
deploy: build ## Deploy to production
	@echo "Deploying to production..."
	# Add your deployment commands here

# Database targets
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	cd $(BACKEND_DIR) && go run ./cmd/migrate

db-seed: ## Seed database with sample data
	@echo "Seeding database..."
	cd $(BACKEND_DIR) && go run ./cmd/seed