SHELL := /bin/bash

GIT_VERSION         ?= v0.1.0
GIT_COMMIT          ?= $(shell git rev-parse HEAD)
BUILD_DATE          ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_MODULE          ?= opendev.org/airship/airshipctl/pkg/version

LDFLAGS             += -X ${GIT_MODULE}.gitVersion=${GIT_VERSION}
LDFLAGS             += -X ${GIT_MODULE}.gitCommit=${GIT_COMMIT}
LDFLAGS             += -X ${GIT_MODULE}.buildDate=${BUILD_DATE}

GO_FLAGS            := -ldflags '-extldflags "-static"' -tags=netgo -trimpath
GO_FLAGS            += -ldflags '$(LDFLAGS)'
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN 2> /dev/null))
GOBIN = $(shell go env GOPATH 2> /dev/null)/bin
else
GOBIN = $(shell go env GOBIN 2> /dev/null)
endif

# Produce CRDs that work back to Kubernetes 1.21.2
CRD_OPTIONS ?= crd:crdVersions=v1
TOOLBINDIR          := tools/bin

# linting
LINTER              := $(TOOLBINDIR)/golangci-lint
LINTER_CONFIG       := .golangci.yaml

# docker
DOCKER_MAKE_TARGET  := build
DOCKER_CMD_FLAGS    :=

# docker image options
DOCKER_REGISTRY     ?= quay.io
DOCKER_FORCE_CLEAN  ?= true
DOCKER_IMAGE_NAME   ?= airshipctl
DOCKER_IMAGE_PREFIX ?= airshipit
DOCKER_IMAGE_TAG    ?= latest
DOCKER_IMAGE        ?= $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
DOCKER_TARGET_STAGE ?= release
PUBLISH             ?= false
# use this variable for image labels added in internal build process
COMMIT              ?= $(shell git rev-parse HEAD)
LABEL               ?= org.airshipit.build=community
LABEL               += --label "org.opencontainers.image.revision=$(COMMIT)"
LABEL               += --label "org.opencontainers.image.created=$(shell date --rfc-3339=seconds --utc)"
LABEL               += --label "org.opencontainers.image.title=$(DOCKER_IMAGE_NAME)"

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

# docker build flags
DOCKER_CMD_FLAGS    += --network=host
DOCKER_CMD_FLAGS    += --force-rm=$(DOCKER_FORCE_CLEAN)
ifeq ($(USE_PROXY), true)
DOCKER_CMD_FLAGS    += --build-arg http_proxy=$(PROXY)
DOCKER_CMD_FLAGS    += --build-arg https_proxy=$(PROXY)
DOCKER_CMD_FLAGS    += --build-arg HTTP_PROXY=$(PROXY)
DOCKER_CMD_FLAGS    += --build-arg HTTPS_PROXY=$(PROXY)
DOCKER_CMD_FLAGS    += --build-arg no_proxy=$(NO_PROXY)
DOCKER_CMD_FLAGS    += --build-arg NO_PROXY=$(NO_PROXY)
endif
ifneq ($(strip $(GOPROXY)),)
DOCKER_CMD_FLAGS    += --build-arg GOPROXY=$(strip $(GOPROXY))
endif

# Godoc server options
GD_PORT             ?= 8080

# Documentation location
DOCS_DIR            ?= docs

# document validation options
UNAME               != uname
export KIND_URL     ?= https://kind.sigs.k8s.io/dl/v0.8.1/kind-$(UNAME)-amd64
KUBECTL_VERSION     ?= v1.21.2
export KUBECTL_URL  ?= https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl

.PHONY: depend
depend:
	@go mod download

.PHONY: build

.PHONY: install
install: depend
install:
	@CGO_ENABLED=0 go install .

# Core of build logic
BIN_DIR      := bin
BIN_SRC_DIR  := krm-functions
BINS         := airshipctl $(subst $(BIN_SRC_DIR)/,,$(wildcard $(BIN_SRC_DIR)/*))
IMGS         := $(BINS)

# This section sets the settings for different subcomponents

# airshipctl is a special case - we need to override it manually:
# its makefile target for image is 'docker-image' - others have
# docker-image-<name of component> targets
airshipctl_IMG_TGT_NAME:=docker-image
# its main.go is in the root of repo - others have main.go in
# $(BIN_SRC_DIR)/<name of component>/main.go
airshipctl_FROM_PATH:=.
# and its Dockerfile is also in the root of repo - others have Dockerfile in
# $(BIN_SRC_DIR)/<name of component>/Dockerfile
docker-image_DOCKERFILE:=Dockerfile

# kubeval-validator, toolbox and toolbox-virsh don't depend on
# airshipctl repo. Their Dockerfiles don't
# need to be called from the root of the repo.
applier_IS_INDEPENDED:=true
kubeval-validator_IS_INDEPENDED:=true
clusterctl_IS_INDEPENDED:=true
clusterctl-v0.3_IS_INDEPENDED:=true
toolbox-virsh_IS_INDEPENDED:=true
# in addition toolbox-virsh docker image needs toolbox docker image to be built first
docker-image-clusterctl-v0.3_DEPENDENCY:=docker-image-clusterctl
docker-image-toolbox-virsh_DEPENDENCY:=docker-image-toolbox

# The template that generates targets for creating binaries per component:
# Targets will be generated only for components that depend on airshipctl repo (part of that go module)
# Note: expressions with ?= won't be executed if the values of that variable was already set to it.
# Using that syntax it's possible to build values overrides for components.
# Note 2: $$ is needed to instruct make-engine that variable should be used after template rendering.
# When template is rendered all $ will be rendered in the template and $$ will be converted to $, e.g.
# if we call map_binary_defaults_tmpl for airshipctl $1 will be converted to 'airshipctl' and we'll get
# ifneq ($(airshipctl_IS_INDEPENDED),true)
# airshipctl_FROM_PATH?=$(BIN_SRC_DIR)/airshipctl/main.go
# ...
# since we defining airshipctl_FROM_PATH above, and ?= is used in the 2nd line
# airshipctl_FROM_PATH will stay the same as it was defined above.
define map_binary_defaults_tmpl
ifneq ($$($1_IS_INDEPENDED),true)
$1_FROM_PATH?=$$(BIN_SRC_DIR)/$1/main.go

$$(warning Adding dynamic target $$(BIN_DIR)/$1)
$$(BIN_DIR)/$1: $$($1_FROM_PATH) depend
	@CGO_ENABLED=0 go build -o $$@ $$(GO_FLAGS) $$<

$$(warning Adding dynamic target $1)
.PHONY: $1
$1: $$(BIN_DIR)/$1

build: $1
endif
endef
map_binary_defaults = $(eval $(call map_binary_defaults_tmpl,$1))
# Go through all components and generate binary targets for each of them
$(foreach bin,$(BINS),$(call map_binary_defaults,$(bin)))

.PHONY: images
.PHONY: images-publish

# The template that generates targets for creating images per components
# There is a special logic to handle per-components overrides
# 2 targets will be generated per component: docker-image-<component name> (possible to override)
# and docker-image-<component name>-publish
define map_image_defaults_tmpl
$1_IMG_TGT_NAME?=docker-image-$1

$$($1_IMG_TGT_NAME)_IMG_TITLE?=$1
$$($1_IMG_TGT_NAME)_IMG_TAG?=$$(DOCKER_IMAGE_TAG)
$$($1_IMG_TGT_NAME)_DOCKERTGT?=$$(DOCKER_TARGET_STAGE)
$$($1_IMG_TGT_NAME)_DOCKERFILE?=$$(BIN_SRC_DIR)/$1/Dockerfile
$$($1_IMG_TGT_NAME)_MAKETGT?=$$(BIN_DIR)/$1

ifeq ($$($1_IS_INDEPENDED),true)
$$($1_IMG_TGT_NAME)_DOCKERROOT?=$$(BIN_SRC_DIR)/$1
else
$$($1_IMG_TGT_NAME)_DOCKERROOT?=.
endif

ifneq ($1,airshipctl)
ifneq ($$(origin DOCKER_BASE_PLUGINS_GO_IMAGE), undefined)
$$($1_IMG_TGT_NAME)_BASE_GO_IMAGE?=$$(DOCKER_BASE_PLUGINS_GO_IMAGE)
endif
endif
$$($1_IMG_TGT_NAME)_BASE_GO_IMAGE?=$$(DOCKER_BASE_GO_IMAGE)
ifneq ($$(strip $$($$($1_IMG_TGT_NAME)_BASE_GO_IMAGE)),)
$$($1_IMG_TGT_NAME)_BUILD_ARG  += GO_IMAGE=$$($$($1_IMG_TGT_NAME)_BASE_GO_IMAGE)
endif

ifneq ($1,airshipctl)
ifneq ($$(origin DOCKER_BASE_PLUGINS_BUILD_IMAGE), undefined)
$$($1_IMG_TGT_NAME)_BASE_BUILD_IMAGE?=$$(DOCKER_BASE_PLUGINS_BUILD_IMAGE)
endif
endif
$$($1_IMG_TGT_NAME)_BASE_BUILD_IMAGE?=$$(DOCKER_BASE_BUILD_IMAGE)
ifneq ($$(strip $$($$($1_IMG_TGT_NAME)_BASE_BUILD_IMAGE)),)
$$($1_IMG_TGT_NAME)_BUILD_ARG  += BUILD_IMAGE=$$($$($1_IMG_TGT_NAME)_BASE_BUILD_IMAGE)
endif

ifneq ($1,airshipctl)
ifneq ($$(origin DOCKER_BASE_PLUGINS_RELEASE_IMAGE), undefined)
$$($1_IMG_TGT_NAME)_BASE_RELEASE_IMAGE?=$$(DOCKER_BASE_PLUGINS_RELEASE_IMAGE)
endif
endif
$$($1_IMG_TGT_NAME)_BASE_RELEASE_IMAGE?=$$(DOCKER_BASE_RELEASE_IMAGE)
ifneq ($$(strip $$($$($1_IMG_TGT_NAME)_BASE_RELEASE_IMAGE)),)
$$($1_IMG_TGT_NAME)_BUILD_ARG  += RELEASE_IMAGE=$$($$($1_IMG_TGT_NAME)_BASE_RELEASE_IMAGE)
endif

ifeq ($1,clusterctl-v0.3)
$$($1_IMG_TGT_NAME)_IMG_TAG=v0.3
$$($1_IMG_TGT_NAME)_IMG_TITLE=clusterctl
endif

$$(warning Adding dynamic target $$($1_IMG_TGT_NAME))
.PHONY: $$($1_IMG_TGT_NAME)
$$($1_IMG_TGT_NAME): $$($$($1_IMG_TGT_NAME)_DEPENDENCY)
	docker build $$($$($1_IMG_TGT_NAME)_DOCKERROOT) $$(DOCKER_CMD_FLAGS)\
		--file $$($$($1_IMG_TGT_NAME)_DOCKERFILE) \
		--label $$(LABEL) \
		--label "org.opencontainers.image.revision=$$(COMMIT)" \
		--label "org.opencontainers.image.created=$$(shell date --rfc-3339=seconds --utc)" \
		--label "org.opencontainers.image.title=$$($$($1_IMG_TGT_NAME)_IMG_TITLE)" \
		--target $$($$($1_IMG_TGT_NAME)_DOCKERTGT) \
		$$(addprefix --build-arg ,$$($$($1_IMG_TGT_NAME)_BUILD_ARG)) \
		--build-arg MAKE_TARGET=$$($$($1_IMG_TGT_NAME)_MAKETGT) \
		--tag $$(DOCKER_REGISTRY)/$$(DOCKER_IMAGE_PREFIX)/$$($$($1_IMG_TGT_NAME)_IMG_TITLE):$$($$($1_IMG_TGT_NAME)_IMG_TAG) \
		$$(foreach tag,$$(DOCKER_IMAGE_EXTRA_TAGS),--tag $$(DOCKER_REGISTRY)/$$(DOCKER_IMAGE_PREFIX)/$1:$$(tag) )
ifeq ($$(PUBLISH), true)
	@docker push $$(DOCKER_REGISTRY)/$$(DOCKER_IMAGE_PREFIX)/$$($$($1_IMG_TGT_NAME)_IMG_TITLE):$$($$($1_IMG_TGT_NAME)_IMG_TAG)
endif

images: $$($1_IMG_TGT_NAME)

$$(warning Adding dynamic target $$($1_IMG_TGT_NAME)-publish)
.PHONY: $$($1_IMG_TGT_NAME)-publish
$$($1_IMG_TGT_NAME)-publish: $$($1_IMG_TGT_NAME)
	@docker push $$(DOCKER_REGISTRY)/$$(DOCKER_IMAGE_PREFIX)/$1:$$(DOCKER_IMAGE_TAG)

images-publish: $$($1_IMG_TGT_NAME)-publish
endef
map_image_defaults = $(eval $(call map_image_defaults_tmpl,$1))
# go through components and render the template
$(foreach img,$(IMGS),$(call map_image_defaults,$(img)))

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

.PHONY: print-docker-image-tag
print-docker-image-tag:
	@echo "$(DOCKER_IMAGE)"

.PHONY: docker-image-test-suite
docker-image-test-suite: docker-image_MAKETGT = "cover update-golden generate check-git-diff"
docker-image-test-suite: docker-image_DOCKERTGT = builder
docker-image-test-suite: docker-image

.PHONY: docker-image-unit-tests
docker-image-unit-tests: docker-image_MAKETGT = cover
docker-image-unit-tests: docker-image_DOCKERTGT = builder
docker-image-unit-tests: docker-image

.PHONY: docker-image-lint
docker-image-lint: docker-image_MAKETGT = "lint check-copyright"
docker-image-lint: docker-image_DOCKERTGT = builder
docker-image-lint: docker-image

.PHONY: docker-image-golint
docker-image-golint: docker-image_MAKETGT = golint
docker-image-golint: docker-image_DOCKERTGT = builder
docker-image-golint: docker-image

.PHONY: clean
clean:
	@rm -fr $(BIN_DIR)
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

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="tools/license_go.txt" paths="./..."

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.1 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=manifests/function/airshipctl-schemas
