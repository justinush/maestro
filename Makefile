GOLANGCI_LINT_VERSION ?= v2.12.1
GOFUMPT_VERSION ?= v0.9.2

MODULE := github.com/justinush/maestro

GO ?= go
GO_PACKAGES ?= $(shell $(GO) list ./...)

DIST_DIR ?= dist
BIN_NAME ?= maestro
MAESTRO ?= $(DIST_DIR)/$(BIN_NAME)

TEST_TIMEOUT ?= 10m

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

.PHONY: check
check: fmt-check lint vet test

.PHONY: ci
ci: verify check smoke build

.PHONY: ci-action
ci-action: verify fmt-check vet test smoke build

.PHONY: release-check
release-check: verify fmt-check vet test smoke

##@ Test

.PHONY: test
test:
	$(GO) test -count=1 -timeout=$(TEST_TIMEOUT) ./...

.PHONY: test-race
test-race:
	$(GO) test -count=1 -race -timeout=$(TEST_TIMEOUT) ./...

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
