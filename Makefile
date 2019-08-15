SHELL := /bin/bash

GO_FLAGS            := -ldflags '-extldflags "-static"' -tags=netgo

BINDIR              := bin
EXECUTABLE_CLI      := airshipctl

# linting
LINTER_CMD          := "github.com/golangci/golangci-lint/cmd/golangci-lint" run
LINTER_CONFIG       := .golangci.yaml

# docker
DOCKER_MAKE_TARGET  := build

# docker image options
DOCKER_REGISTRY     ?= quay.io
DOCKER_IMAGE_NAME   ?= airshipctl
DOCKER_IMAGE_PREFIX ?= airshipit
DOCKER_IMAGE_TAG    ?= dev
DOCKER_IMAGE        ?= $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

# go options
PKG                 := ./...
TESTS               := .

.PHONY: get-modules
get-modules:
	@go mod download

.PHONY: build
build: get-modules
	@GO111MODULE=on CGO_ENABLED=0 go build -o $(BINDIR)/$(EXECUTABLE_CLI) $(GO_FLAGS)

.PHONY: test
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
	@GO111MODULE=on go run $(LINTER_CMD) --config $(LINTER_CONFIG)
	@echo "Linting completed successfully"

.PHONY: docker-image
docker-image:
	@docker build . --build-arg MAKE_TARGET=$(DOCKER_MAKE_TARGET) --tag $(DOCKER_IMAGE)

.PHONY: print-docker-image-tag
print-docker-image-tag:
	@echo "$(DOCKER_IMAGE)"

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
update-golden: delete-golden
update-golden: TESTFLAGS += -update -v
update-golden: PKG = opendev.org/airship/airshipctl/cmd/...
update-golden: unit-tests

# The delete-golden target is a utility for update-golden
.PHONY: delete-golden
delete-golden:
	@find . -type f -name "*.golden" -delete
