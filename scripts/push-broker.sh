#!/usr/bin/env bash

set +x # Hide secrets
set -o errexit
set -o pipefail
set -e

if [[ -z ${MANIFEST} ]]; then
  MANIFEST=manifest.yml
fi

if [[ -z ${APP_NAME} ]]; then
  APP_NAME=cloud-service-broker
fi

if [[ -z ${SECURITY_USER_NAME} ]]; then
  echo "Missing SECURITY_USER_NAME variable"
  exit 1
fi

if [[ -z ${SECURITY_USER_PASSWORD} ]]; then
  echo "Missing SECURITY_USER_PASSWORD variable"
  exit 1
fi

cfmf="/tmp/cf-manifest.$$.yml"
touch "$cfmf"
trap "rm -f $cfmf" EXIT
chmod 600 "$cfmf"
cat "$MANIFEST" >$cfmf

echo "  env:" >>$cfmf
echo "    SECURITY_USER_PASSWORD: ${SECURITY_USER_PASSWORD}" >>$cfmf
echo "    SECURITY_USER_NAME: ${SECURITY_USER_NAME}" >>$cfmf
echo "    TERRAFORM_UPGRADES_ENABLED: ${TERRAFORM_UPGRADES_ENABLED:-true}" >>$cfmf
echo "    BROKERPAK_UPDATES_ENABLED: ${BROKERPAK_UPDATES_ENABLED:-true}" >>$cfmf
echo "    CSB_DISABLE_TF_UPGRADE_PROVIDER_RENAMES: ${CSB_DISABLE_TF_UPGRADE_PROVIDER_RENAMES:-true}" >>$cfmf

if [[ ${GSB_PROVISION_DEFAULTS} ]]; then
  echo "    GSB_PROVISION_DEFAULTS: $(echo "$GSB_PROVISION_DEFAULTS" | jq @json)" >>$cfmf
fi

if [[ ${ARM_SUBSCRIPTION_ID} ]]; then
  echo "    ARM_SUBSCRIPTION_ID: ${ARM_SUBSCRIPTION_ID}" >>$cfmf
fi

if [[ ${ARM_TENANT_ID} ]]; then
  echo "    ARM_TENANT_ID: ${ARM_TENANT_ID}" >>$cfmf
fi

if [[ ${ARM_CLIENT_ID} ]]; then
  echo "    ARM_CLIENT_ID: ${ARM_CLIENT_ID}" >>$cfmf
fi

if [[ ${ARM_CLIENT_SECRET} ]]; then
  echo "    ARM_CLIENT_SECRET: ${ARM_CLIENT_SECRET}" >>$cfmf
fi

if [[ ${GSB_BROKERPAK_BUILTIN_PATH} ]]; then
  echo "    GSB_BROKERPAK_BUILTIN_PATH: ${GSB_BROKERPAK_BUILTIN_PATH}" >>$cfmf
fi

if [[ ${DB_TLS} ]]; then
  echo "    DB_TLS: ${DB_TLS}" >>$cfmf
fi

if [[ ${CH_CRED_HUB_URL} ]]; then
  echo "    CH_CRED_HUB_URL: ${CH_CRED_HUB_URL}" >>$cfmf
fi

if [[ ${CH_UAA_URL} ]]; then
  echo "    CH_UAA_URL: ${CH_UAA_URL}" >>$cfmf
fi

if [[ ${CH_UAA_CLIENT_NAME} ]]; then
  echo "    CH_UAA_CLIENT_NAME: ${CH_UAA_CLIENT_NAME}" >>$cfmf
fi

if [[ ${CH_UAA_CLIENT_SECRET} ]]; then
  echo "    CH_UAA_CLIENT_SECRET: ${CH_UAA_CLIENT_SECRET}" >>$cfmf
fi

if [[ ${CH_SKIP_SSL_VALIDATION} ]]; then
  echo "    CH_SKIP_SSL_VALIDATION: ${CH_SKIP_SSL_VALIDATION}" >>$cfmf
fi

if [[ ${ENCRYPTION_ENABLED} ]]; then
  echo "    ENCRYPTION_ENABLED: ${ENCRYPTION_ENABLED}" >>$cfmf
fi

if [[ ${ENCRYPTION_PASSWORDS} ]]; then
  echo "    ENCRYPTION_PASSWORDS: $(echo "$ENCRYPTION_PASSWORDS" | jq @json)" >>$cfmf
fi

if [[ ${GSB_COMPATIBILITY_ENABLE_GCP_DEPRECATED_SERVICES} ]]; then
  echo "    GSB_COMPATIBILITY_ENABLE_GCP_DEPRECATED_SERVICES: $(echo "$GSB_COMPATIBILITY_ENABLE_GCP_DEPRECATED_SERVICES" | jq @json)" >>$cfmf
fi

cf push --no-start -f "${cfmf}" --var app=${APP_NAME}

if [[ -z ${MSYQL_INSTANCE} ]]; then
  MSYQL_INSTANCE=csb-sql
fi

cf bind-service "${APP_NAME}" "${MSYQL_INSTANCE}"

cf start "${APP_NAME}"

if [[ -z ${BROKER_NAME} ]]; then
  BROKER_NAME=csb-$USER
fi

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(LANG=EN cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped --update-if-exists
