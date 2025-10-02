# Makefile

# Variables
APP_NAME := configuration-management-service
DOCKER_COMPOSE := docker compose
GO := go
MOCKGEN := mockgen

# Default target
.PHONY: all
all: build up

# Docker compose
.PHONY: build
build:
	$(DOCKER_COMPOSE) build

.PHONY: up
up:
	$(DOCKER_COMPOSE) up -d

.PHONY: down
down:
	$(DOCKER_COMPOSE) down

.PHONY: logs
logs:
	$(DOCKER_COMPOSE) logs -f

# Go
.PHONY: test
test:
	$(GO) test ./...

.PHONY: test-coverage
test-coverage:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

.PHONY: tidy
tidy:
	$(GO) mod tidy

# Mocks
# Example: generate repository/service mocks; add more lines as needed.
# Ensure mockgen is installed: go install go.uber.org/mock/mockgen@latest
.PHONY: mocks
mocks:
	$(MOCKGEN) -source=internal/remote_config/repository/repository.go -destination=internal/remote_config/repository/mocks/repository_mock.go -package=mocks
	$(MOCKGEN) -source=internal/remote_config/service/service.go -destination=internal/remote_config/service/mocks/service_mock.go -package=mocks

# Utilities
.PHONY: run
run:
	$(GO) run ./cmd

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: ci
ci: tidy fmt vet test