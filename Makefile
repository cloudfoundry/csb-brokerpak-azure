###### Help ###################################################################

.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Targets ################################################################

IAAS=azure
DOCKER_OPTS=--rm -v $(PWD):/brokerpak -w /brokerpak --network=host
CSB := $(or $(CSB), cfplatformeng/csb)

.PHONY: build
build: $(IAAS)-services-*.brokerpak 

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*.tf ./tools/psqlcmd/build/psqlcmd_*.zip ./tools/sqlfailover/build/sqlfailover_*.zip
	docker run $(DOCKER_OPTS) $(CSB) pak build

SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), $(IAAS)-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), $(IAAS)-broker-pw)
PARALLEL_JOB_COUNT := $(or $(PARALLEL_JOB_COUNT), 2)

.PHONY: run
run: build arm-subscription-id arm-tenant-id arm-client-id arm-client-secret ## start CSB in a docker container
	docker run $(DOCKER_OPTS) \
	-p 8080:8080 \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e ARM_SUBSCRIPTION_ID \
	-e ARM_TENANT_ID \
	-e ARM_CLIENT_ID \
	-e ARM_CLIENT_SECRET \
	-e "DB_TYPE=sqlite3" \
	-e "DB_PATH=/tmp/csb-db" \
	-e GSB_PROVISION_DEFAULTS \
	$(CSB) serve

.PHONY: catalog
catalog: build
	docker run $(DOCKER_OPTS) \	
		-e SECURITY_USER_NAME \
		-e SECURITY_USER_PASSWORD \
		-e USER \
		$(CSB) client catalog

.PHONY: docs
docs: build brokerpak-user-docs.md

brokerpak-user-docs.md: *.yml
	docker run $(DOCKER_OPTS) \
	$(CSB) pak docs /brokerpak/$(shell ls *.brokerpak) > $@

.PHONY: examples
examples: build ## display available examples
	docker run $(DOCKER_OPTS) \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e USER \
	$(CSB) client examples

.PHONY: run-examples
run-examples: build ## run examples against CSB on localhost (run "make run" to start it), set service_name and example_name to run specific example
	docker run $(DOCKER_OPTS) \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e USER \
	$(CSB) client run-examples --service-name="$(service_name)" --example-name="$(example_name)" -j $(PARALLEL_JOB_COUNT)

.PHONY: info
info: build
	docker run $(DOCKER_OPTS) \
	$(CSB) pak info /brokerpak/$(shell ls *.brokerpak)

.PHONY: validate
validate: build
	docker run $(DOCKER_OPTS) \
	$(CSB) pak validate /brokerpak/$(shell ls *.brokerpak)

# fetching bits for cf push broker
cloud-service-broker:
	wget $(shell curl -sL https://api.github.com/repos/cloudfoundry-incubator/cloud-service-broker/releases/latest | jq -r '.assets[] | select(.name == "cloud-service-broker.linux") | .browser_download_url')
	mv ./cloud-service-broker.linux ./cloud-service-broker
	chmod +x ./cloud-service-broker

APP_NAME := $(or $(APP_NAME), cloud-service-broker)
DB_TLS := $(or $(DB_TLS), skip-verify)
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"resource_group": "broker-cf-test"})

.PHONY: push-broker
push-broker: cloud-service-broker build arm-subscription-id arm-tenant-id arm-client-id arm-client-secret
	MANIFEST=cf-manifest.yml APP_NAME=$(APP_NAME) DB_TLS=$(DB_TLS) GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)' ./scripts/push-broker.sh

.PHONY: clean
clean:
	- rm $(IAAS)-services-*.brokerpak
	- rm ./cloud-service-broker
	- rm ./brokerpak-user-docs.md
	- cd tools/psqlcmd; $(MAKE) clean
	- cd tools/sqlfailover; $(MAKE) clean

.PHONY: rebuild
rebuild: clean build

./tools/psqlcmd/build/psqlcmd_*.zip: tools/psqlcmd/*.go
	cd tools/psqlcmd; USE_GO_CONTAINERS=1 $(MAKE) build

./tools/sqlfailover/build/sqlfailover_*.zip: tools/sqlfailover/*.go
	cd tools/sqlfailover; USE_GO_CONTAINERS=1 $(MAKE) build

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
