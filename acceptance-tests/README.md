# Acceptance Test Tools and Scripts

Acceptance tests are run as `cf push`'ed applications that verify connectivity in a real Cloud Foundry environment.

The main pattern is:
1. `cf create-service` the service instance to test
1. `cf bind-service` a writer test app to the provisioned service
1. `cf bind-service` a reader test app to the provisioned service
1. Check the binding contains CredHub references and not credentials
1. Use the writer test app to write some data
1. Use the reader test app to read back the data and check it is the same
1. `cf unbind-service` for both apps
1. `cf delete-service` the tested instance 

## Running the tests
### Pre-requisite software
- The [Go Programming language](https://golang.org/)
- The [Cloud Foundry CLI](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html)
- The [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)

### Environment variables
- ARM_SUBSCRIPTION_ID
- ARM_CLIENT_SECRET
- ARM_TENANT_ID
- ARM_CLIENT_ID

### Environment
- A Cloud Foundry instance logged in and targeted
- The Cloud Service Broker and this brokerpak deployed by running `make push-broker` or equivalent