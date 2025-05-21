# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
AIRCMD=$(GOCMD) run github.com/cosmtrek/air@v1.43.0

# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./
BINARY_NAME := aegisbox
GATEWAY_BINARY_NAME=smpp-gateway
MANAGER_API_BINARY_NAME=manager-api
GATEWAY_PKG_PATH=./cmd/smpp-gateway
MANAGER_API_PKG_PATH=./cmd/manager-api
OUTPUT_DIR=./tmp/bin

# Docker commands
DOCKER_COMPOSE=docker compose
DEPLOY_DIR=./deploy

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	git diff --exit-code


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	$(GOFMT) ./...
	$(GOMOD) tidy -v

## audit: run quality control checks (vet, staticcheck, vulncheck)
.PHONY: audit
audit: 
	$(GOVET) ./...
	@echo "Audit checks passed."


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	$(GOTEST) -v -race -buildvcs ./...

## test/cover: run all tests and display coverage in browser
.PHONY: test/cover
test/cover:
	mkdir -p tmp
	$(GOTEST) -v -race -buildvcs -coverprofile=tmp/coverage.out ./...
	go tool cover -html=tmp/coverage.out

# ==================================================================================== #
# BUILDING
# ==================================================================================== #

.PHONY: build
build: build/gateway build/manager-api

## build/gateway: build the smpp-gateway application
.PHONY: build/gateway
build/gateway:
	@echo "Building Gateway..."
	mkdir -p $(OUTPUT_DIR)
	$(GOBUILD) -o=$(OUTPUT_DIR)/$(GATEWAY_BINARY_NAME) $(GATEWAY_PKG_PATH)

## build/manager-api: build the manager-api application
.PHONY: build/manager-api
build/manager-api:
	@echo "Building Manager API..."
	mkdir -p $(OUTPUT_DIR)
	$(GOBUILD) -o=$(OUTPUT_DIR)/$(MANAGER_API_BINARY_NAME) $(MANAGER_API_PKG_PATH)

## clean: remove temporary build files
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(OUTPUT_DIR)
	rm -f tmp/coverage.out

# ==================================================================================== #
# RUNNING (Requires DB environment variables to be set)
# ==================================================================================== #

## run: run the smpp-gateway application (default run target)
.PHONY: run
run: run/gateway

## run/gateway: build and run the smpp-gateway
.PHONY: run/gateway
run/gateway: build/gateway
	@echo "Running Gateway (ensure DB env vars are set)..."
	$(OUTPUT_DIR)/$(GATEWAY_BINARY_NAME)

## run/manager-api: build and run the manager-api
.PHONY: run/manager-api
run/manager-api: build/manager-api
	@echo "Running Manager API (ensure DB env vars are set)..."
	$(OUTPUT_DIR)/$(MANAGER_API_BINARY_NAME)

## run/live: run the smpp-gateway with live reloading (default live target)
.PHONY: run/live
run/live: run/live/gateway

## run/live/gateway: run the gateway with live reloading using air
.PHONY: run/live/gateway
run/live/gateway:
	@echo "Running Gateway with live reload (air)..."
	$(AIRCMD) \
		--build.cmd "make build/gateway" --build.bin "$(OUTPUT_DIR)/$(GATEWAY_BINARY_NAME)" --build.delay "100" \
		--build.exclude_dir "tmp" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, yaml, yml, env" \
		--misc.clean_on_exit "true"

## run/live/manager-api: run the manager-api with live reloading using air
.PHONY: run/live/manager-api
run/live/manager-api:
	@echo "Running Manager API with live reload (air)..."
	$(AIRCMD) \
		--build.cmd "make build/manager-api" --build.bin "$(OUTPUT_DIR)/$(MANAGER_API_BINARY_NAME)" --build.delay "100" \
		--build.exclude_dir "tmp" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, yaml, yml, env" \
		--misc.clean_on_exit "true"

# ==================================================================================== #
# PRODUCTION BUILDING
# ==================================================================================== #

## production/build: build production Linux AMD64 binaries
.PHONY: production/build
production/build: tidy
	@echo "Building Linux AMD64 binary for Gateway..."
	mkdir -p $(OUTPUT_DIR)/linux_amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags='-s -w' -o=$(OUTPUT_DIR)/linux_amd64/$(GATEWAY_BINARY_NAME) $(GATEWAY_PKG_PATH)
	@echo "Building Linux AMD64 binary for Manager API..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags='-s -w' -o=$(OUTPUT_DIR)/linux_amd64/$(MANAGER_API_BINARY_NAME) $(MANAGER_API_PKG_PATH)
	@echo "Binaries created in $(OUTPUT_DIR)/linux_amd64/"

# ==================================================================================== #
# DOCKER
# ==================================================================================== #

## docker/prepare: create necessary directories for deployment
.PHONY: docker/prepare
docker/prepare:
	@echo "Creating necessary directories for deployment..."
	mkdir -p $(DEPLOY_DIR)/prometheus
	mkdir -p $(DEPLOY_DIR)/loki
	mkdir -p $(DEPLOY_DIR)/promtail
	@echo "Directories created."

## docker/build: build Docker images
.PHONY: docker/build
docker/build: production/build
	@echo "Building Docker images..."
	cd $(DEPLOY_DIR) && $(DOCKER_COMPOSE) build

## docker/deploy: deploy application with Docker Compose
.PHONY: docker/deploy
docker/deploy: docker/build
	@echo "Starting services with Docker Compose..."
	cd $(DEPLOY_DIR) && $(DOCKER_COMPOSE) up -d

## docker/logs: view Docker container logs
.PHONY: docker/logs
docker/logs:
	@echo "Viewing Docker container logs..."
	cd $(DEPLOY_DIR) && $(DOCKER_COMPOSE) logs -f

## docker/stop: stop Docker services
.PHONY: docker/stop
docker/stop:
	@echo "Stopping Docker services..."
	cd $(DEPLOY_DIR) && $(DOCKER_COMPOSE) stop

## docker/down: completely remove Docker services and volumes
.PHONY: docker/down
docker/down:
	@echo "Removing Docker services and volumes..."
	cd $(DEPLOY_DIR) && $(DOCKER_COMPOSE) down -v

# ==================================================================================== #
# DEPLOYMENT
# ==================================================================================== #

## deploy: complete deployment process
.PHONY: deploy
deploy: confirm tidy production/build docker/prepare docker/build docker/deploy
	@echo "Deployment complete! Services should be running now."
	@echo "Access your services at:"
	@echo "  - Manager API: http://localhost:8001"
	@echo "  - SMPP Gateway: localhost:6886"
	@echo "  - Grafana: http://localhost:3000 (admin/secure_password by default)"
	@echo "  - Prometheus: http://localhost:9090"

## db/migrate: run database migrations manually
.PHONY: db/migrate
db/migrate:
	@echo "Running database migrations..."
	cd $(DEPLOY_DIR) && $(DOCKER_COMPOSE) run --rm migration
