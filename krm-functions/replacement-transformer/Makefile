.PHONY: generate license fix vet fmt test build tidy image

SHELL := /bin/bash
GOBIN := $(shell go env GOPATH)/bin

# docker image options
DOCKER_REGISTRY     ?= quay.io
DOCKER_IMAGE_NAME   ?= replacement-transformer
DOCKER_IMAGE_PREFIX ?= airshipit
DOCKER_IMAGE_TAG    ?= dev
DOCKER_IMAGE        ?= $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
PUBLISH             ?= false
DOCKER_FORCE_CLEAN  ?= true

# proxy options
PROXY               ?= http://proxy.foo.com:8000
NO_PROXY            ?= localhost,127.0.0.1,.svc.cluster.local
USE_PROXY           ?= false

.PHONY: build
build:
	(cd image && go build -v -o $(GOBIN)/config-function .)

.PHONY: all
all: generate license build fix vet fmt test lint tidy

.PHONY: fix
fix:
	(cd image && go fix ./...)

.PHONY: fmt
fmt:
	(cd image && go fmt ./...)

.PHONY: generate
generate:
	(which $(GOBIN)/mdtogo || go get sigs.k8s.io/kustomize/cmd/mdtogo)
	(cd image && GOBIN=$(GOBIN) go generate ./...)

.PHONY: tidy
tidy:
	(cd image && go mod tidy)

.PHONY: fix
lint:
	(which $(GOBIN)/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1)
	(cd image && $(GOBIN)/golangci-lint run ./...)

.PHONY: test
test:
	(cd image && go test -cover ./...)

.PHONY: vet
vet:
	(cd image && go vet ./...)

.PHONY: image
image:
ifeq ($(USE_PROXY), true)
	cd image && \
	docker build . --network=host \
		--build-arg http_proxy=$(PROXY) \
		--build-arg https_proxy=$(PROXY) \
		--build-arg HTTP_PROXY=$(PROXY) \
		--build-arg HTTPS_PROXY=$(PROXY) \
		--build-arg no_proxy=$(NO_PROXY) \
		--build-arg NO_PROXY=$(NO_PROXY) \
	  --tag $(DOCKER_IMAGE) \
	  --force-rm=$(DOCKER_FORCE_CLEAN)
else
	cd image && \
	docker build . --network=host \
	    --tag $(DOCKER_IMAGE) \
	    --force-rm=$(DOCKER_FORCE_CLEAN)
endif
ifeq ($(PUBLISH), true)
	@docker push $(DOCKER_IMAGE)
endif
