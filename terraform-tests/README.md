# Terraform Test Tools and Scripts

Terraform tests are `unit` tests that run with the system's installed terraform binary on the terraform files directly. They will download the latest versions of the providers
that comply with the restrictions in the definition file. They won't look at the service offering definitions (.yml) or the manifest.yml files. 

The aim of these tests is to verify that Terraform gets a set of variables inputs, and it will send the expected values to the providers. 
This is accomplished by running `terraform plan --no-refresh` and analysing the output with terraform provided commands and structures. 

Properties that are not given a value will be `Unknown` until an apply operation is performed and therefore no much checking can be performed.
Also, given that no `apply` is done, this won't verify the instances are created with the expected values, and downstream services (providers, IaaS API),
can have defaults or logic implemented that end up changing the end state.

`data` resources needed by terraform to run a given module must be present in the IaaS. 

## Running the tests
### Pre-requisite software
- The [Go Programming language](https://golang.org/)
- [Terraform](https://www.terraform.io/downloads)

### Environment
- `ARM_CLIENT_ID`, `"ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, and `ARM_TENANT_ID` must be set as environment variables as Terraform will attempt to connect to the IaaS.

