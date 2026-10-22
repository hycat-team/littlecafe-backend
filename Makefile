# Config and variables
export

SERVICES := identity order activity wallet ai notification reporting
WIRE_VERSION := v0.7.0
BUF_VERSION := v1.57.2
OAPI_CODEGEN_VERSION := v2.4.1
GO_BUILD_FLAGS := -trimpath -ldflags="-s -w"
BIN_DIR := bin
PROTO_FILES := $(wildcard contracts/proto/*.proto contracts/proto/*/*.proto contracts/proto/*/*/*.proto contracts/proto/*/*/*/*.proto)

DB_HOST ?= localhost
DB_USER ?= littlecafe_$(SVC)
DB_PASSWORD ?= local_$(SVC)_password
DB_NAME ?= littlecafe_$(SVC)
DB_PORT ?= 5432
DB_SSLMODE ?= disable
DB_TIMEZONE ?= Asia/Ho_Chi_Minh
MIGRATION_DSN := "host=$(DB_HOST) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) port=$(DB_PORT) sslmode=$(DB_SSLMODE) TimeZone=$(DB_TIMEZONE)"

SERVER_PORT ?= 8080
CONFIG_BASE ?= $(abspath configs/base.yaml)
CONFIG_OVERLAY ?= $(abspath configs/$(SVC).yaml)
TEST_FLAGS ?=

ifeq ($(OS),Windows_NT)
GO_BIN_EXT := .exe
PWSH ?= powershell
MKDIR_BIN_CMD := if not exist "$(subst /,\,$(BIN_DIR))" mkdir "$(subst /,\,$(BIN_DIR))"
CLEAN_BIN_CMD := if exist "$(subst /,\,$(BIN_DIR))" rmdir /S /Q "$(subst /,\,$(BIN_DIR))"
FMT_CHECK_CMD := powershell -Command '$$files = gofmt -l .; if ($$files) { Write-Error ($$files -join [Environment]::NewLine); exit 1 }'
TEST_CMD := go test $(TEST_FLAGS) ./...
RACE_TEST_CMD := go test -race $(TEST_FLAGS) ./...
RUN_BIN = ..\..\$(subst /,\,$(BIN_DIR))\$(SVC)$(GO_BIN_EXT)
else
GO_BIN_EXT :=
PWSH ?= pwsh
MKDIR_BIN_CMD := mkdir -p "$(BIN_DIR)"
CLEAN_BIN_CMD := rm -rf "$(BIN_DIR)"
FMT_CHECK_CMD := test -z "$$(gofmt -l .)"
TEST_CMD := go test $(TEST_FLAGS) ./...
RACE_TEST_CMD := go test -race $(TEST_FLAGS) ./...
RUN_BIN = ../../$(BIN_DIR)/$(SVC)$(GO_BIN_EXT)
endif

FMT_CMD := gofmt -w .

SVC := $(or $(SVC),$(svc))
name := $(or $(name),$(NAME))
SVC_REQ = $(if $(filter $(SVC),$(SERVICES)),,$(error This target requires svc=<$(SERVICES)> — e.g. make $@ svc=identity))
unexport SVC_REQ

.PHONY: help install-tools install-goose wire wire-all buf-gen buf-lint openapi-gen contract-lint generate generate-check fmt fmt-check tidy tidy-check vet test test-race lint check team-check team-list team-test build build-all run dev clean migrate-status migrate-up migrate-down migrate-create migrate-version docker-infra-up docker-dev-up docker-dev-stop docker-dev-down docker-prod-up docker-prod-stop docker-prod-down nginx-reload-local

help:
	@echo "========================================================================"
	@echo "LITTLECAFE BACKEND BUILD SYSTEM"
	@echo "========================================================================"
	@echo "  make dev        : generate code and build all services"
	@echo "  make run svc=   : run one service locally ($(SERVICES))"
	@echo "  make test       : run all workspace tests"
	@echo "  make check      : full quality gate (fmt, tidy, vet, race tests, build, specs)"
	@echo "  make team-check : validate the specs/<owner>-<NNN>-<slug> layout (strict)"
	@echo "  make team-list  : print the feature table from meta.json files"
	@echo "  make docker-dev-up : start local infrastructure + services"
	@echo "Per-service targets take svc=<$(SERVICES)>, e.g. make build svc=identity"
	@echo "========================================================================"

install-tools:
	@echo Installing pinned development tools...
	@go install github.com/google/wire/cmd/wire@$(WIRE_VERSION)
	@go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	@echo Development tools installed.

install-goose:
	@echo Installing Goose CLI tool...
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo Goose CLI installed.

wire:
	$(SVC_REQ)
	@echo Running Google Wire for service $(SVC)...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/$(SVC)/internal/di/...
	@echo Google Wire finished for $(SVC).

wire-all:
	@echo Running Google Wire for all services...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/identity/internal/di/...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/order/internal/di/...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/activity/internal/di/...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/wallet/internal/di/...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/ai/internal/di/...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/notification/internal/di/...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/reporting/internal/di/...
	@echo Google Wire finished for all services.

buf-gen:
	@echo Generating protobuf stubs into contracts/...
	@$(if $(strip $(PROTO_FILES)),buf generate,echo "No protobuf contracts yet; skipping generation.")
	@echo Protobuf generation finished.

buf-lint:
	@echo Linting protobuf contracts...
	@$(if $(strip $(PROTO_FILES)),buf lint,echo "No protobuf contracts yet; skipping lint.")
	@echo Protobuf lint finished.

openapi-gen:
	@echo Generating strict OpenAPI server types...
	@cd contracts/openapi && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION) --config oapi-codegen.yaml components.yaml
	@echo OpenAPI generation finished.

contract-lint: buf-lint
	@echo OpenAPI and AsyncAPI files are validated in CI with pinned linters.


generate: wire-all openapi-gen

fmt:
	@echo Formatting Go source files...
	@$(FMT_CMD)
	@echo Go source formatting finished.

fmt-check:
	@echo Checking Go source formatting...
	@$(FMT_CHECK_CMD)
	@echo Go source formatting is valid.

tidy:
	@echo "Tidying Go modules (workspace)..."
	@cd shared && go mod tidy
	@cd contracts && go mod tidy
	@cd services/identity && go mod tidy
	@cd services/order && go mod tidy
	@cd services/activity && go mod tidy
	@cd services/wallet && go mod tidy
	@cd services/ai && go mod tidy
	@cd services/notification && go mod tidy
	@cd services/reporting && go mod tidy
	@go work sync
	@echo Go module tidy finished.

tidy-check:
	@echo Checking whether go.mod and go.sum files are tidy...
	@cd shared && go mod tidy -diff
	@cd contracts && go mod tidy -diff
	@cd services/identity && go mod tidy -diff
	@cd services/order && go mod tidy -diff
	@cd services/activity && go mod tidy -diff
	@cd services/wallet && go mod tidy -diff
	@cd services/ai && go mod tidy -diff
	@cd services/notification && go mod tidy -diff
	@cd services/reporting && go mod tidy -diff
	@echo Go module files are tidy.

generate-check:
	@echo Checking generated Wire and OpenAPI sources...
	@$(MAKE) generate
	@git diff --exit-code -- contracts/openapi services/identity/internal/di services/order/internal/di services/activity/internal/di services/wallet/internal/di services/ai/internal/di services/notification/internal/di services/reporting/internal/di
	@echo Generated files are up to date.

vet:
	@echo "Running go vet (all workspace modules)..."
	@cd shared && go vet ./...
	@cd contracts && go vet ./...
	@cd services/identity && go vet ./...
	@cd services/order && go vet ./...
	@cd services/activity && go vet ./...
	@cd services/wallet && go vet ./...
	@cd services/ai && go vet ./...
	@cd services/notification && go vet ./...
	@cd services/reporting && go vet ./...
	@echo Go vet finished.

test:
	@echo "Running Go tests (all workspace modules)..."
	@cd shared && $(TEST_CMD)
	@cd contracts && $(TEST_CMD)
	@cd services/identity && $(TEST_CMD)
	@cd services/order && $(TEST_CMD)
	@cd services/activity && $(TEST_CMD)
	@cd services/wallet && $(TEST_CMD)
	@cd services/ai && $(TEST_CMD)
	@cd services/notification && $(TEST_CMD)
	@cd services/reporting && $(TEST_CMD)
	@echo Go tests finished.

test-race:
	@echo "Running Go tests with race detector (all workspace modules)..."
	@cd shared && $(RACE_TEST_CMD)
	@cd contracts && $(RACE_TEST_CMD)
	@cd services/identity && $(RACE_TEST_CMD)
	@cd services/order && $(RACE_TEST_CMD)
	@cd services/activity && $(RACE_TEST_CMD)
	@cd services/wallet && $(RACE_TEST_CMD)
	@cd services/ai && $(RACE_TEST_CMD)
	@cd services/notification && $(RACE_TEST_CMD)
	@cd services/reporting && $(RACE_TEST_CMD)
	@echo Go race tests finished.

team-check:
	@echo Checking Spec Kit feature layout...
	@$(PWSH) -NoProfile -File scripts/speckit/team-feature.ps1 check -Strict

team-list:
	@$(PWSH) -NoProfile -File scripts/speckit/team-feature.ps1 list

team-test:
	@$(PWSH) -NoProfile -File scripts/speckit/tests/team-feature.tests.ps1

lint: fmt-check vet

check: fmt-check tidy-check generate-check vet test-race contract-lint build-all team-check

build:
	$(SVC_REQ)
	@echo Building service $(SVC)...
	@$(MKDIR_BIN_CMD)
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/$(SVC)$(GO_BIN_EXT)" ./services/$(SVC)/cmd/$(SVC)
	@echo Build succeeded. Binary: $(BIN_DIR)/$(SVC)$(GO_BIN_EXT)

build-all:
	@echo Building service binaries...
	@$(MKDIR_BIN_CMD)
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/identity$(GO_BIN_EXT)" ./services/identity/cmd/identity
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/order$(GO_BIN_EXT)" ./services/order/cmd/order
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/activity$(GO_BIN_EXT)" ./services/activity/cmd/activity
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/wallet$(GO_BIN_EXT)" ./services/wallet/cmd/wallet
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/ai$(GO_BIN_EXT)" ./services/ai/cmd/ai
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/notification$(GO_BIN_EXT)" ./services/notification/cmd/notification
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/reporting$(GO_BIN_EXT)" ./services/reporting/cmd/reporting
	@echo Build succeeded. Binaries in $(BIN_DIR)/

run: build
	$(SVC_REQ)
	@echo Starting $(SVC) service...
	@cd services/$(SVC) && $(RUN_BIN)

dev: generate build-all

clean:
	@echo Cleaning generated binaries...
	@$(CLEAN_BIN_CMD)
	@echo Clean finished.

migrate-status:
	$(SVC_REQ)
	@goose -dir ./services/$(SVC)/database/migrations postgres $(MIGRATION_DSN) status

migrate-up:
	$(SVC_REQ)
	@goose -dir ./services/$(SVC)/database/migrations postgres $(MIGRATION_DSN) up

migrate-down:
	$(SVC_REQ)
	@goose -dir ./services/$(SVC)/database/migrations postgres $(MIGRATION_DSN) down

migrate-create:
	$(SVC_REQ)
ifndef name
	$(error migrate-create requires name=xxx — e.g. make migrate-create svc=identity name=add_feature)
endif
	@goose -dir ./services/$(SVC)/database/migrations create $(name) sql

migrate-version:
	$(SVC_REQ)
	@goose -dir ./services/$(SVC)/database/migrations postgres $(MIGRATION_DSN) version

docker-dev-up:
	@docker compose -f deployments/docker-compose.yml --env-file .env --profile full up -d --build

docker-infra-up:
	@docker compose -f deployments/docker-compose.yml --env-file .env --profile infra up -d

docker-dev-stop:
	@docker compose -f deployments/docker-compose.yml --env-file .env stop

docker-dev-down:
	@docker compose -f deployments/docker-compose.yml --env-file .env down -v

nginx-reload-local:
	@docker exec nginx-littlecafe-backend-dev nginx -t
	@docker exec nginx-littlecafe-backend-dev nginx -s reload

docker-prod-up:
	@docker compose -f deployments/docker-compose.prod.yml --env-file .env.production --env-file .env.deploy up -d

docker-prod-stop:
	@docker compose -f deployments/docker-compose.prod.yml --env-file .env.production --env-file .env.deploy stop

docker-prod-down:
	@docker compose -f deployments/docker-compose.prod.yml --env-file .env.production --env-file .env.deploy down -v

