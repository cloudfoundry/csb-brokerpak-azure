###### Help ###################################################################
.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Setup ##################################################################
IAAS=azure
CSB_VERSION := $(or $(CSB_VERSION), $(shell grep 'github.com/cloudfoundry/cloud-service-broker' go.mod | grep -v replace | awk '{print $$NF}' | sed -e 's/v//'))
CSB_RELEASE_VERSION := $(CSB_VERSION)

####### broker environment variables
SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), $(IAAS)-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), $(IAAS)-broker-pw)
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"resource_group":"broker-cf-test","location":"westus2"})

BROKER_GO_OPTS=PORT=8080 \
				DB_TYPE=sqlite3 \
				DB_PATH=/tmp/csb-db \
				SECURITY_USER_NAME=$(SECURITY_USER_NAME) \
				SECURITY_USER_PASSWORD=$(SECURITY_USER_PASSWORD) \
				ARM_SUBSCRIPTION_ID='$(ARM_SUBSCRIPTION_ID)' \
				ARM_TENANT_ID=$(ARM_TENANT_ID) \
				ARM_CLIENT_ID=$(ARM_CLIENT_ID) \
				ARM_CLIENT_SECRET=$(ARM_CLIENT_SECRET) \
 				PAK_BUILD_CACHE_PATH=$(PAK_BUILD_CACHE_PATH) \
 				GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)'

PAK_PATH=$(PWD)
RUN_CSB=$(BROKER_GO_OPTS) go run github.com/cloudfoundry/cloud-service-broker/v2

LDFLAGS="-X github.com/cloudfoundry/cloud-service-broker/v2/utils.Version=$(CSB_VERSION)"
GET_CSB="env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) github.com/cloudfoundry/cloud-service-broker/v2"

###### Targets ################################################################

.PHONY: build
build: $(IAAS)-services-*.brokerpak ## build brokerpak

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*.tf ./providers/terraform-provider-csbmssqldbrunfailover/cloudfoundry.org/cloud-service-broker/csbmssqldbrunfailover | $(PAK_BUILD_CACHE_PATH)
	$(RUN_CSB) pak build

.PHONY: run
run: arm-subscription-id arm-tenant-id arm-client-id arm-client-secret ## start broker with this brokerpak
	$(RUN_CSB) pak build --target current
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
	 $(RUN_CSB) examples

.PHONY: run-examples
run-examples: build ## run examples tests, set service_name and/or example_name
	$(RUN_CSB) run-examples --service-name="$(service_name)" --example-name="$(example_name)"

###### test ###################################################################

.PHONY: test
test: lint run-integration-tests ## run the tests

.PHONY: run-integration-tests
run-integration-tests: provider-tests ## run integration tests for this brokerpak
	cd ./integration-tests && go run github.com/onsi/ginkgo/v2/ginkgo -r .

.PHONY: run-terraform-tests
run-terraform-tests: ## run terraform tests for this brokerpak
	cd ./terraform-tests && go run github.com/onsi/ginkgo/v2/ginkgo -r .

.PHONY: provider-tests
provider-tests:  ## run the integration tests associated with providers
	cd providers/terraform-provider-csbmssqldbrunfailover; $(MAKE) test

.PHONY: provider-acceptance-tests
provider-acceptance-tests: ## run the tests that are related to infrastructure
	cd providers/terraform-provider-csbmssqldbrunfailover; $(MAKE) run-acceptance-tests

.PHONY: provider-csbmssqldbrunfailover-coverage
provider-csbmssqldbrunfailover-coverage: ## csbmssqldbrunfailover tests coverage score
	cd providers/terraform-provider-csbmssqldbrunfailover; $(MAKE) run-acceptance-tests-coverage

.PHONY: info
info: build ## show brokerpak info
	$(RUN_CSB) pak info $(PAK_PATH)/$(shell ls *.brokerpak)

.PHONY: validate
validate: build  ## validate pak syntax
	$(RUN_CSB) pak validate $(PAK_PATH)/$(shell ls *.brokerpak)

# fetching bits for cf push broker
cloud-service-broker: go.mod ## build or fetch CSB binary
	"$(GET_CSB)"

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
	- cd providers/terraform-provider-csbmssqldbrunfailover; $(MAKE) clean

.PHONY: rebuild
rebuild: clean build

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

$(PAK_BUILD_CACHE_PATH):
	@echo "Folder $(PAK_BUILD_CACHE_PATH) does not exist. Creating it..."
	mkdir -p $@

.PHONY: latest-csb
latest-csb: ## point to the very latest CSB on GitHub
	go get -d github.com/cloudfoundry/cloud-service-broker@main
	go mod tidy

.PHONY: local-csb
local-csb: ## point to a local CSB repo
	echo "replace \"github.com/cloudfoundry/cloud-service-broker\" => \"$$PWD/../cloud-service-broker\"" >>go.mod
	go mod tidy

###### lint ###################################################################

.PHONY: lint
lint: checkgoformat checkgoimports checktfformat vet staticcheck ## checks format, imports and vet

checktfformat: ## checks that Terraform HCL is formatted correctly
	@@if [ "$$(terraform fmt -recursive --check)" ]; then \
		echo "terraform fmt check failed: run 'make format'"; \
		exit 1; \
	fi

checkgoformat: ## checks that the Go code is formatted correctly
	@@if [ -n "$$(gofmt -s -e -l -d .)" ]; then       \
		echo "gofmt check failed: run 'make format'"; \
		exit 1;                                       \
	fi

checkgoimports: ## checks that Go imports are formatted correctly
	@@if [ -n "$$(go run golang.org/x/tools/cmd/goimports -l -d -local csbbrokerpakazure .)" ]; then \
		echo "goimports check failed: run 'make format'";                      \
		exit 1;                                                                \
	fi

vet: ## runs go vet
	go vet ./...

staticcheck: ## runs staticcheck
	go run honnef.co/go/tools/cmd/staticcheck ./...

.PHONY: format
format: ## format the source
	gofmt -s -e -l -w .
	go run golang.org/x/tools/cmd/goimports -l -w -local csbbrokerpakazure .
	terraform fmt --recursive

./providers/terraform-provider-csbmssqldbrunfailover/cloudfoundry.org/cloud-service-broker/csbmssqldbrunfailover:
	cd providers/terraform-provider-csbmssqldbrunfailover; $(MAKE) build
