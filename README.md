# Azure Brokerpak

A brokerpak for the [Cloud Service Broker](https://github.com/pivotal/cloud-service-broker) that provides support for Azure services.

## Development Requirements

* Either Go 1.18 or [Docker](https://docs.docker.com/get-docker/)
* make - covers development lifecycle steps

A docker container for the cloud service broker binary is available at *cfplatformeng/csb*

## Azure account information

To provision services, the brokerpak currently requires Azure account values. The brokerpak expects them in environment variables:

* ARM_SUBSCRIPTION_ID
* ARM_TENANT_ID
* ARM_CLIENT_ID
* ARM_CLIENT_SECRET

## Development Tools

A Makefile supports the full local development lifecycle for the brokerpak.

The make targets can be run either with Docker or installing the required libraries in the local OS.

Available make targets can be listed by running `make`.

### Running with docker

1. Install [Docker](https://docs.docker.com/get-docker/)
2. If you don't have Go installed, the makefile will automatically use Docker. If you do have go installed but still want to use docker, then set the `USE_GO_CONTAINERS` to `true`.

Make targets will run with the *cfplatformeng/csb* docker image. Alternatively, a custom image can be specified by setting the `CSB` environment variable.

### Running with Go

1. Make sure you have the right Go version installed (see `go.mod` file).
2. Make sure `USE_GO_CONTAINERS` environment variable is ***NOT*** set.

The make targets will build the source using the local go installation.

### Other targets

There is a make target to push the broker and brokerpak into a CloudFoundry foundation. It will be necessary to manually configure a few items for the broker to work.

- `make push-broker` will `cf push` the broker into CloudFoundry. Requires the `cf` cli to be installed.

The broker gets pushed into CloudFoundry as *cloud-service-broker-azure*  It will be necessary to bind a MySQL database to the broker to provide broker state storage. See [Azure Installation](./docs/azure-installation.md) docs for more info.

## Broker
The version of Cloud Service Broker to use with this brokerpak is encoded in the `go.mod` file.
The make targets will use this version by default.

## Tests

### Example tests

Services definitions declare examples for each plan they provide. Those examples are then run through the whole cycle of `provision`, `bind`, `unbind`, and `delete` when running

```
terminal 1
>> make run

terminal 2
>> make run-examples
```

## Acceptance tests

See [acceptance tests](acceptance-tests/README.md)

## Integration tests

Integration tests can be run with the following command:

```bash
make run-integration-tests
```

