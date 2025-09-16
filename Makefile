# load environment variables from .env
include .env
export
# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go install github.com/onsi/ginkgo/v2/ginkgo@latest
	@$(GOPATH)/bin/ginkgo -r


# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go install github.com/onsi/ginkgo/v2/ginkgo@latest
	@$(GOPATH)/bin/ginkgo -r internal/database


# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Goose migration config
GOOSE=go run github.com/pressly/goose/v3/cmd/goose@latest
DB_URL=postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)&search_path=$(DB_SCHEMA)
MIGRATIONS_DIR=internal/migrations

# Run all migrations
migrate-up:
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

# Rollback last migration
migrate-down:
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

# Redo last migration
migrate-redo:
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" redo

# Show current migration status
migrate-status:
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

# Create a new migration file
migrate-new:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-new name=create_table"; \
		exit 1; \
	fi; \
	$(GOOSE) create $(name) sql -dir $(MIGRATIONS_DIR)

.PHONY: all build run test clean watch docker-run docker-down itest migrate-up migrate-down migrate-redo migrate-status migrate-new
