.PHONY: help build clean frontend backend test docker docker-build docker-push run dev install

# Variables
BINARY_NAME=guardian-log
VERSION?=latest
DOCKER_IMAGE?=guardian-log
DOCKER_REGISTRY?=
PLATFORMS?=linux/amd64,linux/arm64

# Colors for output
GREEN=\033[0;32m
NC=\033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)Guardian-Log Build System$(NC)"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	cd web && npm install

frontend: ## Build frontend only
	@echo "$(GREEN)Building frontend...$(NC)"
	cd web && npm run build
	@echo "$(GREEN)Frontend build complete!$(NC)"

backend-prep: ## Prepare backend for building (copy frontend dist)
	@echo "$(GREEN)Preparing backend build...$(NC)"
	rm -rf webfs/web
	mkdir -p webfs/web
	cp -r web/dist webfs/web/
	@echo "$(GREEN)Backend prep complete!$(NC)"

backend: frontend backend-prep ## Build backend only
	@echo "$(GREEN)Building backend...$(NC)"
	go build -ldflags="-s -w" -o ./bin/$(BINARY_NAME) ./cmd/guardian-log
	@echo "$(GREEN)Backend build complete!$(NC)"

build: install frontend backend-prep backend ## Build complete application (frontend + backend)
	@echo "$(GREEN)✅ Build complete! Binary: ./bin/$(BINARY_NAME)$(NC)"
	@echo "$(GREEN)Run with: ./bin/$(BINARY_NAME)$(NC)"

build-quick: frontend backend-prep ## Quick build (skip npm install)
	@echo "$(GREEN)Quick building backend...$(NC)"
	go build -ldflags="-s -w" -o ./bin/$(BINARY_NAME) ./cmd/guardian-log
	@echo "$(GREEN)✅ Quick build complete!$(NC)"

run: ## Build and run the application
	@echo "$(GREEN)Building and running...$(NC)"
	@$(MAKE) build
	./bin/$(BINARY_NAME)

dev: ## Run in development mode (separate frontend/backend)
	@echo "$(GREEN)Starting development mode...$(NC)"
	@echo "$(GREEN)Backend: http://localhost:8080$(NC)"
	@echo "$(GREEN)Frontend: http://localhost:5173$(NC)"
	@echo ""
	@echo "Run 'make dev-backend' in one terminal and 'make dev-frontend' in another"
	@echo "Or run 'make dev-both' to start both in parallel (requires tmux)"

dev-backend: ## Run backend with hot-reload (using air)
	@echo "$(GREEN)Starting backend with hot-reload...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(GREEN)air not found, falling back to go run$(NC)"; \
		go run ./cmd/guardian-log; \
	fi

dev-frontend: ## Run frontend dev server with hot-reload
	@echo "$(GREEN)Starting frontend dev server...$(NC)"
	cd web && npm run dev

dev-both: ## Run both frontend and backend in parallel (requires tmux)
	@echo "$(GREEN)Starting both frontend and backend in tmux...$(NC)"
	@if ! command -v tmux > /dev/null; then \
		echo "Error: tmux is required. Install with: sudo apt install tmux"; \
		exit 1; \
	fi
	@tmux new-session -d -s guardian-dev \
		"make dev-backend" \; \
		split-window -h "make dev-frontend" \; \
		attach
	@echo "$(GREEN)Tmux session 'guardian-dev' started$(NC)"
	@echo "$(GREEN)Use Ctrl+B then D to detach, 'tmux attach -t guardian-dev' to reattach$(NC)"

test: ## Run tests
	@echo "$(GREEN)Running Go tests...$(NC)"
	go test -v ./...

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf ./bin/$(BINARY_NAME)
	rm -rf web/dist
	rm -rf web/node_modules
	rm -rf webfs/web
	@echo "$(GREEN)Clean complete!$(NC)"

docker-build: ## Build Docker image for current platform
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(VERSION)$(NC)"

docker-build-multi: ## Build multi-architecture Docker images
	@echo "$(GREEN)Building multi-architecture Docker images...$(NC)"
	docker buildx build --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE):$(VERSION) \
		$(if $(DOCKER_REGISTRY),-t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)) \
		--push \
		.
	@echo "$(GREEN)Multi-arch build complete!$(NC)"

docker-build-local: ## Build multi-architecture images locally (no push)
	@echo "$(GREEN)Building multi-architecture Docker images locally...$(NC)"
	docker buildx build --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE):$(VERSION) \
		--load \
		.
	@echo "$(GREEN)Multi-arch local build complete!$(NC)"

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -p 8080:8080 \
		-v $(PWD)/data:/app/data \
		--env-file .env \
		$(DOCKER_IMAGE):$(VERSION)

docker-compose-up: ## Start with docker-compose
	@echo "$(GREEN)Starting with docker-compose...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Services started!$(NC)"
	@echo "$(GREEN)Dashboard: http://localhost:8080$(NC)"

docker-compose-down: ## Stop docker-compose services
	docker-compose down

docker-compose-logs: ## Show docker-compose logs
	docker-compose logs -f

docker-push: ## Push Docker image to registry
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "Error: DOCKER_REGISTRY not set. Use: make docker-push DOCKER_REGISTRY=your-registry"; \
		exit 1; \
	fi
	@echo "$(GREEN)Pushing to $(DOCKER_REGISTRY)...$(NC)"
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)
	@echo "$(GREEN)Push complete!$(NC)"

lint: ## Run linters
	@echo "$(GREEN)Running Go linters...$(NC)"
	go vet ./...
	@echo "$(GREEN)Running frontend linters...$(NC)"
	cd web && npm run lint || true

fmt: ## Format code
	@echo "$(GREEN)Formatting Go code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Formatting complete!$(NC)"

.DEFAULT_GOAL := help
