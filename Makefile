# Config and variables
export

SERVICES := commerce experience ai platform
WIRE_VERSION := v0.7.0
GOOSE_VERSION := v3.28.0
BUF_VERSION := v1.57.2
OAPI_CODEGEN_VERSION := v2.4.1
REDOCLY_VERSION := 1.34.5
ASYNCAPI_VERSION := 6.2.1
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
COVER_DIR ?= coverage
COVER_PROFILE := $(COVER_DIR)/coverage.out
COVER_BUSINESS_PROFILE := $(COVER_DIR)/coverage.business.out
COVER_HTML := $(COVER_DIR)/coverage.html
COVER_FLAGS ?= -covermode=atomic
COVER_MIN ?= 80
COVER_EXCLUDE ?= cmd/,internal/di/,internal/bootstrap/,database/migrations/,shared/infrastructure/,contracts/openapi/,shared/provider.go,_gen.go,.pb.go
COVER_PKGS := ./shared/... ./contracts/... ./services/commerce/... ./services/experience/... ./services/ai/... ./services/platform/...
COVER_FILTER_CMD := go run scripts/coverage/filter.go -in "$(COVER_PROFILE)" -out "$(COVER_BUSINESS_PROFILE)" -exclude "$(COVER_EXCLUDE)"

# SOPS runs from a local binary on developer machines; the VPS and CI set
# SOPS_MODE=docker to use the pinned image instead.
SOPS_MODE ?= binary

ifeq ($(OS),Windows_NT)
GO_BIN_EXT := .exe
PWSH ?= powershell
MKDIR_BIN_CMD := $(PWSH) -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$(BIN_DIR)' | Out-Null"
CLEAN_BIN_CMD := $(PWSH) -NoProfile -Command "Remove-Item -Recurse -Force -Path '$(BIN_DIR)' -ErrorAction SilentlyContinue"
MKDIR_COVER_CMD := $(PWSH) -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$(COVER_DIR)' | Out-Null"
CLEAN_COVER_CMD := $(PWSH) -NoProfile -Command "Remove-Item -Recurse -Force -Path '$(COVER_DIR)' -ErrorAction SilentlyContinue"
FMT_CHECK_CMD := powershell -Command '$$files = gofmt -l .; if ($$files) { Write-Error ($$files -join [Environment]::NewLine); exit 1 }'
TEST_CMD := go test $(TEST_FLAGS) ./...
RACE_TEST_CMD := go test -race $(TEST_FLAGS) ./...
RUN_BIN = ..\..\$(subst /,\,$(BIN_DIR))\$(SVC)$(GO_BIN_EXT)
else
GO_BIN_EXT :=
PWSH ?= pwsh
MKDIR_BIN_CMD := mkdir -p "$(BIN_DIR)"
CLEAN_BIN_CMD := rm -rf "$(BIN_DIR)"
MKDIR_COVER_CMD := mkdir -p "$(COVER_DIR)"
CLEAN_COVER_CMD := rm -rf "$(COVER_DIR)"
FMT_CHECK_CMD := test -z "$$(gofmt -l .)"
TEST_CMD := go test $(TEST_FLAGS) ./...
RACE_TEST_CMD := go test -race $(TEST_FLAGS) ./...
RUN_BIN = ../../$(BIN_DIR)/$(SVC)$(GO_BIN_EXT)
endif

FMT_CMD := gofmt -w .

SVC := $(or $(SVC),$(svc))
name := $(or $(name),$(NAME))
SVC_REQ = $(if $(filter $(SVC),$(SERVICES)),,$(error This target requires svc=<$(SERVICES)> — e.g. make $@ svc=commerce))
unexport SVC_REQ

.PHONY: help install-tools install-goose wire wire-all buf-gen buf-lint openapi-gen openapi-lint asyncapi-lint openapi-check error-codes-check contract-lint generate generate-check fmt fmt-check tidy tidy-check vet test test-race test-cover test-cover-html coverage-check lint check team-check team-list team-test build build-all run dev clean migrate-status migrate-up migrate-down migrate-create migrate-version docker-infra-up docker-dev-up docker-dev-stop docker-dev-down docker-prod-up docker-prod-stop docker-prod-down nginx-reload-local env-local env-local-check env-edit env-updatekeys env-check env-verify env-sops-test

help:
	@echo "========================================================================"
	@echo "LITTLECAFE BACKEND BUILD SYSTEM"
	@echo "========================================================================"
	@echo "  make dev        : generate code and build all services"
	@echo "  make run svc=   : run one service locally ($(SERVICES))"
	@echo "  make test       : run all workspace tests"	@echo "  make test-cover : run workspace statement coverage -> $(COVER_PROFILE)"
	@echo "  make coverage-check : fail when business coverage is below $(COVER_MIN)%"
	@echo "  make test-cover-html : generate HTML coverage report -> $(COVER_HTML)"
	@echo "  make check      : full quality gate (fmt, tidy, vet, race tests, build, specs)"
	@echo "  make team-check : validate the specs/<owner>-<NNN>-<slug> layout (strict)"
	@echo "  make team-list  : print the feature table from meta.json files"
	@echo "  make docker-dev-up : start local infrastructure + services"
	@echo "Per-service targets take svc=<$(SERVICES)>, e.g. make build svc=commerce"
	@echo "========================================================================"

install-tools:
	@echo Installing pinned development tools...
	@go install github.com/google/wire/cmd/wire@$(WIRE_VERSION)
	@go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	@echo Development tools installed.

install-goose:
	@echo Installing Goose CLI tool...
	@go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)
	@echo Goose CLI installed.

wire:
	$(SVC_REQ)
	@echo Running Google Wire for service $(SVC)...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/$(SVC)/internal/di/...
	@echo Google Wire finished for $(SVC).

wire-all:
	@echo Running Google Wire for all services...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/commerce/internal/di/...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/experience/internal/di/...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/ai/internal/di/...
	@go run github.com/google/wire/cmd/wire@$(WIRE_VERSION) ./services/platform/internal/di/...
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

openapi-lint:
	@echo Linting OpenAPI contracts...
	@npx --yes @redocly/cli@$(REDOCLY_VERSION) lint contracts/openapi/components.yaml

asyncapi-lint:
	@echo Validating AsyncAPI contracts...
	@npx --yes @asyncapi/cli@$(ASYNCAPI_VERSION) validate contracts/asyncapi/asyncapi.yaml

openapi-check: openapi-gen
	@git diff --exit-code -- contracts/openapi
	@echo Generated OpenAPI sources are up to date.

error-codes-check:
	@sh deployments/scripts/check-error-codes.sh

contract-lint: buf-lint openapi-lint asyncapi-lint error-codes-check
	@echo Contract lint finished.


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
	@go work sync
	@cd shared && go mod tidy
	@cd contracts && go mod tidy
	@cd services/commerce && go mod tidy
	@cd services/experience && go mod tidy
	@cd services/ai && go mod tidy
	@cd services/platform && go mod tidy
	@echo Go module tidy finished.

tidy-check:
	@echo Checking whether go.mod and go.sum files are tidy...
	@cd shared && go mod tidy -diff
	@cd contracts && go mod tidy -diff
	@cd services/commerce && go mod tidy -diff
	@cd services/experience && go mod tidy -diff
	@cd services/ai && go mod tidy -diff
	@cd services/platform && go mod tidy -diff
	@echo Go module files are tidy.

generate-check:
	@echo Checking generated Wire and OpenAPI sources...
	@$(MAKE) generate
	@git diff --exit-code -- contracts/openapi go.work.sum services/commerce/internal/di services/experience/internal/di services/ai/internal/di services/platform/internal/di || \
		{ echo "Generated files or go.work.sum are stale; run 'make generate' and commit the result."; exit 1; }
	@echo Generated files are up to date.

vet:
	@echo "Running go vet (all workspace modules)..."
	@cd shared && go vet ./...
	@cd contracts && go vet ./...
	@cd services/commerce && go vet ./...
	@cd services/experience && go vet ./...
	@cd services/ai && go vet ./...
	@cd services/platform && go vet ./...
	@echo Go vet finished.

test:
	@echo "Running Go tests (all workspace modules)..."
	@cd shared && $(TEST_CMD)
	@cd contracts && $(TEST_CMD)
	@cd services/commerce && $(TEST_CMD)
	@cd services/experience && $(TEST_CMD)
	@cd services/ai && $(TEST_CMD)
	@cd services/platform && $(TEST_CMD)
	@echo Go tests finished.

test-race:
	@echo "Running Go tests with race detector (all workspace modules)..."
	@cd shared && $(RACE_TEST_CMD)
	@cd contracts && $(RACE_TEST_CMD)
	@cd services/commerce && $(RACE_TEST_CMD)
	@cd services/experience && $(RACE_TEST_CMD)
	@cd services/ai && $(RACE_TEST_CMD)
	@cd services/platform && $(RACE_TEST_CMD)
	@echo Go race tests finished.

test-cover:
	@echo "Running Go statement coverage (all workspace modules)..."
	@$(MKDIR_COVER_CMD)
	@go test $(TEST_FLAGS) $(COVER_FLAGS) -coverprofile="$(COVER_PROFILE)" $(COVER_PKGS)
	@echo "Coverage profile: $(COVER_PROFILE)"
	@$(COVER_FILTER_CMD)

coverage-check: test-cover
	@go run scripts/coverage/filter.go -in "$(COVER_PROFILE)" -out "$(COVER_BUSINESS_PROFILE)" -exclude "$(COVER_EXCLUDE)" -min $(COVER_MIN)

test-cover-html: test-cover
	@go tool cover -html="$(COVER_BUSINESS_PROFILE)" -o "$(COVER_HTML)"
	@echo "Coverage report: $(COVER_HTML)"

team-check:
	@echo Checking Spec Kit feature layout...
	@$(PWSH) -NoProfile -File scripts/speckit/team-feature.ps1 check -Strict

team-list:
	@$(PWSH) -NoProfile -File scripts/speckit/team-feature.ps1 list

team-test:
	@$(PWSH) -NoProfile -File scripts/speckit/tests/team-feature.tests.ps1

env-local:
	@if [ -f .env ]; then \
		echo ".env already exists; leaving it untouched."; \
	else \
		cp .env.example .env; \
		echo "Created .env from .env.example. Fill in local paths and provider sandbox keys."; \
	fi

env-local-check:
	@sh scripts/env/check-local-env.sh

env-edit:
	@if [ -z "$(ENV)" ]; then echo "usage: make env-edit ENV=staging|production"; exit 64; fi
	@SOPS_MODE=$(SOPS_MODE) sh deployments/scripts/sops-run.sh edit deployments/env/$(ENV).env.enc

env-updatekeys:
	@for env in production staging; do \
		echo "Syncing recipients for $${env}..."; \
		SOPS_MODE=$(SOPS_MODE) sh deployments/scripts/sops-run.sh updatekeys --yes --input-type dotenv deployments/env/$${env}.env.enc; \
	done

env-check:
	@sh deployments/scripts/check-env-keys.sh

env-verify:
	@for env in production staging; do \
		SOPS_MODE=$(SOPS_MODE) sh deployments/scripts/sops-run.sh decrypt \
			--input-type dotenv --output-type dotenv deployments/env/$${env}.env.enc >/dev/null \
			&& echo "$${env}: decrypt OK"; \
	done

env-sops-test:
	@SOPS_MODE=$(SOPS_MODE) sh deployments/scripts/tests/apply-env.tests.sh

lint: fmt-check vet

check: fmt-check tidy-check generate-check vet test-race contract-lint build-all team-check coverage-check env-check

build:
	$(SVC_REQ)
	@echo Building service $(SVC)...
	@$(MKDIR_BIN_CMD)
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/$(SVC)$(GO_BIN_EXT)" ./services/$(SVC)/cmd/$(SVC)
	@echo Build succeeded. Binary: $(BIN_DIR)/$(SVC)$(GO_BIN_EXT)

build-all:
	@echo Building service binaries...
	@$(MKDIR_BIN_CMD)
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/commerce$(GO_BIN_EXT)" ./services/commerce/cmd/commerce
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/experience$(GO_BIN_EXT)" ./services/experience/cmd/experience
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/ai$(GO_BIN_EXT)" ./services/ai/cmd/ai
	@go build $(GO_BUILD_FLAGS) -o "$(BIN_DIR)/platform$(GO_BIN_EXT)" ./services/platform/cmd/platform
	@echo Build succeeded. Binaries in $(BIN_DIR)/

run: build
	$(SVC_REQ)
	@echo Starting $(SVC) service...
	@cd services/$(SVC) && $(RUN_BIN)

dev: generate build-all

clean:
	@echo Cleaning generated binaries and coverage reports...
	@$(CLEAN_BIN_CMD)
	@$(CLEAN_COVER_CMD)
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
	$(error migrate-create requires name=xxx — e.g. make migrate-create svc=commerce name=add_feature)
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

