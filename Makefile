VERSION := 0.1.0
COMMIT = $(shell git rev-parse HEAD)
SHA = $(shell git rev-parse --short HEAD)
CURR_DIR = $(shell pwd)
CURR_DIR_WIN = $(shell cd)
BIN_DIR = $(CURR_DIR)/build
BIN_DIR_WIN = $(CURR_DIR_WIN)/build
export GO111MODULE = on

BRANCH := $(shell bash -c 'if [ "$$TRAVIS_PULL_REQUEST" == "false" ]; then echo $$TRAVIS_BRANCH; else echo $$TRAVIS_PULL_REQUEST_BRANCH; fi')

# Set BRANCH when running make manually
ifeq ($(BRANCH),)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
endif

# Setup the -ldflags option to pass vars defined here to app vars
LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.branch=${BRANCH}"

PKGS = $(shell go list ./...)

PLATFORMS := windows linux darwin
os = $(word 1, $@)

all:
.PHONY: all

apiserver:
ifeq ($(OS),Windows_NT)
	cd cmd/apiserver ; go build -o $(BIN_DIR_WIN)/apiserver.exe; cd ..
else
	cd cmd/apiserver ; go build -o $(BIN_DIR)/apiserver; cd ..
endif
.PHONY: apiserver

collector:
ifeq ($(OS),Windows_NT)
	cd cmd/collector ; go build -o $(BIN_DIR_WIN)/collector.exe; cd ..
else
	cd cmd/collector ; go build -o $(BIN_DIR)/collector; cd ..
endif
.PHONY: collector

lint-ci:
	golangci-lint run --new-from-rev=origin/master --config .golangci.yml

lint:
	golangci-lint run --new-from-rev=master --config .golangci.yml

lint-fix:
	golangci-lint run --new-from-rev=master --config .golangci.yml --fix

test:
	go test -race  ./...

stop:
	@echo "-- stop containers";
	docker container ls -f "name=sm_*" ; true
	@echo "-- drop containers"
	docker rm -f -v $(shell docker container ls -f "name=sm_*" -q) ; true

dev_up: stop ## start local environment
	@echo "RUN dev docker-compose.yml "
	docker compose pull
	docker compose up --build

ci_up: stop ## start ci environment
	@echo "RUN ci docker-compose.yml "
	docker compose up --build -d