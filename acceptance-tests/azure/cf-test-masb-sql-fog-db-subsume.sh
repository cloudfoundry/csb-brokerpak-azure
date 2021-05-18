#!/usr/bin/env bash

set -e
set -o nounset
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

. "${SCRIPT_DIR}/../functions.sh"

if [ $# -lt 5 ]; then
    echo "usage: $0 <resource group> <primary server name> <secondary server name> <admin username> <admin password>"
    exit 1
fi

print_log_message "Starting test masb-sql-fog-db"

SERVER_RESOURCE_GROUP=$1
shift
PRIMARY_SERVER_NAME=$1
shift
SECONDARY_SERVER_NAME=$1
shift
SERVER_ADMIN_USER_NAME=$1
shift
SERVER_ADMIN_PASSWORD=$1
shift

MASB_ID=$(date +%s)

DB_NAME=subsume-db-${MASB_ID}

RESULT=1

MASB_SQLDB_INSTANCE_NAME=mssql-db-${MASB_ID}
MASB_DB_CONFIG="{ \
  \"sqlServerName\": \"${PRIMARY_SERVER_NAME}\", \
  \"sqldbName\": \"${DB_NAME}\", \
  \"resourceGroup\": \"${SERVER_RESOURCE_GROUP}\" \
}"

RESULT=1
print_log_message "Given there is a masb db server"
if create_service azure-sqldb StandardS0 "${MASB_SQLDB_INSTANCE_NAME}" "${MASB_DB_CONFIG}"; then
    MASB_FOG_INSTANCE_NAME=masb-fog-db-${MASB_ID}

    MASB_FOG_CONFIG="{ \
      \"primaryServerName\": \"${PRIMARY_SERVER_NAME}\", \
      \"primaryDbName\": \"${DB_NAME}\", \
      \"secondaryServerName\": \"${SECONDARY_SERVER_NAME}\", \
      \"failoverGroupName\": \"${MASB_FOG_INSTANCE_NAME}\", \
      \"readWriteEndpoint\": { \
        \"failoverPolicy\": \"Automatic\", \
        \"failoverWithDataLossGracePeriodMinutes\": 60 \
      } \
    }"
    print_log_message "Given there is a masb failover group in that server"
    if create_service azure-sqldb-failover-group SecondaryDatabaseWithFailoverGroup "${MASB_FOG_INSTANCE_NAME}" "${MASB_FOG_CONFIG}"; then

        print_log_message "Given the fog db can be bound and works"
        if bind_service_test spring-music "${MASB_FOG_INSTANCE_NAME}"; then

            SUBSUME_CONFIG="{ \
                \"azure_primary_db_id\": \"$(az sql failover-group show --name ${MASB_FOG_INSTANCE_NAME} --server ${PRIMARY_SERVER_NAME} --resource-group ${SERVER_RESOURCE_GROUP} --query databases[0] -o tsv)\", \
                \"azure_secondary_db_id\": \"$(az sql failover-group show --name ${MASB_FOG_INSTANCE_NAME} --server ${SECONDARY_SERVER_NAME} --resource-group ${SERVER_RESOURCE_GROUP} --query databases[0] -o tsv)\", \
                \"azure_fog_id\": \"$(az sql failover-group show --name ${MASB_FOG_INSTANCE_NAME} --server ${PRIMARY_SERVER_NAME} --resource-group ${SERVER_RESOURCE_GROUP} --query id -o tsv)\", \

                \"server_pair\": \"test_server\" \
            }"

            MSSQL_DB_FOG_SERVER_PAIR_CREDS="{ \
                \"test_server\": { \
                    \"admin_username\":\"${SERVER_ADMIN_USER_NAME}\", \
                    \"admin_password\":\"${SERVER_ADMIN_PASSWORD}\", \
                    \"primary\":{ \
                        \"server_name\":\"${PRIMARY_SERVER_NAME}\", \
                        \"resource_group\":\"${SERVER_RESOURCE_GROUP}\" \
                    }, \
                    \"secondary\":{ \
                        \"server_name\":\"${SECONDARY_SERVER_NAME}\", \
                        \"resource_group\":\"${SERVER_RESOURCE_GROUP}\" \
                    } \
                  } \
              }"

            #echo $SUBSUME_CONFIG

            GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS="{ \
                \"server_credential_pairs\":${MSSQL_DB_FOG_SERVER_PAIR_CREDS} \
            }"

            cf set-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS "${GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS}"
            cf set-env cloud-service-broker MSSQL_DB_FOG_SERVER_PAIR_CREDS "${MSSQL_DB_FOG_SERVER_PAIR_CREDS}"
            cf restart cloud-service-broker

            SUBSUMED_INSTANCE_NAME=masb-sql-db-subsume-test-$$
            print_log_message "When CSB subsumes the failover group"
            if create_service csb-azure-mssql-db-failover-group subsume "${SUBSUMED_INSTANCE_NAME}" "${SUBSUME_CONFIG}"; then

                print_log_message "The db can be bound and works"
                if "${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${SUBSUMED_INSTANCE_NAME}"; then
                  print_log_message "subsumed masb fog instance test successful"

                  print_log_message "AND Plan updates work fine"
                  if "${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${SUBSUMED_INSTANCE_NAME}" medium; then
                      print_log_message "AND cannot update to the subsume plan"
                      if update_service_plan "${SUBSUMED_INSTANCE_NAME}" subsume; then
                          cf service "${SUBSUMED_INSTANCE_NAME}"
                          print_log_message "failed: should not have been able to update to subsume plan"
                      else
                          print_log_message "success: plan update rejected as expected"
                          print_log_message "setting return code to 0..."
                          RESULT=0
                      fi
                  else
                      print_log_message "failed: updating plan for subsumed instance to medium failed spring music test "${SUBSUMED_INSTANCE_NAME}""
                  fi
                else
                    print_log_message "failed: spring music test on subsumed masb fog instance "${SUBSUMED_INSTANCE_NAME}" test failed"
                fi
                print_log_message "Teardown"
                delete_service "${SUBSUMED_INSTANCE_NAME}" || cf purge-service-instance -f "${SUBSUMED_INSTANCE_NAME}"
            fi
            print_log_message "Teardown"
            cf unset-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS
            cf unset-env cloud-service-broker MSSQL_DB_FOG_SERVER_PAIR_CREDS
            cf restart cloud-service-broker
        else
            print_log_message "failed: failed to bind service"
        fi
        print_log_message "Teardown"
        delete_service "${MASB_FOG_INSTANCE_NAME}" || cf purge-service-instance -f "${MASB_FOG_INSTANCE_NAME}"
        delete_service "${MASB_SQLDB_INSTANCE_NAME}" || cf purge-service-instance -f "${MASB_SQLDB_INSTANCE_NAME}"
    else
      print_log_message "Teardown"
      delete_service "${MASB_SQLDB_INSTANCE_NAME}"
    fi
fi
print_log_message "Finished test masb-sql-fog-db with code: ${RESULT}"
exit ${RESULT}
