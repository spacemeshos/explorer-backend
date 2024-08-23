VERSION ?= $(shell git describe --tags)
COMMIT = $(shell git rev-parse HEAD)
SHA = $(shell git rev-parse --short HEAD)
CURR_DIR = $(shell pwd)
CURR_DIR_WIN = $(shell cd)
BIN_DIR = $(CURR_DIR)/build
BIN_DIR_WIN = $(CURR_DIR_WIN)/build
export GO111MODULE = on

BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

GOLANGCI_LINT_VERSION := v1.60.3

# Set BRANCH when running make manually
ifeq ($(BRANCH),)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
endif

# Setup the -ldflags option to pass vars defined here to app vars
LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.branch=${BRANCH}"
PKGS = $(shell go list ./...)

PLATFORMS := windows linux darwin
os = $(word 1, $@)

.PHONY: install
install:
	go mod download
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LINT_VERSION)

.PHONY: all
all:

.PHONY: lint-ci
lint-ci:
	golangci-lint run ./...

.PHONY: lint
lint:
	./bin/golangci-lint run --config .golangci.yml

.PHONY: lint-fix
lint-fix:
	./bin/golangci-lint run --config .golangci.yml --fix

.PHONY: test
test: vet lint

.PHONY: vet
vet:
	go vet ./...

.PHONY: dev_up
dev_up: ## start local environment
	@echo "RUN dev docker-compose.yml "
	docker compose pull
	docker compose up --build

.PHONY: ci_up
ci_up: ## start ci environment
	@echo "RUN ci docker-compose.yml "
	docker compose up --build -d

.PHONY: gogen
gogen: ## generate scalegen
	go generate ./...