.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: cloudfoundry.org ## build the provider

cloudfoundry.org: *.go */*.go
	mkdir -p cloudfoundry.org/cloud-service-broker/csbmssqldbrunfailover/1.0.0/linux_amd64
	mkdir -p cloudfoundry.org/cloud-service-broker/csbmssqldbrunfailover/1.0.0/darwin_amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloudfoundry.org/cloud-service-broker/csbmssqldbrunfailover/1.0.0/linux_amd64/terraform-provider-csbmssqldbrunfailover_v1.0.0
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o cloudfoundry.org/cloud-service-broker/csbmssqldbrunfailover/1.0.0/darwin_amd64/terraform-provider-csbmssqldbrunfailover_v1.0.0
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o cloudfoundry.org/cloud-service-broker/csbmssqldbrunfailover/1.0.0/darwin_arm64/terraform-provider-csbmssqldbrunfailover_v1.0.0

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf cloudfoundry.org
	- rm -rf /tmp/tpmssqldbrunfailover-non-fake.txt
	- rm -rf /tmp/tpmssqldbrunfailover-pkgs.txt
	- rm -rf /tmp/tpmssqldbrunfailover-coverage.out

.PHONY: test
test: ## run the tests
	go run github.com/onsi/ginkgo/v2/ginkgo --label-filter="!acceptance" -r

.PHONY: run-acceptance-tests
run-acceptance-tests: ## run the tests that are related to infrastructure
	export TF_ACC=1; \
	go run github.com/onsi/ginkgo/v2/ginkgo --label-filter="acceptance" -r -v

.PHONY: run-acceptance-tests-coverage
run-acceptance-tests-coverage: ## infrastructure tests coverage score
	go list ./... | grep -v fake > /tmp/tpmssqldbrunfailover-non-fake.txt
	paste -sd "," /tmp/tpmssqldbrunfailover-non-fake.txt > /tmp/tpmssqldbrunfailover-pkgs.txt
	export TF_ACC=1; \
	go test -coverpkg=`cat /tmp/tpmssqldbrunfailover-pkgs.txt` -coverprofile=/tmp/tpmssqldbrunfailover-coverage.out ./...
	go tool cover -func /tmp/tpmssqldbrunfailover-coverage.out | grep total
