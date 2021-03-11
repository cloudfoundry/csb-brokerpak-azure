#!/usr/bin/env bash

set +x # Hide secrets

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && echo -e "You must source this script\nsource ${0}" && exit 1

export ARM_SUBSCRIPTION_ID=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/Azure Service Account Key" | jq -r .subscription_id)
export ARM_TENANT_ID=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/Azure Service Account Key" | jq -r .tenant_id)
export ARM_CLIENT_ID=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/Azure Service Account Key" | jq -r .client_id)
export ARM_CLIENT_SECRET=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/Azure Service Account Key" | jq -r .client_secret)

export AZURE_AUTHORIZED_NETWORK_ID=$(lpass show "Shared-CF Platform Engineering/pe-cloud-service-broker/cloud service broker pipeline secrets.yml" | grep azure-subnet-id | cut -d ' ' -f 2)
export AZURE_LOCATION=$(lpass show "Shared-CF Platform Engineering/pe-cloud-service-broker/cloud service broker pipeline secrets.yml" | grep azure-location | cut -d ' ' -f 2)
export GSB_PROVISION_DEFAULTS="{\"resource_group\": \"broker-cf-test\", \"authorized_network\":\"${AZURE_AUTHORIZED_NETWORK_ID}\", \"location\":\"${AZURE_LOCATION}\"}"

export SECURITY_USER_NAME=brokeruser
export SECURITY_USER_PASSWORD=brokeruserpassword
export DB_HOST=localhost
export DB_USERNAME=broker
export DB_PASSWORD=brokerpass
export DB_NAME=brokerdb
export PORT=8080