#
#   make              - default to 'build' target
#   make test         - run unit test
#   make build        - build local binary targets
#   make docker-build - build local binary targets by docker
#   make container    - build containers
#   make push         - push containers
#   make clean        - clean up targets
#
# The makefile is also responsible to populate project version information.

#
# Tweak the variables based on your project.
#

# Current version of the project.
VERSION ?= v0.0.3

# Target binaries. You can build multiple binaries for a single project.
TARGETS := sym-operator

# Container registries.
# REGISTRIES := hub.tencentyun.com/xkcp0324/
REGISTRIES := hub.tencentyun.com/demo001/

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip )
IMAGE_SUFFIX ?= $(strip )

# This repo's root import path (under GOPATH).
ROOT := github.com/yangyongzhi/sym-operator

# Project main package location (can be multiple ones).
CMD_DIR := ./cmd

# Project output directory.
OUTPUT_DIR := ./bin

# docker file direcotory.
DOCKER_DIR := ./docker

# Git commit sha.
COMMIT := $(strip $(shell git rev-parse --short HEAD 2>/dev/null))
COMMIT := $(COMMIT)$(shell git diff-files --quiet || echo '-dirty')
COMMIT := $(if $(COMMIT),$(COMMIT),"Unknown")


GO_VERSION := 1.12.5
ARCH     ?= $(shell go env GOARCH)
BuildDate = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
Commit    = $(shell git rev-parse --short HEAD)
GOENV     := GO111MODULE=on CGO_ENABLED=0 GOOS=$(shell uname -s | tr A-Z a-z) GOARCH=$(ARCH)
GO        := $(GOENV) go build -mod=vendor

#
# Define all targets. At least the following commands are required:
#

.PHONY: build container push test clean

build:
	$(GO) -v -o $(OUTPUT_DIR)/$(TARGETS)     -ldflags "-s -w "	   .

mod-reset-vendor:
	@$(shell [ -f go.mod ] && go mod vendor)

docker-build:
	docker run --rm -v "$$PWD":/go/src/${ROOT} -w /go/src/${ROOT}                      \
	hub.tencentyun.com/xkcp0324/golang:${GO_VERSION} make build

container:
	@for target in $(TARGETS); do                                                      \
	  for registry in $(REGISTRIES); do                                                \
	    image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                \
	    docker build -t $${registry}$${image}:$(VERSION)                               \
	      --build-arg ROOT=$(ROOT) --build-arg TARGET=$${target}                       \
	      --build-arg CMD_DIR=$(CMD_DIR)                                               \
	      --build-arg VERSION=$(VERSION)                                               \
	      --build-arg COMMIT=$(COMMIT)                                                 \
	      -f $(DOCKER_DIR)/$${target}/Dockerfile .;                                    \
	  done                                                                             \
	done

push: container
	@for target in $(TARGETS); do                                                      \
	  for registry in $(REGISTRIES); do                                                \
	    image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                \
	    docker push $${registry}$${image}:$(VERSION);                                  \
	  done                                                                             \
	done

test:
	@go test ./...

clean:
	@rm -vrf ${OUTPUT_DIR}/*
