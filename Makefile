SHELL := /bin/bash

GIT_VERSION         ?= v0.1.0
GIT_MODULE          ?= opendev.org/airship/airshipctl/pkg/version

GO_FLAGS            := -ldflags '-extldflags "-static"' -tags=netgo
GO_FLAGS            += -ldflags "-X ${GIT_MODULE}.gitVersion=${GIT_VERSION}"

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
DOCKER_FORCE_CLEAN  ?= true
DOCKER_IMAGE_NAME   ?= airshipctl
DOCKER_IMAGE_PREFIX ?= airshipit
DOCKER_IMAGE_TAG    ?= dev
DOCKER_IMAGE        ?= $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
DOCKER_TARGET_STAGE ?= release
PUBLISH             ?= false

# go options
PKG                 ?= ./...
TESTS               ?= .
TEST_FLAGS          ?=
COVER_FLAGS         ?=
COVER_PROFILE       ?= cover.out
COVER_EXCLUDE       ?= (zz_generated|errors)

# proxy options
PROXY               ?= http://proxy.foo.com:8000
NO_PROXY            ?= localhost,127.0.0.1,.svc.cluster.local
USE_PROXY           ?= false

# Godoc server options
GD_PORT             ?= 8080

# Documentation location
DOCS_DIR            ?= docs

# document validation options
UNAME               != uname
export KIND_URL     ?= https://kind.sigs.k8s.io/dl/v0.8.1/kind-$(UNAME)-amd64
KUBECTL_VERSION     ?= v1.17.4
export KUBECTL_URL  ?= https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl

.PHONY: depend
depend:
	@go mod download

.PHONY: build
build: depend
	@CGO_ENABLED=0 go build -o $(BINDIR)/$(EXECUTABLE_CLI) $(GO_FLAGS)

.PHONY: install
install: depend
install:
	@CGO_ENABLED=0 go install .

.PHONY: test
test: lint
test: cover
test: check-copyright

.PHONY: unit-tests
unit-tests: TESTFLAGS += -race -v
unit-tests:
	@echo "Performing unit test step..."
	@go test -run $(TESTS) $(PKG) $(TESTFLAGS) $(COVER_FLAGS)
	@echo "All unit tests passed"

.PHONY: cover
cover: COVER_FLAGS = -covermode=atomic -coverprofile=fullcover.out
cover: unit-tests
	@grep -vE "$(COVER_EXCLUDE)" fullcover.out > $(COVER_PROFILE)
	@./tools/coverage_check $(COVER_PROFILE)

.PHONY: fmt
fmt: lint

.PHONY: lint
lint: tidy
lint: $(LINTER)
	@echo "Performing linting step..."
	@./tools/whitespace_linter
	@./$(LINTER) run --config $(LINTER_CONFIG)
	@echo "Linting completed successfully"

.PHONY: tidy
tidy:
	@echo "Checking that go.mod is up to date..."
	@./tools/gomod_check
	@echo "go.mod is up to date"

.PHONY: golint
golint:
	@./tools/golint

.PHONY: images
images: docker-image

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
	    --target $(DOCKER_TARGET_STAGE) \
	    --force-rm=$(DOCKER_FORCE_CLEAN)
else
	@docker build . --network=host \
	    --build-arg MAKE_TARGET=$(DOCKER_MAKE_TARGET) \
	    --tag $(DOCKER_IMAGE) \
	    --target $(DOCKER_TARGET_STAGE) \
	    --force-rm=$(DOCKER_FORCE_CLEAN)
endif
ifeq ($(PUBLISH), true)
	@docker push $(DOCKER_IMAGE)
endif


.PHONY: print-docker-image-tag
print-docker-image-tag:
	@echo "$(DOCKER_IMAGE)"

.PHONY: docker-image-test-suite
docker-image-test-suite: DOCKER_MAKE_TARGET = "lint cover update-golden check-git-diff check-copyright"
docker-image-test-suite: DOCKER_TARGET_STAGE = builder
docker-image-test-suite: docker-image

.PHONY: docker-image-unit-tests
docker-image-unit-tests: DOCKER_MAKE_TARGET = cover
docker-image-unit-tests: DOCKER_TARGET_STAGE = builder
docker-image-unit-tests: docker-image

.PHONY: docker-image-lint
docker-image-lint: DOCKER_MAKE_TARGET = lint
docker-image-lint: DOCKER_TARGET_STAGE = builder
docker-image-lint: docker-image

.PHONY: docker-image-golint
docker-image-golint: DOCKER_MAKE_TARGET = golint
docker-image-golint: DOCKER_TARGET_STAGE = builder
docker-image-golint: docker-image

.PHONY: clean
clean:
	@rm -fr $(BINDIR)
	@rm -fr $(COVER_PROFILE)

.PHONY: docs
docs:
	tox

.PHONY: godoc
godoc:
	@go install golang.org/x/tools/cmd/godoc
	@echo "Follow this link to package documentation: http://localhost:${GD_PORT}/pkg/opendev.org/airship/airshipctl/"
	@godoc -http=":${GD_PORT}"

.PHONY: cli-docs
cli-docs:
	@echo "Generating CLI documentation..."
	@go run $(DOCS_DIR)/tools/generate_cli_docs.go
	@echo "CLI documentation generated"

.PHONY: releasenotes
releasenotes:
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
update-golden: cli-docs

# The delete-golden target is a utility for update-golden
.PHONY: delete-golden
delete-golden:
	@find . -type f -name "*.golden" -delete

# Used by gates after unit-tests and update-golden targets to ensure no files are deleted.
.PHONY: check-git-diff
check-git-diff:
	@./tools/git_diff_check

# add-copyright is a utility to add copyright header to missing files
.PHONY: add-copyright
add-copyright:
	@./tools/add_license.sh

# check-copyright is a utility to check if copyright header is present on all files
.PHONY: check-copyright
check-copyright:
	@./tools/check_copyright

# Validate YAMLs for all sites
.PHONY: validate-docs
validate-docs:
	@./tools/validate_docs
