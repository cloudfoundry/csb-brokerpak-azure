SHELL = /bin/bash
GO-VERSION = 1.18.3
GO-VER = go$(GO-VERSION)

SRC = $(shell find . -name "*.go" | grep -v "_test\." )
VERSION=0.1.0

GO_OK :=  $(or $(USE_GO_CONTAINERS), $(shell which go 1>/dev/null 2>/dev/null; echo $$?))
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)
ifeq ($(GO_OK), 0)
GO=go
GOIMPORTS=goimports
else ifeq ($(DOCKER_OK), 0)
UID:=$(shell id -u)
DOCKER_OPTS=--rm -v $(PWD):/src -w /src
GO=docker run $(DOCKER_OPTS) -e GOARCH -e GOOS golang:$(GO-VERSION) go
GOTOOLS=docker run $(DOCKER_OPTS) jare/go-tools
GOIMPORTS=$(GOTOOLS) goimports
HAS_GO_IMPORTS=true
else
$(error either Go or Docker must be installed)
endif

EXE_NAME=sqlfailover

.DEFAULT_TARGET: build

.PHONY: deps-go-binary
deps-go-binary:
ifeq ($(SKIP_GO_VERSION_CHECK),)
	@@if [ "$$($(GO) version | awk '{print $$3}')" != "${GO-VER}" ]; then \
		echo "Go version does not match: expected: ${GO-VER}, got $$($(GO) version | awk '{print $$3}')"; \
		exit 1; \
	fi
endif

HAS_GO_IMPORTS := $(shell command -v goimports;)

deps-goimports: deps-go-binary
ifndef HAS_GO_IMPORTS
	go get -u golang.org/x/tools/cmd/goimports
endif

.PHONY: build
build: deps-go-binary ./build/$(EXE_NAME)_$(VERSION)_linux_amd64.zip ./build/$(EXE_NAME)_$(VERSION)_darwin_amd64.zip

./build/$(EXE_NAME)_$(VERSION)_linux_amd64.zip: build/linux/$(EXE_NAME)
	zip -j -r $@ $<

./build/$(EXE_NAME)_$(VERSION)_darwin_amd64.zip: build/darwin/$(EXE_NAME)
	zip -j -r $@ $<

./build/linux/$(EXE_NAME): $(SRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GO) build -o $@

./build/darwin/$(EXE_NAME): $(SRC)
	GOARCH=amd64 GOOS=darwin $(GO) build -o $@

.PHONY: lint
lint: deps-goimports
	git ls-files | grep '.go$$' | xargs $(GOIMPORTS) -l -w

.PHONY: test
test: lint

.PHONY: clean
clean:
	rm -rf ./build

.PHONY: rebuild
rebuild: clean build
