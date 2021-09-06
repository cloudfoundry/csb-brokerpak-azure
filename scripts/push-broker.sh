#!/usr/bin/env bash

set +x # Hide secrets
set -o errexit
set -o pipefail

if [[ -z ${MANIFEST} ]]; then
  MANIFEST=manifest.yml
fi

if [[ -z ${APP_NAME} ]]; then
  APP_NAME=cloud-service-broker
fi

cf push --no-start -f "${MANIFEST}" --var app=${APP_NAME}

if [[ -z ${SECURITY_USER_NAME} ]]; then
  echo "Missing SECURITY_USER_NAME variable"
  exit 1
fi

if [[ -z ${SECURITY_USER_PASSWORD} ]]; then
  echo "Missing SECURITY_USER_PASSWORD variable"
  exit 1
fi

cf set-env "${APP_NAME}" SECURITY_USER_PASSWORD "${SECURITY_USER_PASSWORD}"
cf set-env "${APP_NAME}" SECURITY_USER_NAME "${SECURITY_USER_NAME}"

if [[ ${GSB_PROVISION_DEFAULTS} ]]; then
  cf set-env "${APP_NAME}" GSB_PROVISION_DEFAULTS "${GSB_PROVISION_DEFAULTS}"
fi

if [[ ${GOOGLE_CREDENTIALS} ]]; then
  cf set-env "${APP_NAME}" GOOGLE_CREDENTIALS "${GOOGLE_CREDENTIALS}"
fi

if [[ ${GOOGLE_PROJECT} ]]; then
  cf set-env "${APP_NAME}" GOOGLE_PROJECT "${GOOGLE_PROJECT}"
fi

if [[ ${ARM_SUBSCRIPTION_ID} ]]; then
  cf set-env "${APP_NAME}" ARM_SUBSCRIPTION_ID "${ARM_SUBSCRIPTION_ID}"
fi

if [[ ${ARM_TENANT_ID} ]]; then
  cf set-env "${APP_NAME}" ARM_TENANT_ID "${ARM_TENANT_ID}"
fi

if [[ ${ARM_CLIENT_ID} ]]; then
  cf set-env "${APP_NAME}" ARM_CLIENT_ID "${ARM_CLIENT_ID}"
fi

if [[ ${ARM_CLIENT_SECRET} ]]; then
  cf set-env "${APP_NAME}" ARM_CLIENT_SECRET "${ARM_CLIENT_SECRET}"
fi

if [[ ${GSB_BROKERPAK_BUILTIN_PATH} ]]; then
  cf set-env "${APP_NAME}" GSB_BROKERPAK_BUILTIN_PATH "${GSB_BROKERPAK_BUILTIN_PATH}"
fi

if [[ ${DB_TLS} ]]; then
  cf set-env "${APP_NAME}" DB_TLS "${DB_TLS}"
fi

if [[ ${CH_CRED_HUB_URL} ]]; then
  cf set-env "${APP_NAME}" CH_CRED_HUB_URL "${CH_CRED_HUB_URL}"
fi

if [[ ${CH_UAA_URL} ]]; then
  cf set-env "${APP_NAME}" CH_UAA_URL "${CH_UAA_URL}"
fi

if [[ ${CH_UAA_CLIENT_NAME} ]]; then
  cf set-env "${APP_NAME}" CH_UAA_CLIENT_NAME "${CH_UAA_CLIENT_NAME}"
fi

if [[ ${CH_UAA_CLIENT_SECRET} ]]; then
  cf set-env "${APP_NAME}" CH_UAA_CLIENT_SECRET "${CH_UAA_CLIENT_SECRET}"
fi

if [[ ${CH_SKIP_SSL_VALIDATION} ]]; then
  cf set-env "${APP_NAME}" CH_SKIP_SSL_VALIDATION "${CH_SKIP_SSL_VALIDATION}"
fi

if [[ ${EXPERIMENTAL_ENCRYPTION_ENABLED} ]]; then
  cf set-env "${APP_NAME}" EXPERIMENTAL_ENCRYPTION_ENABLED "${EXPERIMENTAL_ENCRYPTION_ENABLED}"
fi

if [[ ${EXPERIMENTAL_ENCRYPTION_PASSWORDS} ]]; then
  cf set-env "${APP_NAME}" EXPERIMENTAL_ENCRYPTION_PASSWORDS "${EXPERIMENTAL_ENCRYPTION_PASSWORDS}"
fi

if [[ -z ${MSYQL_INSTANCE} ]]; then
  MSYQL_INSTANCE=csb-sql
fi

cf bind-service "${APP_NAME}" "${MSYQL_INSTANCE}"

cf start "${APP_NAME}"

if [[ -z ${BROKER_NAME} ]]; then
  BROKER_NAME=csb-$USER
fi

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
