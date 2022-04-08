###### Help ###################################################################
.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Setup ##################################################################
IAAS=azure
CSB_VERSION := $(or $(CSB_VERSION), $(shell grep 'github.com/cloudfoundry/cloud-service-broker' go.mod | grep -v replace | awk '{print $$NF}' | sed -e 's/v//'))
CSB_RELEASE_VERSION := $(CSB_VERSION)

CSB_DOCKER_IMAGE := $(or $(CSB), cfplatformeng/csb:$(CSB_VERSION))
GO_OK :=  $(or $(USE_GO_CONTAINERS), $(shell which go 1>/dev/null 2>/dev/null; echo $$?))
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)

####### broker environment variables
PAK_CACHE=/tmp/.pak-cache
SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), $(IAAS)-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), $(IAAS)-broker-pw)
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"resource_group": "broker-cf-test"})

PARALLEL_JOB_COUNT := $(or $(PARALLEL_JOB_COUNT), 2)

ifeq ($(GO_OK), 0)
GO=go
BROKER_GO_OPTS=PORT=8080 \
				DB_TYPE=sqlite3 \
				DB_PATH=/tmp/csb-db \
				SECURITY_USER_NAME=$(SECURITY_USER_NAME) \
				SECURITY_USER_PASSWORD=$(SECURITY_USER_PASSWORD) \
				ARM_SUBSCRIPTION_ID='$(ARM_SUBSCRIPTION_ID)' \
				ARM_TENANT_ID=$(ARM_TENANT_ID) \
				ARM_CLIENT_ID=$(ARM_CLIENT_ID) \
				ARM_CLIENT_SECRET=$(ARM_CLIENT_SECRET) \
 				PAK_BUILD_CACHE_PATH=$(PAK_CACHE) \
 				GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)'

PAK_PATH=$(PWD)
RUN_CSB=$(BROKER_GO_OPTS) go run github.com/cloudfoundry/cloud-service-broker

LDFLAGS="-X github.com/cloudfoundry/cloud-service-broker/utils.Version=$(CSB_VERSION)"
GET_CSB="env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) github.com/cloudfoundry/cloud-service-broker"
else ifeq ($(DOCKER_OK), 0)
DOCKER_PAK_CACHE :=/pak-cache
BROKER_DOCKER_OPTS=--rm -v $(PAK_CACHE):$(PAK_CACHE) -v $(PWD):/brokerpak -w /brokerpak --network=host \
	-p 8080:8080 \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e ARM_SUBSCRIPTION_ID \
	-e ARM_TENANT_ID \
	-e ARM_CLIENT_ID \
	-e ARM_CLIENT_SECRET \
	-e "DB_TYPE=sqlite3" \
	-e "DB_PATH=/tmp/csb-db" \
	-e "PAK_BUILD_CACHE_PATH=$(PAK_CACHE)" \
	-e GSB_PROVISION_DEFAULTS

RUN_CSB=docker run $(BROKER_DOCKER_OPTS) $(CSB_DOCKER_IMAGE)

#### running go inside a container, this is for integration tests and push-broker
# path inside the container
PAK_PATH=/brokerpak

GO_DOCKER_OPTS=--rm -v $(PAK_CACHE):$(PAK_CACHE) -v $(PWD):/brokerpak -w /brokerpak --network=host
GO=docker run $(GO_DOCKER_OPTS) golang:latest go

GET_CSB="wget -O cloud-service-broker https://github.com/cloudfoundry/cloud-service-broker/releases/download/v$(CSB_RELEASE_VERSION)/cloud-service-broker.linux && chmod +x cloud-service-broker"
else
$(error either Go or Docker must be installed)
endif

###### Targets ################################################################

.PHONY: build
build: $(IAAS)-services-*.brokerpak

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*.tf ./tools/psqlcmd/build/psqlcmd_*.zip ./tools/sqlfailover/build/sqlfailover_*.zip | $(PAK_CACHE)
	$(RUN_CSB) pak build

.PHONY: run
run: build arm-subscription-id arm-tenant-id arm-client-id arm-client-secret ## start CSB in a docker container
	$(RUN_CSB) serve

.PHONY: catalog
catalog: build
		$(RUN_CSB) client catalog

.PHONY: docs
docs: build brokerpak-user-docs.md ## build docs

brokerpak-user-docs.md: *.yml
	$(RUN_CSB) pak docs $(PAK_PATH)/$(shell ls *.brokerpak) > $@ # GO

.PHONY: examples
examples: build ## display available examples
	 $(RUN_CSB) client examples

.PHONY: run-examples
run-examples: build ## run examples against CSB on localhost (run "make run" to start it), set service_name and example_name to run specific example
	$(RUN_CSB) client run-examples --service-name="$(service_name)" --example-name="$(example_name)" -j $(PARALLEL_JOB_COUNT)

.PHONY: info
info: build
	$(RUN_CSB) pak info $(PAK_PATH)/$(shell ls *.brokerpak)

.PHONY: validate
validate: build
	$(RUN_CSB) pak validate $(PAK_PATH)/$(shell ls *.brokerpak)

# fetching bits for cf push broker
cloud-service-broker: go.mod ## build or fetch CSB binary
	$(shell "$(GET_CSB)")

APP_NAME := $(or $(APP_NAME), cloud-service-broker)
DB_TLS := $(or $(DB_TLS), skip-verify)

.PHONY: push-broker
push-broker: cloud-service-broker build arm-subscription-id arm-tenant-id arm-client-id arm-client-secret ## push the broker to targetted Cloud Foundry
	MANIFEST=cf-manifest.yml APP_NAME=$(APP_NAME) DB_TLS=$(DB_TLS) GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)' ./scripts/push-broker.sh

.PHONY: clean
clean: ## clean up build artifacts
	- rm -f $(IAAS)-services-*.brokerpak
	- rm -f ./cloud-service-broker
	- rm -f ./brokerpak-user-docs.md
	- cd tools/psqlcmd; $(MAKE) clean
	- cd tools/sqlfailover; $(MAKE) clean
	- rm -rf $(PAK_CACHE)

.PHONY: rebuild
rebuild: clean build

./tools/psqlcmd/build/psqlcmd_*.zip: tools/psqlcmd/*.go
	cd tools/psqlcmd; $(MAKE) build

./tools/sqlfailover/build/sqlfailover_*.zip: tools/sqlfailover/*.go
	cd tools/sqlfailover; $(MAKE) build

.PHONY: arm-subscription-id
arm-subscription-id:
ifndef ARM_SUBSCRIPTION_ID
	$(error variable ARM_SUBSCRIPTION_ID not defined)
endif

.PHONY: arm-tenant-id
arm-tenant-id:
ifndef ARM_TENANT_ID
	$(error variable ARM_TENANT_ID not defined)
endif

.PHONY: arm-client-id
arm-client-id:
ifndef ARM_CLIENT_ID
	$(error variable ARM_CLIENT_ID not defined)
endif

.PHONY: arm-client-secret
arm-client-secret:
ifndef ARM_CLIENT_SECRET
	$(error variable ARM_CLIENT_SECRET not defined)
endif

$(PAK_CACHE):
	@echo "Folder $(PAK_CACHE) does not exist. Creating it..."
	mkdir -p $@

.PHONY: latest-csb
latest-csb: ## point to the very latest CSB on GitHub
	$(GO) get -d github.com/cloudfoundry/cloud-service-broker@main
	$(GO) mod tidy

.PHONY: local-csb
local-csb: ## point to a local CSB repo
	echo "replace \"github.com/cloudfoundry/cloud-service-broker\" => \"$$PWD/../cloud-service-broker\"" >>go.mod
	$(GO) mod tidy