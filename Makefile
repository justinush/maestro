GOLANGCI_LINT_VERSION ?= v2.12.1
GOFUMPT_VERSION ?= v0.9.2

MODULE := github.com/justinush/maestro

GO ?= go
GO_PACKAGES ?= $(shell $(GO) list ./...)

DIST_DIR ?= dist
BIN_NAME ?= maestro
MAESTRO ?= $(DIST_DIR)/$(BIN_NAME)

TEST_TIMEOUT ?= 10m

MAESTRO_TEST_DATABASE_URL ?= postgres://maestro:maestro@localhost:5434/maestro_test?sslmode=disable
POSTGRES_TEST_CONTAINER ?= maestro-pg-test

GO_INSTALL_BIN := $(shell $(GO) env GOBIN)
ifeq ($(GO_INSTALL_BIN),)
GO_INSTALL_BIN := $(shell $(GO) env GOPATH)/bin
endif
GOFUMPT := $(GO_INSTALL_BIN)/gofumpt
GOLANGCI_LINT := $(GO_INSTALL_BIN)/golangci-lint

CLI_MAIN := ./cmd/maestro

VERSION ?= dev
LDFLAGS ?= -X $(MODULE)/cli.Version=$(VERSION)

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

##@ Dependencies

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: verify
verify:
	$(GO) mod verify

.PHONY: install-golangci-lint
install-golangci-lint:
	@test -x "$(GOLANGCI_LINT)" || $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: install-golangci-lint-force
install-golangci-lint-force:
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: install-gofumpt
install-gofumpt:
	@test -x "$(GOFUMPT)" || $(GO) install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)

.PHONY: install-tools
install-tools: install-golangci-lint install-gofumpt

##@ Quality

.PHONY: fmt
fmt: install-gofumpt
	"$(GOFUMPT)" -extra -w .

.PHONY: fmt-quick
fmt-quick: install-gofumpt
	"$(GOFUMPT)" -extra -w .

.PHONY: fmt-check
fmt-check: install-gofumpt
	@"$(GOFUMPT)" -extra -l . | tee /tmp/gofumpt.out; test ! -s /tmp/gofumpt.out

.PHONY: vet
vet:
	$(GO) vet $(GO_PACKAGES)

.PHONY: lint
lint: install-golangci-lint
	"$(GOLANGCI_LINT)" run

.PHONY: lint-fix
lint-fix: install-golangci-lint
	"$(GOLANGCI_LINT)" run --fix

.PHONY: doc-check
doc-check:
	$(GO) test ./pkg/...
	@$(GO) doc ./pkg/engine Instance | grep -q 'func (in \*Instance) RunUntilBlocked'
	@$(GO) doc ./pkg/engine Instance | grep -q 'func (in \*Instance) SubmitInput'
	@$(GO) doc ./pkg/engine Registry | grep -q 'func (r \*Registry) Register('
	@$(GO) doc ./pkg/engine Registry | grep -q 'func (r \*Registry) Lookup'
	@$(GO) doc ./pkg/maestro Runtime | grep -q 'type Runtime struct'
	@$(GO) doc ./pkg/validate Options | grep -q 'SchemaPath'
	@$(GO) doc ./pkg/definition WorkflowDefinition | grep -q 'InitialStepID'

.PHONY: check
check: fmt-check lint vet test

.PHONY: ci
ci: verify check smoke build

.PHONY: ci-action
ci-action: verify fmt-check vet test smoke build

.PHONY: release-check
release-check: verify fmt-check vet test smoke doc-check

##@ Test

.PHONY: test
test:
	$(GO) test -count=1 -timeout=$(TEST_TIMEOUT) ./...

.PHONY: test-race
test-race:
	$(GO) test -count=1 -race -timeout=$(TEST_TIMEOUT) ./...

##@ Postgres integration

.PHONY: postgres-docker-up
postgres-docker-up:
	@if docker ps --format '{{.Names}}' | grep -qx '$(POSTGRES_TEST_CONTAINER)'; then \
		echo "Postgres test container already running ($(POSTGRES_TEST_CONTAINER))"; \
	elif docker ps -a --format '{{.Names}}' | grep -qx '$(POSTGRES_TEST_CONTAINER)'; then \
		echo "Starting existing container $(POSTGRES_TEST_CONTAINER)..."; \
		docker start $(POSTGRES_TEST_CONTAINER); \
	else \
		echo "Creating Postgres test container $(POSTGRES_TEST_CONTAINER)..."; \
		docker run -d --name $(POSTGRES_TEST_CONTAINER) \
			-e POSTGRES_USER=maestro -e POSTGRES_PASSWORD=maestro -e POSTGRES_DB=maestro_test \
			-p 5434:5432 postgres:16; \
	fi

.PHONY: postgres-wait
postgres-wait:
	@echo "Waiting for Postgres ($(POSTGRES_TEST_CONTAINER))..."
	@for i in $$(seq 1 30); do \
		docker exec $(POSTGRES_TEST_CONTAINER) pg_isready -U maestro -d maestro_test -q 2>/dev/null && exit 0; \
		sleep 1; \
	done; \
	echo "Postgres did not become ready in 30s" >&2; exit 1

.PHONY: postgres-docker-down
postgres-docker-down:
	docker rm -f $(POSTGRES_TEST_CONTAINER) 2>/dev/null || true

.PHONY: postgres-docker-reset
postgres-docker-reset: postgres-docker-down postgres-docker-up postgres-wait

.PHONY: test-postgres
test-postgres: postgres-docker-up postgres-wait
	MAESTRO_TEST_DATABASE_URL="$(MAESTRO_TEST_DATABASE_URL)" \
		$(GO) test -count=1 -timeout=$(TEST_TIMEOUT) ./pkg/run/postgres/...

##@ Build

.PHONY: build
build:
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(DIST_DIR)/$(BIN_NAME) $(CLI_MAIN)

.PHONY: install
install:
	CGO_ENABLED=0 $(GO) install -trimpath -ldflags '$(LDFLAGS)' $(CLI_MAIN)

##@ Examples / smoke

.PHONY: validate-example
validate-example: build
	"$(MAESTRO)" validate -f examples/workflows/workflow-v0-minimal.yaml

.PHONY: simulate-example
simulate-example: build
	"$(MAESTRO)" simulate -s examples/scenarios/scenario-minimal.yaml

.PHONY: simulate-negative
simulate-negative: build
	"$(MAESTRO)" simulate -s examples/scenarios/scenario-invalid-missing-required.yaml
	"$(MAESTRO)" simulate -s examples/scenarios/scenario-invalid-additional-property.yaml
	"$(MAESTRO)" simulate -s examples/scenarios/scenario-cel-runtime-error.yaml
	"$(MAESTRO)" simulate -s examples/scenarios/scenario-cel-invalid.yaml
	"$(MAESTRO)" simulate -s examples/scenarios/scenario-ambiguous-unconditional.yaml

.PHONY: validate-portrait
validate-portrait: build
	"$(MAESTRO)" validate -f examples/kyc/sg/portrait/workflow.yaml

.PHONY: simulate-portrait
simulate-portrait: build
	"$(MAESTRO)" simulate -s examples/kyc/sg/portrait/scenario-happy.yaml
	"$(MAESTRO)" simulate -s examples/kyc/sg/portrait/scenario-partner-rejected.yaml
	"$(MAESTRO)" simulate -s examples/kyc/sg/portrait/scenario-poa-review.yaml

.PHONY: smoke
smoke: validate-example simulate-example simulate-negative validate-portrait simulate-portrait

##@ Cleanup

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)

.PHONY: clean-cache
clean-cache: clean
	$(GO) clean -cache -testcache 2>/dev/null || true
