SHELL := /bin/bash

GO_FLAGS       := -ldflags '-extldflags "-static"' -tags=netgo

BINDIR         := bin
EXECUTABLE_CLI := airshipctl

SCRIPTS_DIR    := scripts

# linting
LINTER_CMD     := "github.com/golangci/golangci-lint/cmd/golangci-lint" run
ADDTL_LINTERS  := goconst,gofmt,unparam

# docker
DOCKER_MAKE_TARGET := build

# go options
PKG          := ./...
TESTS        := .

.PHONY: get-modules
get-modules:
	@go mod download

.PHONY: build
build: get-modules
	@GO111MODULE=on CGO_ENABLED=0 go build -o $(BINDIR)/$(EXECUTABLE_CLI) $(GO_FLAGS)

.PHONY: test
test: build
test: lint
test: TESTFLAGS += -race -v
test: unit-tests

.PHONY: unit-tests
unit-tests: build
	@echo "Performing unit test step..."
	@GO111MODULE=on go test -run $(TESTS) $(PKG) $(TESTFLAGS)
	@echo "All unit tests passed"

.PHONY: lint
lint:
	@echo "Performing linting step..."
	@GO111MODULE=on go run ${LINTER_CMD} --enable ${ADDTL_LINTERS}
	@echo "Linting completed successfully"

.PHONY: docker-image
docker-image:
	@docker build . --build-arg MAKE_TARGET=$(DOCKER_MAKE_TARGET)

.PHONY: docker-image-unit-tests
docker-image-unit-tests: DOCKER_MAKE_TARGET = unit-tests
docker-image-unit-tests: docker-image

.PHONY: docker-image-lint
docker-image-lint: DOCKER_MAKE_TARGET = lint
docker-image-lint: docker-image

.PHONY: clean
clean:
	@rm -fr $(BINDIR)

.PHONY: docs
docs:
	@echo "TODO"

.PHONY: update-golden
update-golden: TESTFLAGS += -update -v
update-golden: PKG = github.com/ian-howell/airshipctl/cmd/...
update-golden:
	@GO111MODULE=on go test $(PKG) $(TESTFLAGS)
