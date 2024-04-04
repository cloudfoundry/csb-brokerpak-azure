# Test, Tools and Scripts

Tests within the `terraform-test` folder are `unit` tests that run with the system's installed OpenTofu binary on the
OpenTofu language files directly. These tests download the latest versions of the providers that comply with the
restrictions in the definition file. They do not examine the service offering definitions (.yml) or the manifest.yml
files.

The objective of these tests is to ensure that OpenTofu receives a set of variable inputs and sends the expected values
to the providers. This verification is achieved by executing `tofu plan -refresh=false` and analyzing the outputs.

Properties without assigned values will remain `Unknown` until an apply operation is executed, limiting the extent of
checks that can be performed. Moreover, since no `apply` operation is conducted, this process does not verify if
instances are created with the expected values. Additionally, downstream services (such as providers and IaaS APIs)
might have defaults or embedded logic that could alter the final state.

`data` resources needed by OpenTofu to run a given module must be present in the IaaS.

## Running the tests

### Pre-requisite software

- The [Go Programming language](https://golang.org/)
- [OpenTofu](https://opentofu.org)

### Environment

- `ARM_CLIENT_ID`, `"ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, and `ARM_TENANT_ID` must be set as environment variables
  as OpenTofu will attempt to connect to the IaaS.

