# # Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./
BINARY_NAME := aegisbox
GATEWAY_BINARY_NAME=smpp-gateway
MANAGER_API_BINARY_NAME=manager-api
GATEWAY_PKG_PATH=./cmd/smpp-gateway
MANAGER_API_PKG_PATH=./cmd/manager-api
OUTPUT_DIR=./tmp/bin


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
audit: lint vulncheck
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
	$(GOBUILD) -o=$(OUTPUT_DIR)/$(GATEWAY_BINARY_NAME) $(GATEWAY_PKG_PATH)

## build/manager-api: build the manager-api application
.PHONY: build/manager-api
build/manager-api:
	@echo "Building Manager API..."
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
# OPERATIONS
# ==================================================================================== #

## push: push changes to the remote Git repository
.PHONY: push
push: tidy audit no-dirty
	git push

## production/deploy: build linux binary and suggest deployment steps
.PHONY: production/deploy
production/deploy: confirm tidy audit no-dirty
	@echo "Building Linux AMD64 binary for Gateway..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags='-s -w' -o=$(OUTPUT_DIR)/linux_amd64/$(GATEWAY_BINARY_NAME) $(GATEWAY_PKG_PATH)
	@echo "Building Linux AMD64 binary for Manager API..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags='-s -w' -o=$(OUTPUT_DIR)/linux_amd64/$(MANAGER_API_BINARY_NAME) $(MANAGER_API_PKG_PATH)
	@echo "Binaries created in $(OUTPUT_DIR)/linux_amd64/"
	@echo "TODO: Add deployment steps here (e.g., scp, docker build/push, systemctl restart)"

