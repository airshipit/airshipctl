DOCKER_REGISTRY     ?= quay.io
DOCKER_IMAGE_PREFIX ?= airshipit
DOCKER_IMAGE_TAG    ?= latest
IMAGES              ?= infra-builder runner
PUBLISH             ?= false

.PHONY: help base libvirt artifact-setup $(IMAGES) images test

SHELL:=/bin/bash
.ONESHELL:

help: ## This help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

images: artifact-setup
images: base
images: libvirt
images: $(IMAGES) ## Build the containers.

artifact-setup:
	docker build --tag $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/aiap-artifact-setup:$(DOCKER_IMAGE_TAG) --build-arg AIRSHIPCTL_REF=$(AIRSHIPCTL_REF) ./artifact-setup
ifeq (true, $(PUBLISH))
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/aiap-artifact-setup:$(DOCKER_IMAGE_TAG)
endif

base:
	docker build --tag $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/aiap-base:$(DOCKER_IMAGE_TAG) --build-arg BASE_IMAGE=ubuntu:20.04 ./base
ifeq (true, $(PUBLISH))
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/aiap-base:$(DOCKER_IMAGE_TAG)
endif

libvirt:
	docker build --tag $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/libvirt:$(DOCKER_IMAGE_TAG) ./libvirt
ifeq (true, $(PUBLISH))
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/libvirt:$(DOCKER_IMAGE_TAG)
endif

$(IMAGES):
	docker build --tag $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/aiap-$@:$(DOCKER_IMAGE_TAG) ./$@
ifeq (true, $(PUBLISH))
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_PREFIX)/aiap-$@:$(DOCKER_IMAGE_TAG)
endif

test: test-airshipctl test-treasuremap

test-airshipctl: images
	kubectl delete pod airship-in-a-pod || true
	kustomize build ./examples/airshipctl | kubectl apply -f -

test-treasuremap: images
	kubectl delete pod airship-in-a-pod || true
	kustomize build ./examples/airshipctl | kubectl apply -f -
