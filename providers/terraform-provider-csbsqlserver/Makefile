.DEFAULT_GOAL = help

GO-VERSION = 1.20.5
GO-VER = go$(GO-VERSION)
GO_OK :=  $(or $(USE_GO_CONTAINERS), $(shell which go 1>/dev/null 2>/dev/null; echo $$?))
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)
ifeq ($(GO_OK), 0)
  GO=go
else ifeq ($(DOCKER_OK), 0)
  GO=docker run --rm -v $(PWD)/../..:/src -w /src/providers/terraform-provider-csbsqlserver -e GOARCH -e GOOS -e CGO_ENABLED golang:$(GO-VERSION) go
else
  $(error either Go or Docker must be installed)
endif

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: deps-go-binary cloudfoundry.org ## build the provider

cloudfoundry.org: *.go */*.go
	mkdir -p cloudfoundry.org/cloud-service-broker/csbsqlserver/1.0.0/linux_amd64
	mkdir -p cloudfoundry.org/cloud-service-broker/csbsqlserver/1.0.0/darwin_amd64
	CGO_ENABLED=0 GOOS=linux $(GO) build -o cloudfoundry.org/cloud-service-broker/csbsqlserver/1.0.0/linux_amd64/terraform-provider-csbsqlserver_v1.0.0
	CGO_ENABLED=0 GOOS=darwin $(GO) build -o cloudfoundry.org/cloud-service-broker/csbsqlserver/1.0.0/darwin_amd64/terraform-provider-csbsqlserver_v1.0.0

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf cloudfoundry.org
	- rm -rf /tmp/tpcsbsqlserver-non-fake.txt
	- rm -rf /tmp/tpcsbsqlserver-pkgs.txt
	- rm -rf /tmp/tpcsbsqlserver-coverage.out

.PHONY: test
test: ## run the tests
	## runs docker, so tricky to make it work inside docker
	go run github.com/onsi/ginkgo/v2/ginkgo -r

.PHONY: test-coverage
test-coverage: ## test coverage score
	go list ./... | grep -v fake > /tmp/tpcsbsqlserver-non-fake.txt
	paste -sd "," /tmp/tpcsbsqlserver-non-fake.txt > /tmp/tpcsbsqlserver-pkgs.txt
	go test -coverpkg=`cat /tmp/tpcsbsqlserver-pkgs.txt` -coverprofile=/tmp/tpcsbsqlserver-coverage.out ./...
	go tool cover -func /tmp/tpcsbsqlserver-coverage.out | grep total

.PHONY: deps-go-binary
deps-go-binary:
ifeq ($(SKIP_GO_VERSION_CHECK),)
	@@if [ "$$($(GO) version | awk '{print $$3}')" != "${GO-VER}" ]; then \
		echo "Go version does not match: expected: ${GO-VER}, got $$($(GO) version | awk '{print $$3}')"; \
		exit 1; \
	fi
endif
