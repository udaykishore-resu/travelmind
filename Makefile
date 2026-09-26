.PHONY: help dev-setup dev-up dev-down run-services test docker-build docker-push k8s-deploy k8s-rollback lint fmt

VERSION ?= dev
REGISTRY ?= gcr.io/travelmind
IMAGE_TAG ?= $(VERSION)
PROJECT_ID ?= travelmind-dev

help:
	@echo "TravelMind - Travel Orchestration Platform"
	@echo ""
	@echo "Development Commands:"
	@echo "  make dev-setup        - Set up local development environment"
	@echo "  make dev-up           - Start Docker Compose services"
	@echo "  make dev-down         - Stop Docker Compose services"
	@echo "  make run-services     - Run backend services locally"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  make test             - Run all tests"
	@echo "  make test-unit        - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make coverage         - Run tests with coverage"
	@echo "  make lint             - Run linters"
	@echo "  make fmt              - Format code"
	@echo ""
	@echo "Docker & Deployment:"
	@echo "  make docker-build     - Build Docker images"
	@echo "  make docker-push      - Push images to registry"
	@echo "  make k8s-deploy       - Deploy to Kubernetes"
	@echo "  make k8s-rollback     - Rollback Kubernetes deployment"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate-up    - Run database migrations"
	@echo "  make db-migrate-down  - Rollback database migrations"
	@echo ""

dev-setup:
	@echo "Setting up development environment..."
	cd backend && go mod download
	@echo "Installing golangci-lint..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "Development environment ready!"

dev-up:
	@echo "Starting Docker Compose services..."
	docker-compose up -d
	@echo "Services started. API Gateway: http://localhost:8080"
	@echo "Grafana: http://localhost:3000"
	@echo "Prometheus: http://localhost:9091"
	@echo "Jaeger: http://localhost:16686"

dev-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down

dev-logs:
	docker-compose logs -f api-gateway

run-services:
	@echo "Running backend services..."
	cd backend && go run ./cmd/api-gateway/main.go

test:
	@echo "Running tests..."
	cd backend && go test -v -cover ./...

test-unit:
	@echo "Running unit tests..."
	cd backend && go test -v -short ./...

test-integration:
	@echo "Running integration tests..."
	cd backend && go test -v -run Integration ./...

coverage:
	@echo "Running tests with coverage..."
	cd backend && go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

lint:
	@echo "Running linters..."
	cd backend && golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	cd backend && gofmt -w .
	cd backend && go mod tidy

docker-build:
	@echo "Building Docker images..."
	docker build -t $(REGISTRY)/travelmind-api-gateway:$(IMAGE_TAG) ./backend
	@echo "Build complete: $(REGISTRY)/travelmind-api-gateway:$(IMAGE_TAG)"

docker-push:
	@echo "Pushing Docker images to registry..."
	docker push $(REGISTRY)/travelmind-api-gateway:$(IMAGE_TAG)
	@echo "Push complete!"

docker-run:
	docker run -p 8080:8080 -p 9090:9090 \
		-e DATABASE_URL="postgres://travelmind:password@host.docker.internal:5432/travelmind" \
		-e REDIS_URL="redis://host.docker.internal:6379" \
		$(REGISTRY)/travelmind-api-gateway:$(IMAGE_TAG)

k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f k8s/api-gateway-deployment.yaml
	kubectl rollout status deployment/api-gateway -n travelmind
	@echo "Deployment complete!"

k8s-logs:
	kubectl logs -f deployment/api-gateway -n travelmind

k8s-rollback:
	@echo "Rolling back deployment..."
	kubectl rollout undo deployment/api-gateway -n travelmind

k8s-scale:
	@echo "Scaling deployment to $(REPLICAS) replicas..."
	kubectl scale deployment api-gateway --replicas=$(REPLICAS) -n travelmind

db-migrate-up:
	@echo "Running database migrations..."
	cd backend && go run ./cmd/migrations/main.go up

db-migrate-down:
	@echo "Rolling back database migrations..."
	cd backend && go run ./cmd/migrations/main.go down

gen-mocks:
	@echo "Generating mocks..."
	cd backend && go generate ./...

swagger-docs:
	@echo "Generating Swagger documentation..."
	cd backend && swag init -g cmd/api-gateway/main.go

clean:
	@echo "Cleaning up..."
	rm -rf backend/coverage.out
	docker-compose down -v
	go clean ./...

.PHONY: help dev-setup dev-up dev-down dev-logs run-services test test-unit test-integration coverage lint fmt docker-build docker-push docker-run k8s-deploy k8s-logs k8s-rollback k8s-scale db-migrate-up db-migrate-down gen-mocks swagger-docs clean
