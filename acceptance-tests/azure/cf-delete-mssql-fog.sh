#!/usr/bin/env bash

set -e
set -o nounset

NAME=$1; shift

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

delete_service "${NAME}"

cf unset-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS
cf unset-env cloud-service-broker MSSQL_DB_FOG_SERVER_PAIR_CREDS
cf restart cloud-service-broker     

PRIMARY_SERVER_NAME="mssql-${NAME}-p"
SECONDARY_SERVER_NAME="mssql-${NAME}-s"

if [ $# -gt 0 ]; then
  PRIMARY_SERVER_NAME=$1; shift
fi

if [ $# -gt 0 ]; then
  SECONDARY_SERVER_NAME=$1; shift
fi

${SCRIPT_DIR}/cf-delete-mssql-server.sh "${PRIMARY_SERVER_NAME}" &
${SCRIPT_DIR}/cf-delete-mssql-server.sh "${SECONDARY_SERVER_NAME}" &

wait

exit $?