SHELL := /bin/bash

GO_FLAGS            := -ldflags '-extldflags "-static"' -tags=netgo

BINDIR              := bin
EXECUTABLE_CLI      := airshipctl
TOOLBINDIR          := tools/bin

# linting
LINTER              := $(TOOLBINDIR)/golangci-lint
LINTER_CONFIG       := .golangci.yaml

# docker
DOCKER_MAKE_TARGET  := build

# docker image options
DOCKER_REGISTRY     ?= quay.io
DOCKER_IMAGE_NAME   ?= airshipctl
DOCKER_IMAGE_PREFIX ?= airshipit
DOCKER_IMAGE_TAG    ?= dev
DOCKER_IMAGE        ?= $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
DOCKER_TARGET_STAGE ?= release

# go options
PKG                 ?= ./...
TESTS               ?= .
TEST_FLAGS          ?=
COVER_FLAGS         ?=
COVER_PROFILE       ?= cover.out

# proxy options
PROXY               ?= http://proxy.foo.com:8000
NO_PROXY            ?= localhost,127.0.0.1,.svc.cluster.local
USE_PROXY           ?= false

.PHONY: get-modules
get-modules:
	@go mod download

.PHONY: build
build: get-modules
	@CGO_ENABLED=0 go build -o $(BINDIR)/$(EXECUTABLE_CLI) $(GO_FLAGS)

.PHONY: test
test: lint
test: cover

.PHONY: unit-tests
unit-tests: TESTFLAGS += -race -v
unit-tests:
	@echo "Performing unit test step..."
	@go test -run $(TESTS) $(PKG) $(TESTFLAGS) $(COVER_FLAGS)
	@echo "All unit tests passed"

.PHONY: cover
cover: COVER_FLAGS = -covermode=atomic -coverprofile=$(COVER_PROFILE)
cover: unit-tests
	@./tools/coverage_check $(COVER_PROFILE)

.PHONY: lint
lint: tidy
lint: $(LINTER)
	@echo "Performing linting step..."
	@./$(LINTER) run --config $(LINTER_CONFIG)
	@echo "Linting completed successfully"

.PHONY: tidy
tidy:
	@echo "Checking that go.mod is up to date..."
	@./tools/gomod_check
	@echo "go.mod is up to date"

.PHONY: docker-image
docker-image:
ifeq ($(USE_PROXY), true)
	@docker build . --network=host \
		--build-arg http_proxy=$(PROXY) \
		--build-arg https_proxy=$(PROXY) \
		--build-arg HTTP_PROXY=$(PROXY) \
		--build-arg HTTPS_PROXY=$(PROXY) \
		--build-arg no_proxy=$(NO_PROXY) \
		--build-arg NO_PROXY=$(NO_PROXY) \
	    --build-arg MAKE_TARGET=$(DOCKER_MAKE_TARGET) \
	    --tag $(DOCKER_IMAGE) \
	    --target $(DOCKER_TARGET_STAGE)
else
	@docker build . --network=host \
	    --build-arg MAKE_TARGET=$(DOCKER_MAKE_TARGET) \
	    --tag $(DOCKER_IMAGE) \
	    --target $(DOCKER_TARGET_STAGE)
endif


.PHONY: print-docker-image-tag
print-docker-image-tag:
	@echo "$(DOCKER_IMAGE)"

.PHONY: docker-image-unit-tests
docker-image-unit-tests: DOCKER_MAKE_TARGET = cover
docker-image-unit-tests: DOCKER_TARGET_STAGE = builder
docker-image-unit-tests: docker-image

.PHONY: docker-image-lint
docker-image-lint: DOCKER_MAKE_TARGET = lint
docker-image-lint: DOCKER_TARGET_STAGE = builder
docker-image-lint: docker-image

.PHONY: clean
clean:
	@rm -fr $(BINDIR)
	@rm -fr $(COVER_PROFILE)

.PHONY: docs
docs:
	@echo "TODO"

$(TOOLBINDIR):
	mkdir -p $(TOOLBINDIR)

$(LINTER): $(TOOLBINDIR)
	./tools/install_linter

.PHONY: update-golden
update-golden: delete-golden
update-golden: TESTFLAGS += -update
update-golden: PKG = opendev.org/airship/airshipctl/cmd/...
update-golden: unit-tests

# The delete-golden target is a utility for update-golden
.PHONY: delete-golden
delete-golden:
	@find . -type f -name "*.golden" -delete
