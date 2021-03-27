IMAGE_REGISTRY    ?= quay.io/airshipit
IMAGES            := infra-builder artifact-setup runner
IMAGE_TAG         ?= latest

PUSH_IMAGES       ?= false

.PHONY: help base libvirt $(IMAGES) build test

SHELL:=/bin/bash
.ONESHELL:

help: ## This help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: base
build: libvirt
build: $(IMAGES) ## Build the containers.

base:
	docker build --tag ianhowell/base:$(IMAGE_TAG) --build-arg BASE_IMAGE=ubuntu:20.04 ./base
ifeq (true, $(PUSH_IMAGES))
	docker push ianhowell/base:$(IMAGE_TAG)
endif

libvirt:
	docker build --tag ianhowell/libvirt:$(IMAGE_TAG) ./libvirt
ifeq (true, $(PUSH_IMAGES))
	docker push ianhowell/libvirt:$(IMAGE_TAG)
endif

$(IMAGES):
	docker build --tag $(IMAGE_REGISTRY)/aiap-$@:$(IMAGE_TAG) ./$@
ifeq (true, $(PUSH_IMAGES))
	docker push $(IMAGE_REGISTRY)/aiap-$@:$(IMAGE_TAG)
endif

test: test-airshipctl test-treasuremap

test-airshipctl: build
	kubectl delete pod  airship-in-a-pod || true
	kustomize build ./examples/airshipctl | kubectl apply -f -

test-treasuremap: build
	kubectl delete pod  airship-in-a-pod || true
	kustomize build ./examples/airshipctl | kubectl apply -f -
